package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"logs-simulator/internal/generator"
	"logs-simulator/internal/sender"
)

func main() {
	// ── Flags ──
	total := flag.Int("n", 500, "Total number of traces to generate")
	rate := flag.Int("rate", 50, "Requests per second (0 = burst)")
	workers := flag.Int("workers", 10, "Number of concurrent workers")
	apiKey := flag.String("key", "logs_dev_api_key", "API Key for authentication")
	baseURL := flag.String("url", "http://localhost", "Base URL of the Logs platform")
	mode := flag.String("mode", "native", "Ingestion mode: native or otlp")
	search := flag.Bool("search", false, "Run search validation after ingestion")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║           🚀  Logs Simulator — Load Generator            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n⚡  Received shutdown signal...")
		cancel()
	}()

	cfg := sender.Config{
		BaseURL: *baseURL,
		APIKey:  *apiKey,
		Mode:    *mode,
		Workers: *workers,
		Rate:    *rate,
		Verbose: *verbose,
	}

	gen := generator.New()
	s := sender.New(cfg)

	start := time.Now()
	fmt.Printf("📊 Generating %d traces with %d workers...\n", *total, *workers)
	fmt.Printf("   Mode: %s | Rate: %d req/s | Base URL: %s\n\n", *mode, *rate, *baseURL)

	// ── Run ingestion ──
	stats, err := s.Run(ctx, gen, *total)
	if err != nil {
		log.Fatalf("❌  Simulator failed: %v", err)
	}

	elapsed := time.Since(start)

	// ── Print report ──
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                   📋  Ingestion Report                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  ⏱️   Total time:       %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  📤  Total traces:     %d\n", stats.Total)
	fmt.Printf("  ✅  Success:          %d\n", stats.Success)
	fmt.Printf("  ❌  Failed:           %d\n", stats.Failed)
	fmt.Printf("  ⚡  Throughput:       %.0f traces/sec\n", stats.Throughput())
	fmt.Printf("  📏  Avg latency:      %v\n", stats.AvgLatency().Round(time.Millisecond))
	fmt.Printf("  🔝  P99 latency:      %v\n", stats.P99Latency().Round(time.Millisecond))
	fmt.Printf("  🧵  Workers:          %d\n", *workers)
	fmt.Printf("  💾  Goroutines:       %d\n", runtime.NumGoroutine())
	fmt.Println()

	// ── Search validation ──
	if *search && stats.Success > 0 {
		fmt.Println("🔍  Running search validation...")
		time.Sleep(2 * time.Second) // wait for OpenSearch indexing

		searchClient := sender.NewSearchClient(cfg)
		terms := []string{"SELECT", "INSERT", "POST /api", "ERROR", "payment", "timeout"}

		for _, term := range terms {
			count, err := searchClient.Search(ctx, term)
			if err != nil {
				fmt.Printf("   ❌  Search '%s': error — %v\n", term, err)
			} else {
				fmt.Printf("   ✅  Search '%s': %d results\n", term, count)
			}
		}
		fmt.Println()
	}

	fmt.Println("✅  Simulator finished!")
}
