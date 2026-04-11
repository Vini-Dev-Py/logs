package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ─── Config ────────────────────────────────────────────────────────────────

type Config struct {
	TargetURL    string // e.g. http://localhost/ingest/v1/log-events
	OTLPURL      string // e.g. http://localhost/v1/traces
	APIKey       string
	Mode         string // "native" or "otlp"
	Workers      int
	BurstSize    int // total traces per burst
	Bursts       int // number of bursts
	BurstDelay   time.Duration
	SearchURL    string // e.g. http://localhost/api/search
	SearchQuery  string // auto-search after bursts
	CompanyID    string
}

func loadConfig() Config {
	cfg := Config{
		TargetURL:  env("TARGET_URL", "http://localhost/ingest/v1/log-events"),
		OTLPURL:    env("OTLP_URL", "http://localhost/v1/traces"),
		APIKey:     env("API_KEY", "logs_dev_api_key"),
		Mode:       env("MODE", "native"),
		Workers:    parseInt(env("WORKERS", "10")),
		BurstSize:  parseInt(env("BURST_SIZE", "50")),
		Bursts:     parseInt(env("BURSTS", "20")),
		BurstDelay: parseDuration(env("BURST_DELAY", "2s")),
		SearchURL:  env("SEARCH_URL", "http://localhost/api/search"),
		SearchQuery: env("SEARCH_QUERY", "SELECT"),
		CompanyID:  env("COMPANY_ID", "de696723-dd06-4315-9735-ee9cf156623b"),
	}
	if cfg.Mode == "" {
		cfg.Mode = "native"
	}
	return cfg
}

// ─── Counters ──────────────────────────────────────────────────────────────

var (
	totalSent     int64
	totalOK       int64
	totalFail     int64
	totalSearchOK int64
	totalSearchFail int64
)

// ─── Service Scenarios ─────────────────────────────────────────────────────

var scenarios = []ScenarioFn{
	scenarioECommerce,
	scenarioPaymentProcessing,
	scenarioUserAuthentication,
	scenarioReportGeneration,
	scenarioWebhookDelivery,
	scenarioDataSync,
}

type ScenarioFn func(traceID string) []LogEvent

type LogEvent struct {
	TraceID      string  `json:"traceId"`
	NodeID       string  `json:"nodeId"`
	ParentNodeID *string `json:"parentNodeId,omitempty"`
	ServiceName  string  `json:"serviceName"`
	CompanyID    string  `json:"companyId"`
	EventID      string  `json:"eventId,omitempty"`
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

// ─── Realistic Scenarios ───────────────────────────────────────────────────

func scenarioECommerce(traceID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	// API Gateway
	node1 := makeNode(traceID, "", "api-gateway", "HTTP", "GET /api/products", "OK", now, 12)
	events = append(events, node1)

	// Auth check
	node2 := makeNode(traceID, node1.NodeID, "api-gateway", "HTTP", "POST /auth/verify", "OK", now.Add(13*time.Millisecond), 8)
	events = append(events, node2)

	// Product Service
	node3 := makeNode(traceID, node1.NodeID, "product-service", "HTTP", "GET /products", "OK", now.Add(25*time.Millisecond), 45)
	events = append(events, node3)

	// DB Query - products
	node4 := makeNodeWithDB(traceID, node3.NodeID, "product-service", "DB", "SELECT products", "OK", now.Add(30*time.Millisecond), 20,
		"postgres", "SELECT * FROM products WHERE category = 'electronics' AND active = true ORDER BY created_at DESC LIMIT 50", 48)
	events = append(events, node4)

	// Cache check
	node5 := makeNode(traceID, node1.NodeID, "product-service", "HTTP", "GET redis://cache", "OK", now.Add(55*time.Millisecond), 3)
	events = append(events, node5)

	// Recommendation Service
	node6 := makeNode(traceID, node1.NodeID, "recommendation-svc", "HTTP", "POST /recommend", "OK", now.Add(60*time.Millisecond), 120)
	events = append(events, node6)

	// ML inference
	node7 := makeNode(traceID, node6.NodeID, "ml-engine", "CUSTOM", "predict(user_profile)", "OK", now.Add(65*time.Millisecond), 85)
	events = append(events, node7)

	return events
}

func scenarioPaymentProcessing(traceID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	// Checkout API
	node1 := makeNode(traceID, "", "checkout-api", "HTTP", "POST /checkout", "OK", now, 250)

	// Error injection (5% chance)
	status := "OK"
	if rand.Float64() < 0.05 {
		status = "ERROR"
	}

	events = append(events, node1)

	// Order Service
	node2 := makeNode(traceID, node1.NodeID, "order-service", "HTTP", "POST /orders", status, now.Add(15*time.Millisecond), 180)
	events = append(events, node2)

	// DB - insert order
	node3 := makeNodeWithDB(traceID, node2.NodeID, "order-service", "DB", "INSERT orders", status, now.Add(20*time.Millisecond), 25,
		"postgres", fmt.Sprintf("INSERT INTO orders (id, user_id, total, status) VALUES ('%s', 42, 299.90, 'pending')", uuid.New().String()[:8]), 1)
	events = append(events, node3)

	// Payment Gateway
	node4 := makeNode(traceID, node1.NodeID, "payment-gateway", "EXTERNAL_API", "POST https://api.stripe.com/v1/charges", status, now.Add(30*time.Millisecond), 350)
	events = append(events, node4)

	if status == "OK" {
		// Queue notification
		node5 := makeNode(traceID, node1.NodeID, "notification-svc", "HTTP", "POST rabbitmq://queue/email", "OK", now.Add(400*time.Millisecond), 5)
		events = append(events, node5)
	} else {
		// Error path
		node5 := makeNode(traceID, node4.NodeID, "payment-gateway", "HTTP", "POST /rollback", "OK", now.Add(380*time.Millisecond), 15)
		events = append(events, node5)

		node6 := makeNodeWithDB(traceID, node2.NodeID, "order-service", "DB", "UPDATE orders", "OK", now.Add(400*time.Millisecond), 10,
			"postgres", fmt.Sprintf("UPDATE orders SET status = 'failed' WHERE id = '%s'", uuid.New().String()[:8]), 1)
		events = append(events, node6)
	}

	return events
}

func scenarioUserAuthentication(traceID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	// Login endpoint
	node1 := makeNode(traceID, "", "auth-service", "HTTP", "POST /auth/login", "OK", now, 85)
	events = append(events, node1)

	// DB lookup
	node2 := makeNodeWithDB(traceID, node1.NodeID, "auth-service", "DB", "SELECT users", "OK", now.Add(5*time.Millisecond), 15,
		"postgres", "SELECT id, email, password_hash, role FROM users WHERE email = 'user@example.com' LIMIT 1", 1)
	events = append(events, node2)

	// JWT generation
	node3 := makeNode(traceID, node1.NodeID, "auth-service", "CUSTOM", "generate_jwt", "OK", now.Add(25*time.Millisecond), 3)
	events = append(events, node3)

	// Session store
	node4 := makeNode(traceID, node1.NodeID, "session-store", "HTTP", "SET redis://session", "OK", now.Add(30*time.Millisecond), 8)
	events = append(events, node4)

	// Audit log
	node5 := makeNode(traceID, node1.NodeID, "audit-service", "HTTP", "POST kafka://audit", "OK", now.Add(40*time.Millisecond), 12)
	events = append(events, node5)

	return events
}

func scenarioReportGeneration(traceID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	// Report API
	node1 := makeNode(traceID, "", "report-api", "HTTP", "GET /reports/monthly", "OK", now, 2500)
	events = append(events, node1)

	// DB - heavy query
	node2 := makeNodeWithDB(traceID, node1.NodeID, "report-api", "DB", "SELECT aggregations", "OK", now.Add(10*time.Millisecond), 1800,
		"postgres", `SELECT DATE_TRUNC('month', created_at) AS month, COUNT(*) AS total, SUM(amount) AS revenue 
FROM transactions 
WHERE created_at >= NOW() - INTERVAL '30 days' 
GROUP BY month ORDER BY month DESC`)
	events = append(events, node2)

	// Cache miss
	node3 := makeNode(traceID, node1.NodeID, "report-api", "HTTP", "GET redis://cache", "ERROR", now.Add(15*time.Millisecond), 2)
	events = append(events, node3)

	// PDF generation
	node4 := makeNode(traceID, node1.NodeID, "pdf-generator", "CUSTOM", "render_template", "OK", now.Add(1850*time.Millisecond), 400)
	events = append(events, node4)

	// Upload to S3
	node5 := makeNode(traceID, node1.NodeID, "storage-service", "EXTERNAL_API", "PUT https://s3.amazonaws.com/reports", "OK", now.Add(2300*time.Millisecond), 150)
	events = append(events, node5)

	return events
}

func scenarioWebhookDelivery(traceID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	// Webhook dispatcher
	node1 := makeNode(traceID, "", "webhook-dispatcher", "HTTP", "POST /dispatch", "OK", now, 5)
	events = append(events, node1)

	// Queue
	node2 := makeNode(traceID, node1.NodeID, "webhook-dispatcher", "HTTP", "POST rabbitmq://webhooks", "OK", now.Add(2*time.Millisecond), 3)
	events = append(events, node2)

	// Worker consume
	node3 := makeNode(traceID, node2.NodeID, "webhook-worker", "HTTP", "GET rabbitmq://webhooks/consume", "OK", now.Add(50*time.Millisecond), 200)
	events = append(events, node3)

	// External delivery - success or failure
	status := "OK"
	if rand.Float64() < 0.1 {
		status = "ERROR"
	}

	node4 := makeNode(traceID, node3.NodeID, "webhook-worker", "EXTERNAL_API", "POST https://customer.example.com/hook", status, now.Add(60*time.Millisecond), 450)
	events = append(events, node4)

	if status == "OK" {
		node5 := makeNodeWithDB(traceID, node3.NodeID, "webhook-worker", "DB", "UPDATE deliveries", "OK", now.Add(520*time.Millisecond), 8,
			"postgres", fmt.Sprintf("UPDATE webhook_deliveries SET status = 'delivered', delivered_at = NOW() WHERE id = '%s'", uuid.New().String()[:8]), 1)
		events = append(events, node5)
	}

	return events
}

func scenarioDataSync(traceID string) []LogEvent {
	var events []LogEvent
	now := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	// Sync trigger
	node1 := makeNode(traceID, "", "sync-service", "CUSTOM", "scheduled_sync", "OK", now, 5000)
	events = append(events, node1)

	// DB - fetch changes
	node2 := makeNodeWithDB(traceID, node1.NodeID, "sync-service", "DB", "SELECT changes", "OK", now.Add(10*time.Millisecond), 2000,
		"postgres", "SELECT * FROM cdc_log WHERE synced = false ORDER BY created_at LIMIT 1000")
	events = append(events, node2)

	// Transform
	node3 := makeNode(traceID, node1.NodeID, "transform-engine", "CUSTOM", "map_fields", "OK", now.Add(2050*time.Millisecond), 150)
	events = append(events, node3)

	// External API - push
	node4 := makeNode(traceID, node3.NodeID, "sync-service", "EXTERNAL_API", "POST https://partner-api.example.com/sync", "OK", now.Add(2200*time.Millisecond), 1200)
	events = append(events, node4)

	// DB - mark synced
	node5 := makeNodeWithDB(traceID, node1.NodeID, "sync-service", "DB", "UPDATE cdc_log", "OK", now.Add(3450*time.Millisecond), 50,
		"postgres", "UPDATE cdc_log SET synced = true, synced_at = NOW() WHERE id IN (SELECT id FROM cdc_log WHERE synced = false LIMIT 1000)")
	events = append(events, node5)

	return events
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func makeNode(traceID, parentNodeID, service, opType, name, status string, start time.Time, durationMs int) LogEvent {
	evt := LogEvent{
		TraceID:     traceID,
		NodeID:      uuid.New().String(),
		ServiceName: service,
		CompanyID:   "default-company",
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

	// HTTP metadata
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

func makeNodeWithDB(traceID, parentNodeID, service, opType, name, status string, start time.Time, durationMs int, dbSystem, dbQuery string, dbRows ...int) LogEvent {
	evt := makeNode(traceID, parentNodeID, service, opType, name, status, start, durationMs)
	evt.DB = map[string]any{
		"system": dbSystem,
		"query":  dbQuery,
	}
	if len(dbRows) > 0 {
		evt.DB["rows"] = dbRows[0]
	}
	return evt
}

// ─── Native Sender ─────────────────────────────────────────────────────────

func sendNative(client *http.Client, url, apiKey string, evt LogEvent) {
	body := map[string]any{
		"traceId":      evt.TraceID,
		"nodeId":       evt.NodeID,
		"parentNodeId": evt.ParentNodeID,
		"serviceName":  evt.ServiceName,
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
		atomic.AddInt64(&totalFail, 1)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(&totalOK, 1)
	} else {
		atomic.AddInt64(&totalFail, 1)
	}
}

// ─── OTLP Sender ───────────────────────────────────────────────────────────

func sendOTLPBatch(client *http.Client, url, apiKey string, events []LogEvent) {
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

		// Status
		if evt.Operation.Status == "ERROR" {
			s.Status.Code = 2
		} else {
			s.Status.Code = 1
		}

		// Kind mapping
		switch evt.Operation.Type {
		case "HTTP":
			s.Kind = 2 // SERVER
		case "EXTERNAL_API":
			s.Kind = 3 // CLIENT
		case "DB":
			s.Kind = 2
		default:
			s.Kind = 1 // INTERNAL
		}

		// Attributes
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
		atomic.AddInt64(&totalFail, 1)
		fmt.Printf("  ❌ OTLP send error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(&totalOK, 1)
	} else {
		atomic.AddInt64(&totalFail, 1)
	}
}

func hexUUID(id string) string {
	// Simple conversion: remove dashes for OTLP hex format
	result := ""
	for _, c := range id {
		if c != '-' {
			result += string(c)
		}
	}
	// Pad to 32 chars
	for len(result) < 32 {
		result += "0"
	}
	return result
}

func ptr(s string) *string { return &s }

// ─── Search Validator ──────────────────────────────────────────────────────

func runSearch(client *http.Client, searchURL, apiKey, query string) {
	url := fmt.Sprintf("%s?query=%s", searchURL, query)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt64(&totalSearchFail, 1)
		fmt.Printf("  ❌ Search error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		atomic.AddInt64(&totalSearchOK, 1)
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			items, _ := result["items"].([]any)
			fmt.Printf("  🔍 Search '%s': %d results\n", query, len(items))
		}
	} else {
		atomic.AddInt64(&totalSearchFail, 1)
		fmt.Printf("  ❌ Search '%s': HTTP %d\n", query, resp.StatusCode)
	}
}

// ─── Main ──────────────────────────────────────────────────────────────────

func main() {
	cfg := loadConfig()
	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  🚀 Logs Simulator — Load & Scalability Tester          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("\n")
	fmt.Printf("  Mode:       %s\n", cfg.Mode)
	fmt.Printf("  Target:     %s\n", cfg.TargetURL)
	fmt.Printf("  Workers:    %d\n", cfg.Workers)
	fmt.Printf("  Burst Size: %d traces/burst\n", cfg.BurstSize)
	fmt.Printf("  Bursts:     %d\n", cfg.Bursts)
	fmt.Printf("  Burst Delay:%v\n", cfg.BurstDelay)
	fmt.Printf("  Total:      ~%d traces expected\n\n", cfg.BurstSize*cfg.Bursts)

	// Wait for target to be ready
	fmt.Print("⏳ Waiting for target to be ready... ")
	ready := false
	for i := 0; i < 30; i++ {
		payload := []byte(`{"traceId":"hc","nodeId":"hc","serviceName":"hc","operation":{"type":"HTTP","name":"HC","status":"OK","startAt":"2026-01-01T00:00:00Z","endAt":"2026-01-01T00:00:01Z","durationMs":1}}`)
		req, _ := http.NewRequest("POST", cfg.TargetURL, bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
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
		os.Exit(1)
	}
	fmt.Println("✅ Ready!")
	fmt.Println()

	startTime := time.Now()

	for burst := 0; burst < cfg.Bursts; burst++ {
		burstStart := time.Now()
		fmt.Printf("━━━ Burst %d/%d ━━━\n", burst+1, cfg.Bursts)

		// Generate traces
		traceIDs := make([]string, cfg.BurstSize)
		allEvents := make([][]LogEvent, cfg.BurstSize)
		for i := 0; i < cfg.BurstSize; i++ {
			traceIDs[i] = uuid.New().String()
			scenarioFn := scenarios[rand.Intn(len(scenarios))]
			allEvents[i] = scenarioFn(traceIDs[i])
		}

		// Send via workers
		var wg sync.WaitGroup
		sem := make(chan struct{}, cfg.Workers)

		for i := 0; i < cfg.BurstSize; i++ {
			wg.Add(1)
			sem <- struct{}{}
			go func(idx int) {
				defer wg.Done()
				defer func() { <-sem }()

				events := allEvents[idx]
				atomic.AddInt64(&totalSent, int64(len(events)))

				if cfg.Mode == "otlp" {
					sendOTLPBatch(client, cfg.OTLPURL, cfg.APIKey, events)
				} else {
					for _, evt := range events {
						sendNative(client, cfg.TargetURL, cfg.APIKey, evt)
					}
				}
			}(i)
		}

		wg.Wait()
		burstDuration := time.Since(burstStart)
		fmt.Printf("  ✅ Burst complete in %v (%d events sent)\n\n", burstDuration.Round(time.Millisecond), atomic.LoadInt64(&totalSent))

		// Run search every 5 bursts
		if (burst+1)%5 == 0 && cfg.SearchQuery != "" {
			fmt.Println("🔍 Running search validation...")
			runSearch(client, cfg.SearchURL, cfg.APIKey, cfg.SearchQuery)
			runSearch(client, cfg.SearchURL, cfg.APIKey, "INSERT")
			runSearch(client, cfg.SearchURL, cfg.APIKey, "SELECT")
			fmt.Println()
		}

		if burst < cfg.Bursts-1 {
			time.Sleep(cfg.BurstDelay)
		}
	}

	// Final report
	elapsed := time.Since(startTime)
	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  📊 Final Report                                         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("\n")
	fmt.Printf("  Total Events Sent:    %d\n", atomic.LoadInt64(&totalSent))
	fmt.Printf("  Successful (2xx):     %d\n", atomic.LoadInt64(&totalOK))
	fmt.Printf("  Failed:               %d\n", atomic.LoadInt64(&totalFail))
	fmt.Printf("  Search Queries OK:    %d\n", atomic.LoadInt64(&totalSearchOK))
	fmt.Printf("  Search Queries Fail:  %d\n", atomic.LoadInt64(&totalSearchFail))
	fmt.Printf("  Total Time:           %v\n", elapsed.Round(time.Millisecond))

	sent := float64(atomic.LoadInt64(&totalSent))
	if elapsed.Seconds() > 0 {
		fmt.Printf("  Throughput:           %.0f events/sec\n", sent/elapsed.Seconds())
	}

	if atomic.LoadInt64(&totalSent) > 0 {
		successRate := float64(atomic.LoadInt64(&totalOK)) / float64(atomic.LoadInt64(&totalSent)) * 100
		fmt.Printf("  Success Rate:         %.1f%%\n", successRate)
	}

	fmt.Println()
	if atomic.LoadInt64(&totalFail) == 0 {
		fmt.Println("🎉 All events ingested successfully!")
	} else {
		fmt.Printf("⚠️  %d events failed — check logs for details\n", atomic.LoadInt64(&totalFail))
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
