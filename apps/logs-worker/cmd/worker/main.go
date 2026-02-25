package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Event struct {
	EventID      string                 `json:"eventId"`
	CompanyID    string                 `json:"companyId"`
	TraceID      string                 `json:"traceId"`
	NodeID       string                 `json:"nodeId"`
	ParentNodeID *string                `json:"parentNodeId"`
	ServiceName  string                 `json:"serviceName"`
	Operation    map[string]any         `json:"operation"`
	HTTP         map[string]any         `json:"http"`
	DB           map[string]any         `json:"db"`
	Metadata     map[string]any         `json:"metadata"`
}

func main() {
	hosts := strings.Split(env("CASSANDRA_HOSTS", "localhost"), ",")
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "logs"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	conn, err := amqp.Dial(env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"))
	if err != nil {
		log.Fatal(err)
	}
	ch, _ := conn.Channel()
	msgs, _ := ch.Consume("log_events", "", false, false, false, false, nil)
	log.Println("logs-worker consuming")

	for msg := range msgs {
		var e Event
		if err := json.Unmarshal(msg.Body, &e); err != nil {
			_ = msg.Nack(false, false)
			continue
		}
		applied, err := session.Query("INSERT INTO event_dedup(company_id,event_id,received_at) VALUES(?,?,?) IF NOT EXISTS", e.CompanyID, e.EventID, time.Now()).ScanCAS()
		if err != nil || !applied {
			_ = msg.Ack(false)
			continue
		}

		meta, _ := json.Marshal(e.Metadata)
		start := parseTime(e.Operation["startAt"])
		end := parseTime(e.Operation["endAt"])
		duration := toInt(e.Operation["durationMs"])
		status := toStr(e.Operation["status"])
		typeOp := toStr(e.Operation["type"])
		name := toStr(e.Operation["name"])
		hMethod := toStr(e.HTTP["method"])
		hPath := toStr(e.HTTP["path"])
		hStatus := toInt(e.HTTP["statusCode"])
		dbSystem := toStr(e.DB["system"])
		dbQuery := toStr(e.DB["query"])
		dbRows := toInt(e.DB["rows"])

		_ = session.Query(`INSERT INTO nodes_by_trace (trace_id,node_id,parent_node_id,company_id,service_name,type,name,status,start_at,end_at,duration_ms,http_method,http_path,http_status,db_system,db_query,db_rows,metadata)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, e.TraceID, e.NodeID, e.ParentNodeID, e.CompanyID, e.ServiceName, typeOp, name, status, start, end, duration, hMethod, hPath, hStatus, dbSystem, dbQuery, dbRows, string(meta)).Exec()

		if e.ParentNodeID != nil && *e.ParentNodeID != "" {
			_ = session.Query("INSERT INTO edges_by_trace(trace_id,from_node_id,to_node_id,kind) VALUES(?,?,?,?)", e.TraceID, *e.ParentNodeID, e.NodeID, "PARENT_CHILD").Exec()
		}

		day := start.Format("2006-01-02")
		_ = session.Query(`INSERT INTO traces_by_company_day(company_id,day,started_at,trace_id,status,root_operation,service_name,http_method,http_path,http_status,duration_ms)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`, e.CompanyID, day, start, e.TraceID, status, name, e.ServiceName, hMethod, hPath, hStatus, duration).Exec()

		_ = msg.Ack(false)
	}
}

func env(k, d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }
func toStr(v any) string { if s,ok:=v.(string); ok { return s }; return "" }
func toInt(v any) int { switch t:=v.(type){case float64:return int(t);case int:return t;default:return 0} }
func parseTime(v any) time.Time { if s,ok:=v.(string); ok { t,err:=time.Parse(time.RFC3339, s); if err==nil { return t } }; return time.Now() }
