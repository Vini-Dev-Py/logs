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

type PaginatedResult[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

func (r Repo) ListTraces(companyID string, from, to time.Time, status, service string, page, pageSize int) (PaginatedResult[TraceSummary], error) {
	day := from.Format("2006-01-02")
	iter := r.Session.Query("SELECT started_at,trace_id,status,root_operation,service_name,http_method,http_path,http_status,duration_ms FROM traces_by_company_day WHERE company_id=? AND day=?", companyID, day).Iter()
	
	allTraces := []TraceSummary{}
	var s TraceSummary
	for iter.Scan(&s.StartedAt, &s.TraceID, &s.Status, &s.RootOperation, &s.ServiceName, &s.HTTPMethod, &s.HTTPPath, &s.HTTPStatus, &s.DurationMS) {
		if s.StartedAt.After(from) && s.StartedAt.Before(to) {
			if status != "" && s.Status != status {
				continue
			}
			if service != "" && s.ServiceName != service {
				continue
			}
			allTraces = append(allTraces, s)
		}
	}
	if err := iter.Close(); err != nil {
		return PaginatedResult[TraceSummary]{}, err
	}

	// Sort by StartedAt descending
	for i := 0; i < len(allTraces); i++ {
		for j := i + 1; j < len(allTraces); j++ {
			if allTraces[j].StartedAt.After(allTraces[i].StartedAt) {
				allTraces[i], allTraces[j] = allTraces[j], allTraces[i]
			}
		}
	}

	total := len(allTraces)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Apply pagination
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	var paginatedItems []TraceSummary
	if start < total {
		paginatedItems = allTraces[start:end]
	} else {
		paginatedItems = []TraceSummary{}
	}

	return PaginatedResult[TraceSummary]{
		Items:      paginatedItems,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
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

type EndpointSummary struct {
	ServiceName string `json:"serviceName"`
	HTTPMethod  string `json:"httpMethod"`
	HTTPPath    string `json:"httpPath"`
	Calls       int    `json:"calls"`
}

func (r Repo) ListEndpoints(companyID string, day string) ([]EndpointSummary, error) {
	iter := r.Session.Query("SELECT service_name, http_method, http_path, calls FROM endpoint_calls_by_company_day WHERE company_id=? AND day=?", companyID, day).Iter()
	out := []EndpointSummary{}
	var s EndpointSummary
	for iter.Scan(&s.ServiceName, &s.HTTPMethod, &s.HTTPPath, &s.Calls) {
		out = append(out, s)
	}
	// Note: We might sort out locally descending by calls to easily return the top.
	return out, iter.Close()
}
