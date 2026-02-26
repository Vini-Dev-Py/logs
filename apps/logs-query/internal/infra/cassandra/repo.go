package cassandra

import (
	"encoding/json"
	"time"

	"github.com/gocql/gocql"
)

type Repo struct{ Session *gocql.Session }

type TraceSummary struct {
	StartedAt     time.Time `json:"startedAt"`
	TraceID       string    `json:"traceId"`
	Status        string    `json:"status"`
	RootOperation string    `json:"rootOperation"`
	ServiceName   string    `json:"serviceName"`
	HTTPMethod    string    `json:"httpMethod"`
	HTTPPath      string    `json:"httpPath"`
	HTTPStatus    int       `json:"httpStatus"`
	DurationMS    int       `json:"durationMs"`
}

func (r Repo) ListTraces(companyID string, from, to time.Time, status, service string) ([]TraceSummary, error) {
	day := from.Format("2006-01-02")
	iter := r.Session.Query("SELECT started_at,trace_id,status,root_operation,service_name,http_method,http_path,http_status,duration_ms FROM traces_by_company_day WHERE company_id=? AND day=?", companyID, day).Iter()
	out := []TraceSummary{}
	var s TraceSummary
	for iter.Scan(&s.StartedAt, &s.TraceID, &s.Status, &s.RootOperation, &s.ServiceName, &s.HTTPMethod, &s.HTTPPath, &s.HTTPStatus, &s.DurationMS) {
		if s.StartedAt.After(from) && s.StartedAt.Before(to) {
			if status != "" && s.Status != status {
				continue
			}
			if service != "" && s.ServiceName != service {
				continue
			}
			out = append(out, s)
		}
	}
	return out, iter.Close()
}

func (r Repo) TraceGraph(traceID string) (map[string]any, error) {
	iter := r.Session.Query("SELECT node_id,parent_node_id,service_name,type,name,status,start_at,end_at,duration_ms,http_method,http_path,http_status,db_system,db_query,db_rows,metadata FROM nodes_by_trace WHERE trace_id=?", traceID).Iter()
	var nodes []map[string]any
	var nodeID, parentID, serviceName, typ, name, status, hm, hp, dbs, dbq, meta string
	var start, end time.Time
	var duration, hs, dbr int
	for iter.Scan(&nodeID, &parentID, &serviceName, &typ, &name, &status, &start, &end, &duration, &hm, &hp, &hs, &dbs, &dbq, &dbr, &meta) {
		nodes = append(nodes, map[string]any{"id": nodeID, "data": map[string]any{"label": name, "type": typ, "status": status, "serviceName": serviceName, "durationMs": duration, "startAt": start, "endAt": end, "httpMethod": hm, "httpPath": hp, "httpStatus": hs, "dbSystem": dbs, "dbQuery": dbq, "dbRows": dbr, "metadata": json.RawMessage(meta)}, "parentNodeId": parentID})
	}
	edges := []map[string]any{}
	for _, n := range nodes {
		if p, ok := n["parentNodeId"].(string); ok && p != "" {
			edges = append(edges, map[string]any{"id": p + "-" + n["id"].(string), "source": p, "target": n["id"].(string)})
		}
	}
	return map[string]any{"nodes": nodes, "edges": edges}, iter.Close()
}
