package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"logs-worker/internal/app"
	search "shared-search"

	"github.com/gocql/gocql"
)

type Repo struct {
	Session *gocql.Session
	Search  search.Client
}

func (r Repo) Persist(e app.Event) error {
	applied, err := r.Session.Query("INSERT INTO event_dedup(company_id,event_id,received_at) VALUES(?,?,?) IF NOT EXISTS", e.CompanyID, e.EventID, time.Now()).ScanCAS()
	if err != nil || !applied {
		return err
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
	if err := r.Session.Query(`INSERT INTO nodes_by_trace (trace_id,node_id,parent_node_id,company_id,service_name,type,name,status,start_at,end_at,duration_ms,http_method,http_path,http_status,db_system,db_query,db_rows,metadata)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, e.TraceID, e.NodeID, e.ParentNodeID, e.CompanyID, e.ServiceName, typeOp, name, status, start, end, duration, hMethod, hPath, hStatus, dbSystem, dbQuery, dbRows, string(meta)).Exec(); err != nil {
		return err
	}
	if e.ParentNodeID != nil && *e.ParentNodeID != "" {
		_ = r.Session.Query("INSERT INTO edges_by_trace(trace_id,from_node_id,to_node_id,kind) VALUES(?,?,?,?)", e.TraceID, *e.ParentNodeID, e.NodeID, "PARENT_CHILD").Exec()
	}

	// Index in OpenSearch
	if r.Search != nil {
		doc := map[string]interface{}{
			"traceId":     e.TraceID,
			"nodeId":      e.NodeID,
			"companyId":   e.CompanyID,
			"serviceName": e.ServiceName,
			"type":        typeOp,
			"name":        name,
			"dbSystem":    dbSystem,
			"dbQuery":     dbQuery,
			"httpPath":    hPath,
			"metadata":    string(meta),
		}
		docID := fmt.Sprintf("%s_%s", e.CompanyID, e.NodeID)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = r.Search.Index(ctx, "nodes", docID, doc)
		cancel()
	}

	day := start.Format("2006-01-02")
	
	// Increment endpoint counters
	if typeOp == "HTTP" && hMethod != "" && hPath != "" {
		_ = r.Session.Query(`UPDATE endpoint_calls_by_company_day SET calls = calls + 1 WHERE company_id=? AND day=? AND service_name=? AND http_method=? AND http_path=?`,
			e.CompanyID, day, e.ServiceName, hMethod, hPath).Exec()
	}

	return r.Session.Query(`INSERT INTO traces_by_company_day(company_id,day,started_at,trace_id,status,root_operation,service_name,http_method,http_path,http_status,duration_ms)
	VALUES (?,?,?,?,?,?,?,?,?,?,?)`, e.CompanyID, day, start, e.TraceID, status, name, e.ServiceName, hMethod, hPath, hStatus, duration).Exec()
}

func toStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func toInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	default:
		return 0
	}
}
func parseTime(v any) time.Time {
	if s, ok := v.(string); ok {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			return t
		}
	}
	return time.Now()
}
