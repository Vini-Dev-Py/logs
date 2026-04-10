package otelreceiver_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"logs-ingest/internal/app/otelreceiver"
)

const validPayload = `{
  "resourceSpans": [{
    "resource": {
      "attributes": [{"key": "service.name", "value": {"stringValue": "my-service"}}]
    },
    "scopeSpans": [{
      "scope": {"name": "test-lib"},
      "spans": [{
        "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
        "spanId":  "00f067aa0ba902b7",
        "parentSpanId": "",
        "name": "GET /users",
        "kind": 2,
        "startTimeUnixNano": "1741600000000000000",
        "endTimeUnixNano":   "1741600000120000000",
        "attributes": [
          {"key": "http.method", "value": {"stringValue": "GET"}},
          {"key": "http.target", "value": {"stringValue": "/users"}}
        ],
        "status": {"code": 1, "message": "OK"}
      }]
    }]
  }]
}`

func makeRequest(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/v1/traces", strings.NewReader(body))
}

func TestParseOTLPRequest_ValidPayload(t *testing.T) {
	r := makeRequest(validPayload)
	events, err := otelreceiver.ParseOTLPRequest(r, "company-abc")

	require.NoError(t, err)
	require.Len(t, events, 1)

	evt := events[0]
	assert.Equal(t, "company-abc", evt.CompanyID)
	assert.Equal(t, "my-service", evt.ServiceName)
	assert.Equal(t, "HTTP", evt.Operation["type"])
	assert.Equal(t, "GET /users", evt.Operation["name"])
	assert.NotNil(t, evt.HTTP)
	assert.Nil(t, evt.ParentNodeID) // parentSpanId is empty
}

func TestParseOTLPRequest_EmptySpans(t *testing.T) {
	payload := `{"resourceSpans": []}`
	r := makeRequest(payload)
	events, err := otelreceiver.ParseOTLPRequest(r, "company-abc")

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestParseOTLPRequest_MalformedJSON(t *testing.T) {
	r := makeRequest(`{not valid json}`)
	events, err := otelreceiver.ParseOTLPRequest(r, "company-abc")

	assert.Error(t, err)
	assert.Nil(t, events)
}

func TestParseOTLPRequest_DBSpan(t *testing.T) {
	payload := `{
    "resourceSpans": [{
      "resource": {"attributes": [{"key": "service.name", "value": {"stringValue": "db-service"}}]},
      "scopeSpans": [{
        "spans": [{
          "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
          "spanId": "00f067aa0ba902b8",
          "name": "SELECT users",
          "startTimeUnixNano": "1741600000000000000",
          "endTimeUnixNano": "1741600000005000000",
          "attributes": [
            {"key": "db.system", "value": {"stringValue": "postgres"}},
            {"key": "db.statement", "value": {"stringValue": "SELECT * FROM users"}}
          ],
          "status": {"code": 2}
        }]
      }]
    }]
  }`

	r := makeRequest(payload)
	events, err := otelreceiver.ParseOTLPRequest(r, "company-db")
	require.NoError(t, err)
	require.Len(t, events, 1)

	evt := events[0]
	assert.Equal(t, "DB", evt.Operation["type"])
	assert.Equal(t, "ERROR", evt.Operation["status"])
	assert.Equal(t, "postgres", evt.DB["system"])
}
