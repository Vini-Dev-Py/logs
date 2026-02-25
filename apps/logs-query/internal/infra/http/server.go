package httpx

import (
	"encoding/json"
	"net/http"
	"time"

	"logs-query/internal/infra/cassandra"

	"github.com/go-chi/chi/v5"
)

type Server struct{ repo cassandra.Repo }

func New(repo cassandra.Repo) *Server { return &Server{repo: repo} }

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Get("/query/v1/traces", s.list)
	r.Get("/query/v1/traces/{traceId}", s.byID)
	return r
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	companyID, fromS, toS := r.URL.Query().Get("companyId"), r.URL.Query().Get("from"), r.URL.Query().Get("to")
	from, _ := time.Parse(time.RFC3339, fromS)
	to, _ := time.Parse(time.RFC3339, toS)
	if companyID == "" || from.IsZero() || to.IsZero() {
		http.Error(w, "companyId, from, to required", 400)
		return
	}
	items, err := s.repo.ListTraces(companyID, from, to, r.URL.Query().Get("status"), r.URL.Query().Get("service"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}

func (s *Server) byID(w http.ResponseWriter, r *http.Request) {
	graph, err := s.repo.TraceGraph(chi.URLParam(r, "traceId"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(graph)
}
