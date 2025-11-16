package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rctiplus/otel-lib"
	"go.opentelemetry.io/otel/attribute"
)

var telemetry *otellib.Telemetry

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry dengan basic configuration
	var err error
	telemetry, err = otellib.NewBuilder().
		WithConfig(otellib.DefaultConfig().
			WithServiceName("example-service").
			WithServiceVersion("1.0.0").
			WithEnvironment("development")).
		Build(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/users", handleUsers)

	// Wrap dengan OpenTelemetry middleware
	handler := telemetry.HTTPMiddleware(mux, "example-service")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Example Service starting on port %s", port)
	log.Printf("📊 OpenTelemetry enabled - Sending to %s", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))

	// Graceful shutdown
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set span attributes
	telemetry.SetSpanAttributes(ctx,
		attribute.String("endpoint", "home"),
		attribute.String("handler", "handleHome"),
	)

	// Record metrics
	telemetry.RecordMetrics(ctx, "GET", "/", 200)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Welcome to Example Service"}`))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	telemetry.SetSpanAttributes(ctx,
		attribute.String("endpoint", "health"),
		attribute.String("handler", "handleHealth"),
	)

	telemetry.RecordMetrics(ctx, "GET", "/health", 200)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Start a new span
	ctx, span := telemetry.StartSpan(ctx, "handleUsers")
	defer span.End()

	// Set attributes
	telemetry.SetSpanAttributes(ctx,
		attribute.String("endpoint", "users"),
		attribute.String("method", r.Method),
		attribute.String("handler", "handleUsers"),
	)

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	// Record metrics
	telemetry.RecordMetrics(ctx, r.Method, "/api/users", 200)

	// Set additional span attributes
	telemetry.SetSpanAttributes(ctx,
		attribute.Int("users.count", 3),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]}`))
}

