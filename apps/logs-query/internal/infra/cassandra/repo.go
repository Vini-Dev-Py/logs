package cassandra

import (
	"encoding/json"
	"sort"
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
	// Generate days in REVERSE order (newest first) to leverage Cassandra's DESC ordering
	var days []string
	current := to
	for !current.Before(from) {
		days = append(days, current.Format("2006-01-02"))
		current = current.AddDate(0, 0, -1)
	}

	// Smart limit: fetch enough rows to fill the page + buffer for filtering
	// For page 1, we can be aggressive; for deeper pages we need more
	fetchLimit := pageSize * 5
	if fetchLimit > 500 {
		fetchLimit = 500 // Cap to avoid OOM
	}

	allTraces := []TraceSummary{}
	collected := 0

	// Query each day partition (newest first - leverages Cassandra DESC ordering)
	for _, day := range days {
		// Stop early if we already have enough for all pages up to requested
		needed := page * pageSize
		if collected >= needed && len(days) > 1 {
			break
		}

		// Use LIMIT to avoid fetching entire day partitions
		q := r.Session.Query(
			"SELECT started_at,trace_id,status,root_operation,service_name,http_method,http_path,http_status,duration_ms FROM traces_by_company_day WHERE company_id=? AND day=? LIMIT ?",
			companyID, day, fetchLimit,
		)
		iter := q.Iter()

		for {
			var s TraceSummary
			var st, rootOp, svc, method, path *string
			var httpStatus, duration *int

			if !iter.Scan(&s.StartedAt, &s.TraceID, &st, &rootOp, &svc, &method, &path, &httpStatus, &duration) {
				break
			}

			// Dereference nullable pointers
			if st != nil {
				s.Status = *st
			}
			if rootOp != nil {
				s.RootOperation = *rootOp
			}
			if svc != nil {
				s.ServiceName = *svc
			}
			if method != nil {
				s.HTTPMethod = *method
			}
			if path != nil {
				s.HTTPPath = *path
			}
			if httpStatus != nil {
				s.HTTPStatus = *httpStatus
			}
			if duration != nil {
				s.DurationMS = *duration
			}

			// Time range filter (Cassandra LIMIT may return rows outside range)
			if s.StartedAt.Before(from) || s.StartedAt.After(to) {
				continue
			}

			// Apply filters
			if status != "" && s.Status != status {
				continue
			}
			if service != "" && s.ServiceName != service {
				continue
			}

			allTraces = append(allTraces, s)
			collected++
		}

		if err := iter.Close(); err != nil {
			return PaginatedResult[TraceSummary]{}, err
		}
	}

	// Data is already mostly sorted (Cassandra DESC within each day)
	// But since we query multiple days, do a final sort
	sort.Slice(allTraces, func(i, j int) bool {
		return allTraces[i].StartedAt.After(allTraces[j].StartedAt)
	})

	total := len(allTraces)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
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

func (r Repo) ListEndpoints(companyID string, from, to time.Time) ([]EndpointSummary, error) {
	// Generate all days in the date range
	var days []string
	current := from
	for !current.After(to) {
		days = append(days, current.Format("2006-01-02"))
		current = current.AddDate(0, 0, 1)
	}

	// Aggregate results by service+method+path
	aggregated := make(map[string]*EndpointSummary)

	for _, day := range days {
		iter := r.Session.Query("SELECT service_name, http_method, http_path, calls FROM endpoint_calls_by_company_day WHERE company_id=? AND day=?", companyID, day).Iter()
		var s EndpointSummary
		for iter.Scan(&s.ServiceName, &s.HTTPMethod, &s.HTTPPath, &s.Calls) {
			key := s.ServiceName + "|" + s.HTTPMethod + "|" + s.HTTPPath
			if existing, ok := aggregated[key]; ok {
				existing.Calls += s.Calls
			} else {
				aggregated[key] = &EndpointSummary{
					ServiceName: s.ServiceName,
					HTTPMethod:  s.HTTPMethod,
					HTTPPath:    s.HTTPPath,
					Calls:       s.Calls,
				}
			}
		}
		if err := iter.Close(); err != nil {
			return nil, err
		}
	}

	// Convert map to slice and sort by calls descending
	out := make([]EndpointSummary, 0, len(aggregated))
	for _, v := range aggregated {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Calls > out[j].Calls
	})

	return out, nil
}
