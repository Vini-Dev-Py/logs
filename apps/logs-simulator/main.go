package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// ─── Config ────────────────────────────────────────────────────────────────

type Config struct {
	Port        string
	TargetURL   string
	OTLPURL     string
	APIKey      string
	Mode        string
	Workers     int
	BurstSize   int
	Bursts      int
	BurstDelay  time.Duration
	SearchURL   string
	SearchQuery string
	CompanyID   string
}

func loadConfig() Config {
	return Config{
		Port:        env("PORT", "8085"),
		TargetURL:   env("TARGET_URL", "http://logs-ingest:8082/ingest/v1/log-events"),
		OTLPURL:     env("OTLP_URL", "http://logs-ingest:8082/v1/traces"),
		APIKey:      env("API_KEY", "logs_dev_api_key"),
		Mode:        env("MODE", "native"),
		Workers:     parseInt(env("WORKERS", "10")),
		BurstSize:   parseInt(env("BURST_SIZE", "50")),
		Bursts:      parseInt(env("BURSTS", "20")),
		BurstDelay:  parseDuration(env("BURST_DELAY", "2s")),
		SearchURL:   env("SEARCH_URL", "http://logs-bff:8081/api/search"),
		SearchQuery: env("SEARCH_QUERY", "SELECT"),
		CompanyID:   env("COMPANY_ID", "c152767f-4e9b-4311-b896-a7a37c9ccbfe"),
	}
}

// ─── Simulator Server ──────────────────────────────────────────────────────

type SimulatorServer struct {
	cfg     Config
	client  *http.Client
	current *SimulationState
	mu      sync.RWMutex
}

type SimulationState struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"` // "running", "completed", "failed"
	StartedAt     time.Time `json:"startedAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	Config        SimConfig `json:"config"`
	TotalSent     int64     `json:"totalSent"`
	TotalOK       int64     `json:"totalOk"`
	TotalFail     int64     `json:"totalFail"`
	TotalSearchOK int64     `json:"totalSearchOk"`
	TotalSearchFail int64   `json:"totalSearchFail"`
}

type SimConfig struct {
	Mode       string `json:"mode"`
	BurstSize  int    `json:"burstSize"`
	Bursts     int    `json:"bursts"`
	CompanyID  string `json:"companyId"`
}

type SimRequest struct {
	BurstSize  *int    `json:"burstSize,omitempty"`
	Bursts     *int    `json:"bursts,omitempty"`
	Mode       *string `json:"mode,omitempty"`
	CompanyID  *string `json:"companyId,omitempty"`
}

func NewSimulatorServer(cfg Config) *SimulatorServer {
	return &SimulatorServer{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SimulatorServer) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /simulate", s.handleStart)
	mux.HandleFunc("GET /simulate", s.handleStatus)
	mux.HandleFunc("POST /simulate/stop", s.handleStop)
	mux.HandleFunc("GET /health", s.handleHealth)

	// Seed endpoint for backwards compatibility
	mux.HandleFunc("POST /seed", s.handleStart)

	server := &http.Server{
		Addr:    ":" + s.cfg.Port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	fmt.Printf("🎭 Simulator server listening on :%s\n", s.cfg.Port)
	fmt.Printf("   POST /simulate  - Start simulation\n")
	fmt.Printf("   GET  /simulate  - Get current simulation status\n")
	fmt.Printf("   POST /simulate/stop - Stop running simulation\n")
	fmt.Printf("   POST /seed      - Start simulation (alias)\n")
	fmt.Printf("   GET  /health    - Health check\n\n")

	return server.ListenAndServe()
}

func (s *SimulatorServer) handleStart(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.current != nil && s.current.Status == "running" {
		s.mu.Unlock()
		http.Error(w, `{"error":"simulation already running"}`, http.StatusConflict)
		return
	}
	s.mu.Unlock()

	// Parse optional request body
	req := SimRequest{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid json: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}
	}

	// Build simulation config
	cfg := SimConfig{
		Mode:      s.cfg.Mode,
		BurstSize: s.cfg.BurstSize,
		Bursts:    s.cfg.Bursts,
		CompanyID: s.cfg.CompanyID,
	}

	if req.Mode != nil {
		cfg.Mode = *req.Mode
	}
	if req.BurstSize != nil {
		cfg.BurstSize = *req.BurstSize
	}
	if req.Bursts != nil {
		cfg.Bursts = *req.Bursts
	}
	if req.CompanyID != nil {
		cfg.CompanyID = *req.CompanyID
	}

	sim := &SimulationState{
		ID:        uuid.New().String(),
		Status:    "running",
		StartedAt: time.Now(),
		Config:    cfg,
	}

	s.mu.Lock()
	s.current = sim
	s.mu.Unlock()

	// Run simulation in background
	go s.runSimulation(sim, cfg)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]any{
		"id":     sim.ID,
		"status": "running",
		"message": fmt.Sprintf("Simulation started: %d bursts x %d traces", cfg.Bursts, cfg.BurstSize),
	})
}

func (s *SimulatorServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.current == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "idle",
			"message": "No simulation has been run",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.current)
}

func (s *SimulatorServer) handleStop(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	if s.current == nil || s.current.Status != "running" {
		s.mu.RUnlock()
		http.Error(w, `{"error":"no simulation running"}`, http.StatusNotFound)
		return
	}
	s.mu.RUnlock()

	s.mu.Lock()
	s.current.Status = "stopped"
	now := time.Now()
	s.current.CompletedAt = &now
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "stopped",
	})
}

func (s *SimulatorServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "healthy",
		"service": "logs-simulator",
	})
}

func (s *SimulatorServer) runSimulation(sim *SimulationState, cfg SimConfig) {
	defer func() {
		if r := recover(); r != nil {
			s.mu.Lock()
			sim.Status = "failed"
			sim.TotalFail++
			now := time.Now()
			sim.CompletedAt = &now
			s.mu.Unlock()
			fmt.Printf("❌ Simulation panicked: %v\n", r)
		}
	}()

	fmt.Printf("\n╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  🚀 Simulation #%s started                              ║\n", sim.ID[:8])
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
	fmt.Printf("  Mode:       %s\n", cfg.Mode)
	fmt.Printf("  Target:     %s\n", s.cfg.TargetURL)
	fmt.Printf("  Workers:    %d\n", s.cfg.Workers)
	fmt.Printf("  Burst Size: %d traces/burst\n", cfg.BurstSize)
	fmt.Printf("  Bursts:     %d\n", cfg.Bursts)
	fmt.Printf("  CompanyID:  %s\n\n", cfg.CompanyID)

	// Wait for target to be ready
	fmt.Print("⏳ Waiting for target to be ready... ")
	ready := false
	for i := 0; i < 10; i++ {
		payload := []byte(`{"traceId":"hc","nodeId":"hc","serviceName":"hc","operation":{"type":"HTTP","name":"HC","status":"OK","startAt":"2026-01-01T00:00:00Z","endAt":"2026-01-01T00:00:01Z","durationMs":1}}`)
		req, _ := http.NewRequest("POST", s.cfg.TargetURL, bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.client.Do(req)
		if err == nil {
			code := resp.StatusCode
			resp.Body.Close()
			if code == 202 || code == 401 {
				ready = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}
	if !ready {
		fmt.Println("❌ Target not reachable!")
		s.mu.Lock()
		sim.Status = "failed"
		sim.TotalFail++
		now := time.Now()
		sim.CompletedAt = &now
		s.mu.Unlock()
		return
	}
	fmt.Println("✅ Ready!")
	fmt.Println()

	startTime := time.Now()

	for burst := 0; burst < cfg.Bursts; burst++ {
		// Check if simulation was stopped
		s.mu.RLock()
		if sim.Status != "running" {
			s.mu.RUnlock()
			fmt.Println("⏹️  Simulation stopped by user")
			return
		}
		s.mu.RUnlock()

		burstStart := time.Now()
		fmt.Printf("━━━ Burst %d/%d ━━━\n", burst+1, cfg.Bursts)

		// Generate traces
		traceIDs := make([]string, cfg.BurstSize)
		allEvents := make([][]LogEvent, cfg.BurstSize)
		for i := 0; i < cfg.BurstSize; i++ {
			traceIDs[i] = uuid.New().String()
			scenarioFn := scenarios[rand.Intn(len(scenarios))]
			allEvents[i] = scenarioFn(traceIDs[i], cfg.CompanyID)
		}

		// Send via workers
		var wg sync.WaitGroup
		sem := make(chan struct{}, s.cfg.Workers)

		for i := 0; i < cfg.BurstSize; i++ {
			wg.Add(1)
			sem <- struct{}{}
			go func(idx int) {
				defer wg.Done()
				defer func() { <-sem }()

				events := allEvents[idx]
				atomic.AddInt64(&sim.TotalSent, int64(len(events)))

				if cfg.Mode == "otlp" {
					sendOTLPBatch(s.client, s.cfg.OTLPURL, s.cfg.APIKey, events, &sim.TotalOK, &sim.TotalFail)
				} else {
					for _, evt := range events {
						sendNative(s.client, s.cfg.TargetURL, s.cfg.APIKey, evt, &sim.TotalOK, &sim.TotalFail)
					}
				}
			}(i)
		}

		wg.Wait()
		burstDuration := time.Since(burstStart)
		fmt.Printf("  ✅ Burst complete in %v (%d events sent)\n\n", burstDuration.Round(time.Millisecond), atomic.LoadInt64(&sim.TotalSent))

		// Run search every 5 bursts
		if (burst+1)%5 == 0 && s.cfg.SearchQuery != "" {
			fmt.Println("🔍 Running search validation...")
			runSearch(s.client, s.cfg.SearchURL, s.cfg.APIKey, s.cfg.SearchQuery, &sim.TotalSearchOK, &sim.TotalSearchFail)
			runSearch(s.client, s.cfg.SearchURL, s.cfg.APIKey, "INSERT", &sim.TotalSearchOK, &sim.TotalSearchFail)
			runSearch(s.client, s.cfg.SearchURL, s.cfg.APIKey, "SELECT", &sim.TotalSearchOK, &sim.TotalSearchFail)
			fmt.Println()
		}

		if burst < cfg.Bursts-1 {
			time.Sleep(s.cfg.BurstDelay)
		}
	}

	// Mark as completed
	elapsed := time.Since(startTime)
	s.mu.Lock()
	sim.Status = "completed"
	now := time.Now()
	sim.CompletedAt = &now
	s.mu.Unlock()

	// Final report
	fmt.Printf("\n╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  📊 Simulation #%s completed                            ║\n", sim.ID[:8])
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
	fmt.Printf("  Total Events Sent:    %d\n", atomic.LoadInt64(&sim.TotalSent))
	fmt.Printf("  Successful (2xx):     %d\n", atomic.LoadInt64(&sim.TotalOK))
	fmt.Printf("  Failed:               %d\n", atomic.LoadInt64(&sim.TotalFail))
	fmt.Printf("  Search Queries OK:    %d\n", atomic.LoadInt64(&sim.TotalSearchOK))
	fmt.Printf("  Search Queries Fail:  %d\n", atomic.LoadInt64(&sim.TotalSearchFail))
	fmt.Printf("  Total Time:           %v\n", elapsed.Round(time.Millisecond))

	sent := float64(atomic.LoadInt64(&sim.TotalSent))
	if elapsed.Seconds() > 0 {
		fmt.Printf("  Throughput:           %.0f events/sec\n", sent/elapsed.Seconds())
	}

	if atomic.LoadInt64(&sim.TotalSent) > 0 {
		successRate := float64(atomic.LoadInt64(&sim.TotalOK)) / float64(atomic.LoadInt64(&sim.TotalSent)) * 100
		fmt.Printf("  Success Rate:         %.1f%%\n", successRate)
	}

	if atomic.LoadInt64(&sim.TotalFail) == 0 {
		fmt.Println("\n🎉 All events ingested successfully!")
	} else {
		fmt.Printf("\n⚠️  %d events failed — check logs for details\n", atomic.LoadInt64(&sim.TotalFail))
	}
}

// ─── LogEvent & Scenarios ──────────────────────────────────────────────────

type LogEvent struct {
	TraceID      string         `json:"traceId"`
	NodeID       string         `json:"nodeId"`
	ParentNodeID *string        `json:"parentNodeId,omitempty"`
	ServiceName  string         `json:"serviceName"`
	CompanyID    string         `json:"companyId"`
	EventID      string         `json:"eventId,omitempty"`
	Operation    struct {
		Type       string `json:"type"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		StartAt    string `json:"startAt"`
		EndAt      string `json:"endAt"`
		DurationMs int    `json:"durationMs"`
	} `json:"operation"`
	HTTP     map[string]any `json:"http,omitempty"`
	DB       map[string]any `json:"db,omitempty"`
	Metadata string         `json:"metadata,omitempty"`
}

type ScenarioFn func(traceID, companyID string) []LogEvent

var scenarios = []ScenarioFn{
	scenarioECommerce,
	scenarioPaymentProcessing,
	scenarioUserAuthentication,
	scenarioReportGeneration,
	scenarioWebhookDelivery,
	scenarioDataSync,
}

func scenarioECommerce(traceID, companyID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	node1 := makeNode(traceID, "", "api-gateway", "HTTP", "GET /api/products", "OK", now, 12, companyID)
	events = append(events, node1)

	node2 := makeNode(traceID, node1.NodeID, "api-gateway", "HTTP", "POST /auth/verify", "OK", now.Add(13*time.Millisecond), 8, companyID)
	events = append(events, node2)

	node3 := makeNode(traceID, node1.NodeID, "product-service", "HTTP", "GET /products", "OK", now.Add(25*time.Millisecond), 45, companyID)
	events = append(events, node3)

	node4 := makeNodeWithDB(traceID, node3.NodeID, "product-service", "DB", "SELECT products", "OK", now.Add(30*time.Millisecond), 20,
		"postgres", "SELECT * FROM products WHERE category = 'electronics' AND active = true ORDER BY created_at DESC LIMIT 50", 48, companyID)
	events = append(events, node4)

	node5 := makeNode(traceID, node1.NodeID, "product-service", "HTTP", "GET redis://cache", "OK", now.Add(55*time.Millisecond), 3, companyID)
	events = append(events, node5)

	node6 := makeNode(traceID, node1.NodeID, "recommendation-svc", "HTTP", "POST /recommend", "OK", now.Add(60*time.Millisecond), 120, companyID)
	events = append(events, node6)

	node7 := makeNode(traceID, node6.NodeID, "ml-engine", "CUSTOM", "predict(user_profile)", "OK", now.Add(65*time.Millisecond), 85, companyID)
	events = append(events, node7)

	return events
}

func scenarioPaymentProcessing(traceID, companyID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	node1 := makeNode(traceID, "", "checkout-api", "HTTP", "POST /checkout", "OK", now, 250, companyID)
	events = append(events, node1)

	status := "OK"
	if rand.Float64() < 0.05 {
		status = "ERROR"
	}

	node2 := makeNode(traceID, node1.NodeID, "order-service", "HTTP", "POST /orders", status, now.Add(15*time.Millisecond), 180, companyID)
	events = append(events, node2)

	node3 := makeNodeWithDB(traceID, node2.NodeID, "order-service", "DB", "INSERT orders", status, now.Add(20*time.Millisecond), 25,
		"postgres", fmt.Sprintf("INSERT INTO orders (id, user_id, total, status) VALUES ('%s', 42, 299.90, 'pending')", uuid.New().String()[:8]), 1, companyID)
	events = append(events, node3)

	node4 := makeNode(traceID, node1.NodeID, "payment-gateway", "EXTERNAL_API", "POST https://api.stripe.com/v1/charges", status, now.Add(30*time.Millisecond), 350, companyID)
	events = append(events, node4)

	if status == "OK" {
		node5 := makeNode(traceID, node1.NodeID, "notification-svc", "HTTP", "POST rabbitmq://queue/email", "OK", now.Add(400*time.Millisecond), 5, companyID)
		events = append(events, node5)
	} else {
		node5 := makeNode(traceID, node4.NodeID, "payment-gateway", "HTTP", "POST /rollback", "OK", now.Add(380*time.Millisecond), 15, companyID)
		events = append(events, node5)

		node6 := makeNodeWithDB(traceID, node2.NodeID, "order-service", "DB", "UPDATE orders", "OK", now.Add(400*time.Millisecond), 10,
			"postgres", fmt.Sprintf("UPDATE orders SET status = 'failed' WHERE id = '%s'", uuid.New().String()[:8]), 1, companyID)
		events = append(events, node6)
	}

	return events
}

func scenarioUserAuthentication(traceID, companyID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	node1 := makeNode(traceID, "", "auth-service", "HTTP", "POST /auth/login", "OK", now, 85, companyID)
	events = append(events, node1)

	node2 := makeNodeWithDB(traceID, node1.NodeID, "auth-service", "DB", "SELECT users", "OK", now.Add(5*time.Millisecond), 15,
		"postgres", "SELECT id, email, password_hash, role FROM users WHERE email = 'user@example.com' LIMIT 1", 1, companyID)
	events = append(events, node2)

	node3 := makeNode(traceID, node1.NodeID, "auth-service", "CUSTOM", "generate_jwt", "OK", now.Add(25*time.Millisecond), 3, companyID)
	events = append(events, node3)

	node4 := makeNode(traceID, node1.NodeID, "session-store", "HTTP", "SET redis://session", "OK", now.Add(30*time.Millisecond), 8, companyID)
	events = append(events, node4)

	node5 := makeNode(traceID, node1.NodeID, "audit-service", "HTTP", "POST kafka://audit", "OK", now.Add(40*time.Millisecond), 12, companyID)
	events = append(events, node5)

	return events
}

func scenarioReportGeneration(traceID, companyID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	node1 := makeNode(traceID, "", "report-api", "HTTP", "GET /reports/monthly", "OK", now, 2500, companyID)
	events = append(events, node1)

	node2 := makeNodeWithDB(traceID, node1.NodeID, "report-api", "DB", "SELECT aggregations", "OK", now.Add(10*time.Millisecond), 1800,
		"postgres", `SELECT DATE_TRUNC('month', created_at) AS month, COUNT(*) AS total, SUM(amount) AS revenue
FROM transactions
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY month ORDER BY month DESC`, 0, companyID)
	events = append(events, node2)

	node3 := makeNode(traceID, node1.NodeID, "report-api", "HTTP", "GET redis://cache", "ERROR", now.Add(15*time.Millisecond), 2, companyID)
	events = append(events, node3)

	node4 := makeNode(traceID, node1.NodeID, "pdf-generator", "CUSTOM", "render_template", "OK", now.Add(1850*time.Millisecond), 400, companyID)
	events = append(events, node4)

	node5 := makeNode(traceID, node1.NodeID, "storage-service", "EXTERNAL_API", "PUT https://s3.amazonaws.com/reports", "OK", now.Add(2300*time.Millisecond), 150, companyID)
	events = append(events, node5)

	return events
}

func scenarioWebhookDelivery(traceID, companyID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	node1 := makeNode(traceID, "", "webhook-dispatcher", "HTTP", "POST /dispatch", "OK", now, 5, companyID)
	events = append(events, node1)

	node2 := makeNode(traceID, node1.NodeID, "webhook-dispatcher", "HTTP", "POST rabbitmq://webhooks", "OK", now.Add(2*time.Millisecond), 3, companyID)
	events = append(events, node2)

	node3 := makeNode(traceID, node2.NodeID, "webhook-worker", "HTTP", "GET rabbitmq://webhooks/consume", "OK", now.Add(50*time.Millisecond), 200, companyID)
	events = append(events, node3)

	status := "OK"
	if rand.Float64() < 0.1 {
		status = "ERROR"
	}

	node4 := makeNode(traceID, node3.NodeID, "webhook-worker", "EXTERNAL_API", "POST https://customer.example.com/hook", status, now.Add(60*time.Millisecond), 450, companyID)
	events = append(events, node4)

	if status == "OK" {
		node5 := makeNodeWithDB(traceID, node3.NodeID, "webhook-worker", "DB", "UPDATE deliveries", "OK", now.Add(520*time.Millisecond), 8,
			"postgres", fmt.Sprintf("UPDATE webhook_deliveries SET status = 'delivered', delivered_at = NOW() WHERE id = '%s'", uuid.New().String()[:8]), 1, companyID)
		events = append(events, node5)
	}

	return events
}

func scenarioDataSync(traceID, companyID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	node1 := makeNode(traceID, "", "sync-service", "CUSTOM", "scheduled_sync", "OK", now, 5000, companyID)
	events = append(events, node1)

	node2 := makeNodeWithDB(traceID, node1.NodeID, "sync-service", "DB", "SELECT changes", "OK", now.Add(10*time.Millisecond), 2000,
		"postgres", "SELECT * FROM cdc_log WHERE synced = false ORDER BY created_at LIMIT 1000", 0, companyID)
	events = append(events, node2)

	node3 := makeNode(traceID, node1.NodeID, "transform-engine", "CUSTOM", "map_fields", "OK", now.Add(2050*time.Millisecond), 150, companyID)
	events = append(events, node3)

	node4 := makeNode(traceID, node3.NodeID, "sync-service", "EXTERNAL_API", "POST https://partner-api.example.com/sync", "OK", now.Add(2200*time.Millisecond), 1200, companyID)
	events = append(events, node4)

	node5 := makeNodeWithDB(traceID, node1.NodeID, "sync-service", "DB", "UPDATE cdc_log", "OK", now.Add(3450*time.Millisecond), 50,
		"postgres", "UPDATE cdc_log SET synced = true, synced_at = NOW() WHERE id IN (SELECT id FROM cdc_log WHERE synced = false LIMIT 1000)", 0, companyID)
	events = append(events, node5)

	return events
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func makeNode(traceID, parentNodeID, service, opType, name, status string, start time.Time, durationMs int, companyID string) LogEvent {
	evt := LogEvent{
		TraceID:     traceID,
		NodeID:      uuid.New().String(),
		ServiceName: service,
		CompanyID:   companyID,
		EventID:     uuid.New().String(),
	}
	if parentNodeID != "" {
		evt.ParentNodeID = &parentNodeID
	}
	evt.Operation.Type = opType
	evt.Operation.Name = name
	evt.Operation.Status = status
	evt.Operation.StartAt = start.Format(time.RFC3339Nano)
	evt.Operation.EndAt = start.Add(time.Duration(durationMs) * time.Millisecond).Format(time.RFC3339Nano)
	evt.Operation.DurationMs = durationMs

	if opType == "HTTP" || opType == "EXTERNAL_API" {
		method, path := "GET", "/"
		if len(name) > 0 {
			fmt.Sscanf(name, "%s %s", &method, &path)
		}
		evt.HTTP = map[string]any{
			"method":     method,
			"path":       path,
			"statusCode": 200,
		}
		if status == "ERROR" {
			evt.HTTP["statusCode"] = 500
		}
	}

	return evt
}

func makeNodeWithDB(traceID, parentNodeID, service, opType, name, status string, start time.Time, durationMs int, dbSystem, dbQuery string, dbRows int, companyID string) LogEvent {
	evt := makeNode(traceID, parentNodeID, service, opType, name, status, start, durationMs, companyID)
	evt.DB = map[string]any{
		"system": dbSystem,
		"query":  dbQuery,
	}
	if dbRows > 0 {
		evt.DB["rows"] = dbRows
	}
	return evt
}

// ─── Native Sender ─────────────────────────────────────────────────────────

func sendNative(client *http.Client, url, apiKey string, evt LogEvent, okCounter, failCounter *int64) {
	body := map[string]any{
		"traceId":      evt.TraceID,
		"nodeId":       evt.NodeID,
		"parentNodeId": evt.ParentNodeID,
		"serviceName":  evt.ServiceName,
		"companyId":    evt.CompanyID,
		"operation":    evt.Operation,
	}
	if evt.HTTP != nil {
		body["http"] = evt.HTTP
	}
	if evt.DB != nil {
		body["db"] = evt.DB
	}
	if evt.Metadata != "" {
		body["metadata"] = evt.Metadata
	}

	payload, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt64(failCounter, 1)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(okCounter, 1)
	} else {
		atomic.AddInt64(failCounter, 1)
	}
}

// ─── OTLP Sender ───────────────────────────────────────────────────────────

func sendOTLPBatch(client *http.Client, url, apiKey string, events []LogEvent, okCounter, failCounter *int64) {
	type attrVal struct {
		StringValue *string `json:"stringValue,omitempty"`
		IntValue    *string `json:"intValue,omitempty"`
	}
	type attrKV struct {
		Key   string   `json:"key"`
		Value attrVal `json:"value"`
	}

	type span struct {
		TraceID           string    `json:"traceId"`
		SpanID            string    `json:"spanId"`
		ParentSpanID      string    `json:"parentSpanId"`
		Name              string    `json:"name"`
		Kind              int       `json:"kind"`
		StartTimeUnixNano string    `json:"startTimeUnixNano"`
		EndTimeUnixNano   string    `json:"endTimeUnixNano"`
		Attributes        []attrKV  `json:"attributes"`
		Status            struct {
			Code int `json:"code"`
		} `json:"status"`
	}

	var spans []span
	for _, evt := range events {
		t, _ := time.Parse(time.RFC3339Nano, evt.Operation.StartAt)
		e, _ := time.Parse(time.RFC3339Nano, evt.Operation.EndAt)

		s := span{
			TraceID:           hexUUID(evt.TraceID),
			SpanID:            hexUUID(evt.NodeID),
			Name:              evt.Operation.Name,
			StartTimeUnixNano: fmt.Sprintf("%d", t.UnixNano()),
			EndTimeUnixNano:   fmt.Sprintf("%d", e.UnixNano()),
		}

		if evt.ParentNodeID != nil && *evt.ParentNodeID != "" {
			s.ParentSpanID = hexUUID(*evt.ParentNodeID)
		}

		if evt.Operation.Status == "ERROR" {
			s.Status.Code = 2
		} else {
			s.Status.Code = 1
		}

		switch evt.Operation.Type {
		case "HTTP":
			s.Kind = 2
		case "EXTERNAL_API":
			s.Kind = 3
		case "DB":
			s.Kind = 2
		default:
			s.Kind = 1
		}

		if evt.HTTP != nil {
			if m, ok := evt.HTTP["method"].(string); ok {
				s.Attributes = append(s.Attributes, attrKV{"http.method", attrVal{StringValue: &m}})
			}
			if p, ok := evt.HTTP["path"].(string); ok {
				s.Attributes = append(s.Attributes, attrKV{"http.target", attrVal{StringValue: &p}})
			}
		}
		if evt.DB != nil {
			if sys, ok := evt.DB["system"].(string); ok {
				s.Attributes = append(s.Attributes, attrKV{"db.system", attrVal{StringValue: &sys}})
			}
			if q, ok := evt.DB["query"].(string); ok {
				s.Attributes = append(s.Attributes, attrKV{"db.statement", attrVal{StringValue: &q}})
			}
		}

		spans = append(spans, s)
	}

	payload := map[string]any{
		"resourceSpans": []any{
			map[string]any{
				"resource": map[string]any{
					"attributes": []attrKV{
						{"service.name", attrVal{StringValue: ptr("logs-simulator")}},
					},
				},
				"scopeSpans": []any{
					map[string]any{
						"scope": map[string]any{"name": "logs-simulator"},
						"spans": spans,
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt64(failCounter, 1)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(okCounter, 1)
	} else {
		atomic.AddInt64(failCounter, 1)
	}
}

func hexUUID(id string) string {
	result := ""
	for _, c := range id {
		if c != '-' {
			result += string(c)
		}
	}
	for len(result) < 32 {
		result += "0"
	}
	return result
}

func ptr(s string) *string { return &s }

// ─── Search Validator ──────────────────────────────────────────────────────

func runSearch(client *http.Client, searchURL, apiKey, query string, okCounter, failCounter *int64) {
	url := fmt.Sprintf("%s?query=%s", searchURL, query)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt64(failCounter, 1)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		atomic.AddInt64(okCounter, 1)
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			items, _ := result["items"].([]any)
			fmt.Printf("  🔍 Search '%s': %d results\n", query, len(items))
		}
	} else {
		atomic.AddInt64(failCounter, 1)
	}
}

// ─── Env helpers ───────────────────────────────────────────────────────────

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	if n == 0 {
		return 1
	}
	return n
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Second
	}
	return d
}

// ─── Main ──────────────────────────────────────────────────────────────────

func main() {
	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  🎭 Logs Simulator Server — On-Demand Load Tester       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	srv := NewSimulatorServer(cfg)
	if err := srv.Run(ctx); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "❌ Simulator server failed: %v\n", err)
		os.Exit(1)
	}
}
