// Package sharedotel provides a one-call setup for OpenTelemetry SDK
// with an OTLP HTTP exporter targeting the logs-ingest receiver.
//
// Usage in main.go:
//
//	shutdown, err := sharedotel.Init(context.Background(), "my-service")
//	if err != nil { log.Fatal(err) }
//	defer shutdown()
package sharedotel

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Init initialises the global OTEL TracerProvider and returns a shutdown function.
// The exporter target is controlled by OTEL_EXPORTER_OTLP_ENDPOINT env var
// (default: http://logs-ingest:8080).
// The API key is taken from LOGS_API_KEY env var and set as an Authorization header.
func Init(ctx context.Context, serviceName string) (func(), error) {
	endpoint := env("OTEL_EXPORTER_OTLP_ENDPOINT", "http://logs-ingest:8080")
	apiKey := env("LOGS_API_KEY", "")

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(endpoint + "/v1/traces"),
		otlptracehttp.WithInsecure(),
	}
	if apiKey != "" {
		opts = append(opts, otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + apiKey,
		}))
	}

	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(2*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}

	return shutdown, nil
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
