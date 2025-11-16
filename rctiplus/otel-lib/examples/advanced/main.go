package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rctiplus/otel-lib"
	"go.opentelemetry.io/otel/attribute"
)

var telemetry *otellib.Telemetry

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	ctx := context.Background()

	// Advanced configuration dengan custom attributes dan settings
	config := otellib.DefaultConfig().
		WithServiceName("advanced-example-service").
		WithServiceVersion("2.0.0").
		WithEnvironment("production").
		WithAttributes(
			attribute.String("team", "backend"),
			attribute.String("region", "us-east-1"),
			attribute.String("deployment", "kubernetes"),
		).
		WithMetricInterval(10 * time.Second).
		WithHistogramBoundaries([]float64{0, 10, 50, 100, 250, 500, 1000, 2500, 5000}).
		WithPropagators("tracecontext", "baggage")

	var err error
	telemetry, err = otellib.NewBuilder().
		WithConfig(config).
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

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/api/users", handleUsers)
	mux.HandleFunc("/api/users/", handleUserByID)
	mux.HandleFunc("/api/external", handleExternalCall)

	handler := telemetry.HTTPMiddleware(mux, "advanced-example-service")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Advanced Example Service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	telemetry.SetSpanAttributes(ctx, attribute.String("endpoint", "home"))
	telemetry.RecordMetrics(ctx, "GET", "/", 200)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Advanced Example Service"})
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := telemetry.StartSpan(ctx, "handleUsers")
	defer span.End()

	telemetry.SetSpanAttributes(ctx,
		attribute.String("method", r.Method),
		attribute.String("endpoint", "users"),
	)

	// Simulate database query
	users := []User{
		{ID: 1, Name: "John Doe"},
		{ID: 2, Name: "Jane Smith"},
		{ID: 3, Name: "Bob Wilson"},
	}

	telemetry.SetSpanAttributes(ctx,
		attribute.Int("users.count", len(users)),
	)

	telemetry.RecordMetrics(ctx, r.Method, "/api/users", 200)
	respondJSON(w, http.StatusOK, map[string]interface{}{"users": users})
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := telemetry.StartSpan(ctx, "handleUserByID")
	defer span.End()

	// Extract ID from URL (simplified)
	userID := 1 // In real app, extract from URL
	telemetry.SetSpanAttributes(ctx,
		attribute.Int("user.id", userID),
		attribute.String("endpoint", "user_by_id"),
	)

	// Simulate database query
	time.Sleep(20 * time.Millisecond)

	user := User{ID: userID, Name: "John Doe"}
	telemetry.SetSpanAttributes(ctx, attribute.Bool("user.found", true))

	telemetry.RecordMetrics(ctx, r.Method, "/api/users/{id}", 200)
	respondJSON(w, http.StatusOK, user)
}

func handleExternalCall(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := telemetry.StartSpan(ctx, "handleExternalCall")
	defer span.End()

	telemetry.SetSpanAttributes(ctx,
		attribute.String("endpoint", "external"),
		attribute.String("operation", "call_external_api"),
	)

	// Call external service using instrumented HTTP client
	err := callExternalService(ctx, "https://httpbin.org/delay/1")
	if err != nil {
		telemetry.RecordSpanError(ctx, err)
		telemetry.RecordMetrics(ctx, r.Method, "/api/external", 500)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	telemetry.RecordMetrics(ctx, r.Method, "/api/external", 200)
	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func callExternalService(ctx context.Context, url string) error {
	ctx, span := telemetry.StartSpan(ctx, "callExternalService")
	defer span.End()

	telemetry.SetSpanAttributes(ctx,
		attribute.String("http.url", url),
		attribute.String("http.method", "GET"),
	)

	// Use instrumented HTTP client from library
	client := telemetry.HTTPClient()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		telemetry.RecordSpanError(ctx, err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		telemetry.RecordSpanError(ctx, err)
		return err
	}
	defer resp.Body.Close()

	telemetry.SetSpanAttributes(ctx,
		attribute.Int("http.status_code", resp.StatusCode),
		attribute.String("peer.service", "httpbin.org"),
	)

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("external service returned status: %d", resp.StatusCode)
		telemetry.RecordSpanError(ctx, err)
		return err
	}

	return nil
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

