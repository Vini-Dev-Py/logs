package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"logs-bff/internal/config"
	"logs-bff/internal/infra/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	cfg  config.Config
	repo postgres.Repositories
}

func New(cfg config.Config, repo postgres.Repositories) *Server { return &Server{cfg: cfg, repo: repo} }

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Post("/api/auth/login", s.login)
	r.Group(func(pr chi.Router) {
		pr.Use(s.auth)
		pr.Get("/api/me", s.me)
		pr.Get("/api/traces", s.traces)
		pr.Get("/api/traces/{traceId}", s.traceByID)
		pr.Post("/api/traces/{traceId}/annotations", s.createAnnotation)
		pr.Put("/api/annotations/{id}", s.updateAnnotation)
		pr.Delete("/api/annotations/{id}", s.deleteAnnotation)
		pr.Get("/api/search", s.searchNodes)
	})
	return r
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body struct{ Email, Password string }
	_ = json.NewDecoder(r.Body).Decode(&body)
	u, err := s.repo.FindByEmail(r.Context(), body.Email)
	if err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	if err := comparePassword(u.PasswordHash, body.Password); err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": u.ID, "companyId": u.CompanyID, "email": u.Email, "exp": time.Now().Add(24 * time.Hour).Unix()})
	token, _ := t.SignedString([]byte(s.cfg.JWTSecret))
	_ = json.NewEncoder(w).Encode(map[string]any{"token": token, "user": map[string]any{"id": u.ID, "companyId": u.CompanyID, "name": u.Name, "email": u.Email, "role": u.Role}})
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		tok, err := jwt.Parse(h, func(token *jwt.Token) (any, error) { return []byte(s.cfg.JWTSecret), nil })
		if err != nil || !tok.Valid {
			http.Error(w, "unauthorized", 401)
			return
		}
		claims := tok.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), "companyId", claims["companyId"].(string))
		ctx = context.WithValue(ctx, "userId", claims["sub"].(string))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u, err := s.repo.FindByID(r.Context(), r.Context().Value("userId").(string))
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"id": u.ID, "companyId": u.CompanyID, "name": u.Name, "email": u.Email, "role": u.Role})
}

func (s *Server) traces(w http.ResponseWriter, r *http.Request) {
	cid := r.Context().Value("companyId").(string)
	resp, err := http.Get(fmt.Sprintf("%s/query/v1/traces?companyId=%s&from=%s&to=%s&status=%s&service=%s", s.cfg.QueryURL, cid, r.URL.Query().Get("from"), r.URL.Query().Get("to"), r.URL.Query().Get("status"), r.URL.Query().Get("service")))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) traceByID(w http.ResponseWriter, r *http.Request) {
	traceID := chi.URLParam(r, "traceId")
	resp, err := http.Get(fmt.Sprintf("%s/query/v1/traces/%s", s.cfg.QueryURL, traceID))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()
	var payload map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	anns, err := s.repo.ListByTrace(r.Context(), r.Context().Value("companyId").(string), traceID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	payload["annotations"] = anns
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) createAnnotation(w http.ResponseWriter, r *http.Request) {
	var b struct {
		NodeID, Text string
		X, Y         float64
	}
	_ = json.NewDecoder(r.Body).Decode(&b)
	id, err := s.repo.Create(r.Context(), r.Context().Value("companyId").(string), r.Context().Value("userId").(string), chi.URLParam(r, "traceId"), b.NodeID, b.X, b.Y, b.Text)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (s *Server) updateAnnotation(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Text string  `json:"text"`
		X    float64 `json:"x"`
		Y    float64 `json:"y"`
	}
	_ = json.NewDecoder(r.Body).Decode(&b)
	if err := s.repo.Update(r.Context(), r.Context().Value("companyId").(string), chi.URLParam(r, "id"), b.Text, b.X, b.Y); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteAnnotation(w http.ResponseWriter, r *http.Request) {
	if err := s.repo.Delete(r.Context(), r.Context().Value("companyId").(string), chi.URLParam(r, "id")); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) searchNodes(w http.ResponseWriter, r *http.Request) {
	cid := r.Context().Value("companyId").(string)
	query := r.URL.Query().Get("query")

	if query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("%s/query/v1/search?companyId=%s&query=%s", s.cfg.QueryURL, cid, query)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
