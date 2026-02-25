package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type LogEvent struct {
	EventID      string                 `json:"eventId"`
	TraceID      string                 `json:"traceId"`
	NodeID       string                 `json:"nodeId"`
	ParentNodeID *string                `json:"parentNodeId"`
	ServiceName  string                 `json:"serviceName"`
	Operation    map[string]any         `json:"operation"`
	HTTP         map[string]any         `json:"http"`
	DB           map[string]any         `json:"db"`
	Metadata     map[string]any         `json:"metadata"`
	CompanyID    string                 `json:"companyId,omitempty"`
	APIKey       string                 `json:"apiKey,omitempty"`
}

func main() {
	port := env("PORT", "8082")
	db, err := pgxpool.New(context.Background(), env("DATABASE_URL", ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	conn, err := amqp.Dial(env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"))
	if err != nil {
		log.Fatal(err)
	}
	ch, _ := conn.Channel()
	_, _ = ch.QueueDeclare("log_events", true, false, false, false, nil)

	cache := map[string]string{}
	var mu sync.RWMutex

	r := chi.NewRouter()
	r.Post("/ingest/v1/log-events", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		if len(apiKey) > 7 {
			apiKey = apiKey[7:]
		}
		if apiKey == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}

		mu.RLock()
		companyID, ok := cache[apiKey]
		mu.RUnlock()
		if !ok {
			err = db.QueryRow(r.Context(), "SELECT id::text FROM companies WHERE api_key=$1", apiKey).Scan(&companyID)
			if err != nil {
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}
			mu.Lock()
			cache[apiKey] = companyID
			mu.Unlock()
		}

		var evt LogEvent
		if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if evt.EventID == "" {
			evt.EventID = uuid.NewString()
		}
		evt.CompanyID = companyID

		body, _ := json.Marshal(evt)
		_ = ch.PublishWithContext(r.Context(), "", "log_events", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"accepted":true}`))
	})

	log.Printf("logs-ingest listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
