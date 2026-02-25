package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gocql/gocql"
)

type TraceSummary struct {
	StartedAt    time.Time `json:"startedAt"`
	TraceID      string    `json:"traceId"`
	Status       string    `json:"status"`
	RootOperation string   `json:"rootOperation"`
	ServiceName  string    `json:"serviceName"`
	HTTPMethod   string    `json:"httpMethod"`
	HTTPPath     string    `json:"httpPath"`
	HTTPStatus   int       `json:"httpStatus"`
	DurationMS   int       `json:"durationMs"`
}

func main() {
	port := env("PORT", "8084")
	cluster := gocql.NewCluster(strings.Split(env("CASSANDRA_HOSTS", "localhost"), ",")...)
	cluster.Keyspace = "logs"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	r := chi.NewRouter()
	r.Get("/query/v1/traces", func(w http.ResponseWriter, r *http.Request) {
		companyID := r.URL.Query().Get("companyId")
		fromS := r.URL.Query().Get("from")
		toS := r.URL.Query().Get("to")
		from, _ := time.Parse(time.RFC3339, fromS)
		to, _ := time.Parse(time.RFC3339, toS)
		if companyID == "" || from.IsZero() || to.IsZero() {
			http.Error(w, "companyId, from, to required", 400)
			return
		}

		day := from.Format("2006-01-02")
		iter := session.Query("SELECT started_at,trace_id,status,root_operation,service_name,http_method,http_path,http_status,duration_ms FROM traces_by_company_day WHERE company_id=? AND day=?", companyID, day).Iter()
		var out []TraceSummary
		var s TraceSummary
		for iter.Scan(&s.StartedAt, &s.TraceID, &s.Status, &s.RootOperation, &s.ServiceName, &s.HTTPMethod, &s.HTTPPath, &s.HTTPStatus, &s.DurationMS) {
			if s.StartedAt.After(from) && s.StartedAt.Before(to) {
				out = append(out, s)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": out})
	})

	r.Get("/query/v1/traces/{traceId}", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceId")
		iter := session.Query("SELECT node_id,parent_node_id,service_name,type,name,status,start_at,end_at,duration_ms,http_method,http_path,http_status,db_system,db_query,db_rows,metadata FROM nodes_by_trace WHERE trace_id=?", traceID).Iter()
		var nodes []map[string]any
		var nodeID, parentID, serviceName, typ, name, status, hm, hp, dbs, dbq, meta string
		var start, end time.Time
		var duration, hs, dbr int
		for iter.Scan(&nodeID, &parentID, &serviceName, &typ, &name, &status, &start, &end, &duration, &hm, &hp, &hs, &dbs, &dbq, &dbr, &meta) {
			nodes = append(nodes, map[string]any{
				"id": nodeID,
				"data": map[string]any{"label": name, "type": typ, "status": status, "serviceName": serviceName, "durationMs": duration, "startAt": start, "endAt": end, "httpMethod": hm, "httpPath": hp, "httpStatus": hs, "dbSystem": dbs, "dbQuery": dbq, "dbRows": dbr, "metadata": json.RawMessage(meta)},
				"parentNodeId": parentID,
			})
		}
		edges := []map[string]any{}
		for _, n := range nodes {
			if p, ok := n["parentNodeId"].(string); ok && p != "" {
				edges = append(edges, map[string]any{"id": p + "-" + n["id"].(string), "source": p, "target": n["id"].(string)})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"nodes": nodes, "edges": edges})
	})

	log.Printf("logs-query listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func env(k, d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }
