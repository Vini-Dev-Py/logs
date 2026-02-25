package httpx

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"logs-ingest/internal/app"
	"logs-ingest/internal/infra/postgres"
	"logs-ingest/internal/infra/rabbit"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Server struct {
	repo  postgres.CompanyRepo
	pub   *rabbit.Publisher
	cache map[string]string
	mu    sync.RWMutex
}

func New(repo postgres.CompanyRepo, pub *rabbit.Publisher) *Server {
	return &Server{repo: repo, pub: pub, cache: map[string]string{}}
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Post("/ingest/v1/log-events", s.ingest)
	return r
}

func (s *Server) ingest(w http.ResponseWriter, r *http.Request) {
	apiKey := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if apiKey == "" {
		http.Error(w, "missing api key", 401)
		return
	}
	s.mu.RLock()
	companyID, ok := s.cache[apiKey]
	s.mu.RUnlock()
	if !ok {
		var err error
		companyID, err = s.repo.CompanyIDByAPIKey(r.Context(), apiKey)
		if err != nil {
			http.Error(w, "invalid api key", 401)
			return
		}
		s.mu.Lock()
		s.cache[apiKey] = companyID
		s.mu.Unlock()
	}
	var evt app.LogEvent
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if evt.EventID == "" {
		evt.EventID = uuid.NewString()
	}
	evt.CompanyID = companyID
	if err := s.pub.Publish(r.Context(), evt); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"accepted": true})
}
