package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rctiplus/otel-lib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	telemetry      *otellib.Telemetry
	customCounter  metric.Int64Counter
	customGauge    metric.Int64UpDownCounter
	customHistogram metric.Float64Histogram
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	var err error
	telemetry, err = otellib.NewBuilder().
		WithConfig(otellib.DefaultConfig().
			WithServiceName("custom-metrics-service").
			WithServiceVersion("1.0.0").
			WithEnvironment("demo")).
		Build(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer telemetry.Shutdown(ctx)

	// Create custom metrics using the meter from telemetry
	meter := telemetry.Meter()

	customCounter, err = meter.Int64Counter(
		"custom.business.operations.total",
		metric.WithDescription("Total number of business operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		log.Fatalf("Failed to create custom counter: %v", err)
	}

	customGauge, err = meter.Int64UpDownCounter(
		"custom.cache.size",
		metric.WithDescription("Current cache size"),
		metric.WithUnit("{item}"),
	)
	if err != nil {
		log.Fatalf("Failed to create custom gauge: %v", err)
	}

	customHistogram, err = meter.Float64Histogram(
		"custom.processing.duration",
		metric.WithDescription("Duration of custom processing"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		log.Fatalf("Failed to create custom histogram: %v", err)
	}

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/api/process", handleProcess)
	mux.HandleFunc("/api/cache", handleCache)

	handler := telemetry.HTTPMiddleware(mux, "custom-metrics-service")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Custom Metrics Service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	telemetry.SetSpanAttributes(ctx, attribute.String("endpoint", "home"))
	telemetry.RecordMetrics(ctx, "GET", "/", 200)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Custom Metrics Service"))
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := telemetry.StartSpan(ctx, "handleProcess")
	defer span.End()

	start := time.Now()

	// Simulate processing
	time.Sleep(100 * time.Millisecond)

	duration := time.Since(start).Milliseconds()

	// Record custom metrics
	customCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation.type", "data_processing"),
		attribute.String("status", "success"),
	))

	customHistogram.Record(ctx, float64(duration), metric.WithAttributes(
		attribute.String("operation.type", "data_processing"),
	))

	telemetry.SetSpanAttributes(ctx,
		attribute.Int64("processing.duration_ms", duration),
	)

	telemetry.RecordMetrics(ctx, r.Method, "/api/process", 200)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "processed"}`))
}

func handleCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := telemetry.StartSpan(ctx, "handleCache")
	defer span.End()

	// Simulate cache operations
	cacheSize := int64(150)

	// Update gauge
	customGauge.Add(ctx, cacheSize, metric.WithAttributes(
		attribute.String("cache.type", "in-memory"),
	))

	telemetry.SetSpanAttributes(ctx,
		attribute.Int64("cache.size", cacheSize),
	)

	telemetry.RecordMetrics(ctx, r.Method, "/api/cache", 200)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"cache_size": %d}`, cacheSize)))
}

