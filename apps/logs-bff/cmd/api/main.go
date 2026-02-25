package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string `json:"id"`
	CompanyID string `json:"companyId"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

func main() {
	port := env("PORT", "8081")
	jwtSecret := env("JWT_SECRET", "secret")
	queryURL := env("QUERY_URL", "http://localhost:8084")
	db, err := pgxpool.New(context.Background(), env("DATABASE_URL", ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Email, Password string }
		_ = json.NewDecoder(r.Body).Decode(&body)
		var u User
		var hash string
		err := db.QueryRow(r.Context(), "SELECT id::text, company_id::text, name, email, role, password_hash FROM users WHERE email=$1", body.Email).Scan(&u.ID, &u.CompanyID, &u.Name, &u.Email, &u.Role, &hash)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)) != nil {
			http.Error(w, "invalid credentials", 401)
			return
		}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": u.ID, "companyId": u.CompanyID, "email": u.Email, "exp": time.Now().Add(24 * time.Hour).Unix()})
		token, _ := t.SignedString([]byte(jwtSecret))
		_ = json.NewEncoder(w).Encode(map[string]any{"token": token, "user": u})
	})

	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			tok, err := jwt.Parse(h, func(token *jwt.Token) (any, error) { return []byte(jwtSecret), nil })
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

	r.Group(func(pr chi.Router) {
		pr.Use(auth)
		pr.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			uid := r.Context().Value("userId").(string)
			var u User
			_ = db.QueryRow(r.Context(), "SELECT id::text, company_id::text, name, email, role FROM users WHERE id=$1", uid).Scan(&u.ID, &u.CompanyID, &u.Name, &u.Email, &u.Role)
			_ = json.NewEncoder(w).Encode(u)
		})

		pr.Get("/api/traces", func(w http.ResponseWriter, r *http.Request) {
			cid := r.Context().Value("companyId").(string)
			from := r.URL.Query().Get("from")
			to := r.URL.Query().Get("to")
			resp, err := http.Get(fmt.Sprintf("%s/query/v1/traces?companyId=%s&from=%s&to=%s", queryURL, cid, from, to))
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer resp.Body.Close()
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
		})

		pr.Get("/api/traces/{traceId}", func(w http.ResponseWriter, r *http.Request) {
			traceID := chi.URLParam(r, "traceId")
			resp, err := http.Get(fmt.Sprintf("%s/query/v1/traces/%s", queryURL, traceID))
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer resp.Body.Close()
			var payload map[string]any
			_ = json.NewDecoder(resp.Body).Decode(&payload)
			cid := r.Context().Value("companyId").(string)
			rows, _ := db.Query(r.Context(), "SELECT id::text,node_id,x,y,text,created_at FROM annotations WHERE company_id=$1 AND trace_id=$2", cid, traceID)
			anns := []map[string]any{}
			for rows.Next() {
				var id, nodeID, text string
				var x, y float64
				var createdAt time.Time
				_ = rows.Scan(&id, &nodeID, &x, &y, &text, &createdAt)
				anns = append(anns, map[string]any{"id": id, "nodeId": nodeID, "x": x, "y": y, "text": text, "createdAt": createdAt})
			}
			payload["annotations"] = anns
			_ = json.NewEncoder(w).Encode(payload)
		})

		pr.Post("/api/traces/{traceId}/annotations", func(w http.ResponseWriter, r *http.Request) {
			cid := r.Context().Value("companyId").(string)
			uid := r.Context().Value("userId").(string)
			traceID := chi.URLParam(r, "traceId")
			var b struct {
				NodeID string  `json:"nodeId"`
				Text   string  `json:"text"`
				X      float64 `json:"x"`
				Y      float64 `json:"y"`
			}
			_ = json.NewDecoder(r.Body).Decode(&b)
			id := uuid.NewString()
			_, err := db.Exec(r.Context(), "INSERT INTO annotations(id,company_id,trace_id,node_id,x,y,text,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8)", id, cid, traceID, b.NodeID, b.X, b.Y, b.Text, uid)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
		})

		pr.Put("/api/annotations/{id}", func(w http.ResponseWriter, r *http.Request) {
			cid := r.Context().Value("companyId").(string)
			id := chi.URLParam(r, "id")
			var b struct { Text string `json:"text"`; X float64 `json:"x"`; Y float64 `json:"y"` }
			_ = json.NewDecoder(r.Body).Decode(&b)
			_, err := db.Exec(r.Context(), "UPDATE annotations SET text=$1,x=$2,y=$3 WHERE id=$4 AND company_id=$5", b.Text, b.X, b.Y, id, cid)
			if err != nil { http.Error(w, err.Error(), 500); return }
			w.WriteHeader(http.StatusNoContent)
		})

		pr.Delete("/api/annotations/{id}", func(w http.ResponseWriter, r *http.Request) {
			cid := r.Context().Value("companyId").(string)
			id := chi.URLParam(r, "id")
			_, err := db.Exec(r.Context(), "DELETE FROM annotations WHERE id=$1 AND company_id=$2", id, cid)
			if err != nil { http.Error(w, err.Error(), 500); return }
			w.WriteHeader(http.StatusNoContent)
		})

	})

	log.Printf("logs-bff listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
