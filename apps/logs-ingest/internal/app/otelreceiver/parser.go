package otelreceiver

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"logs-ingest/internal/app"

	"github.com/google/uuid"
)

// OTLPSpan represents a minimal OTLP span in JSON format
// Compatible with go.opentelemetry.io/proto/otlp JSON encoding
type otlpRequest struct {
	ResourceSpans []resourceSpan `json:"resourceSpans"`
}

type resourceSpan struct {
	Resource    resource    `json:"resource"`
	ScopeSpans []scopeSpan `json:"scopeSpans"`
}

type resource struct {
	Attributes []kvAttribute `json:"attributes"`
}

type scopeSpan struct {
	Scope       instrScope `json:"scope"`
	Spans       []span     `json:"spans"`
}

type instrScope struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type span struct {
	TraceID           string        `json:"traceId"`
	SpanID            string        `json:"spanId"`
	ParentSpanID      string        `json:"parentSpanId"`
	Name              string        `json:"name"`
	Kind              int           `json:"kind"`
	StartTimeUnixNano string        `json:"startTimeUnixNano"`
	EndTimeUnixNano   string        `json:"endTimeUnixNano"`
	Attributes        []kvAttribute `json:"attributes"`
	Status            spanStatus    `json:"status"`
}

type spanStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type kvAttribute struct {
	Key   string         `json:"key"`
	Value map[string]any `json:"value"`
}

// ParseOTLPRequest converts a raw OTLP HTTP/JSON request body into a list of LogEvents
func ParseOTLPRequest(r *http.Request, companyID string) ([]app.LogEvent, error) {
	var req otlpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode OTLP request: %w", err)
	}

	var events []app.LogEvent
	for _, rs := range req.ResourceSpans {
		serviceName := attrString(rs.Resource.Attributes, "service.name")
		if serviceName == "" {
			serviceName = "unknown"
		}

		for _, ss := range rs.ScopeSpans {
			for _, s := range ss.Spans {
				traceID := hexToUUID(s.TraceID)
				nodeID := hexToUUID(s.SpanID)

				status := "OK"
				if s.Status.Code == 2 { // STATUS_CODE_ERROR in OTLP
					status = "ERROR"
				}

				startAt := nanoToTime(s.StartTimeUnixNano)
				endAt := nanoToTime(s.EndTimeUnixNano)
				durationMs := int(endAt.Sub(startAt).Milliseconds())

				httpMethod := attrString(s.Attributes, "http.method")
				httpPath := attrString(s.Attributes, "http.target")
				if httpPath == "" {
					httpPath = attrString(s.Attributes, "http.url")
				}
				httpStatusStr := attrString(s.Attributes, "http.status_code")
				dbSystem := attrString(s.Attributes, "db.system")
				dbQuery := attrString(s.Attributes, "db.statement")

				opType := spanKindToType(s.Kind, httpMethod, dbSystem)

				evt := app.LogEvent{
					EventID:     uuid.NewString(),
					TraceID:     traceID,
					NodeID:      nodeID,
					ServiceName: serviceName,
					CompanyID:   companyID,
					Operation: map[string]any{
						"type":       opType,
						"name":       s.Name,
						"status":     status,
						"startAt":    startAt.Format(time.RFC3339Nano),
						"endAt":      endAt.Format(time.RFC3339Nano),
						"durationMs": durationMs,
					},
				}

				if s.ParentSpanID != "" {
					parent := hexToUUID(s.ParentSpanID)
					evt.ParentNodeID = &parent
				}

				if httpMethod != "" {
					evt.HTTP = map[string]any{
						"method":     httpMethod,
						"path":       httpPath,
						"statusCode": httpStatusStr,
					}
				}

				if dbSystem != "" {
					evt.DB = map[string]any{
						"system": dbSystem,
						"query":  dbQuery,
					}
				}

				events = append(events, evt)
			}
		}
	}

	return events, nil
}

func attrString(attrs []kvAttribute, key string) string {
	for _, a := range attrs {
		if a.Key == key {
			if sv, ok := a.Value["stringValue"].(string); ok {
				return sv
			}
			if iv, ok := a.Value["intValue"].(float64); ok {
				return fmt.Sprintf("%d", int(iv))
			}
		}
	}
	return ""
}

func hexToUUID(h string) string {
	if len(h) >= 32 {
		b, err := hex.DecodeString(h)
		if err == nil && len(b) >= 16 {
			return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
		}
	}
	if h == "" {
		return uuid.NewString()
	}
	return h
}

func nanoToTime(nanoStr string) time.Time {
	var ns int64
	fmt.Sscanf(nanoStr, "%d", &ns)
	if ns == 0 {
		return time.Now()
	}
	return time.Unix(0, ns)
}

func spanKindToType(kind int, httpMethod, dbSystem string) string {
	if dbSystem != "" {
		return "DB"
	}
	if httpMethod != "" {
		return "HTTP"
	}
	switch kind {
	case 3: // CLIENT
		return "HTTP"
	case 4: // PRODUCER
		return "EXTERNAL_API"
	default:
		return "CUSTOM"
	}
}
