package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	otellib "github.com/diki-haryadi/otel-lib"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	telemetry *otellib.Telemetry
)

func initTelemetry(ctx context.Context) (*otellib.Telemetry, error) {
	config := otellib.DefaultConfig().
		WithServiceName("frontend").
		WithServiceVersion("1.0.0").
		WithEnvironment("production")

	return otellib.NewBuilder().
		WithConfig(config).
		Build(ctx)
}

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.route", "/"))

	start := time.Now()
	defer func() {
		duration := float64(time.Since(start).Milliseconds())
		telemetry.RequestDuration().Record(ctx, duration)
		telemetry.RecordMetrics(ctx, r.Method, "/", http.StatusOK)
	}()

	log.Printf("Frontend: Handling home request")

	response := map[string]string{
		"service": "frontend",
		"message": "Welcome to the Alloy Demo!",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.route", "/products"))

	start := time.Now()
	statusCode := http.StatusOK
	defer func() {
		duration := float64(time.Since(start).Milliseconds())
		telemetry.RequestDuration().Record(ctx, duration)
		telemetry.RecordMetrics(ctx, r.Method, "/products", statusCode)
	}()

	log.Printf("Frontend: Fetching products from backend")

	// Call backend service
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://backend:8081"
	}

	client := telemetry.HTTPClient()
	req, err := http.NewRequestWithContext(ctx, "GET", backendURL+"/api/products", nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		statusCode = http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		statusCode = http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		statusCode = http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}

	statusCode = resp.StatusCode
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.route", "/checkout"))

	start := time.Now()
	statusCode := http.StatusOK
	defer func() {
		duration := float64(time.Since(start).Milliseconds())
		telemetry.RequestDuration().Record(ctx, duration)
		telemetry.RecordMetrics(ctx, r.Method, "/checkout", statusCode)
	}()

	log.Printf("Frontend: Processing checkout")

	// Call payment service
	paymentURL := os.Getenv("PAYMENT_URL")
	if paymentURL == "" {
		paymentURL = "http://payment:8082"
	}

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequestWithContext(ctx, "POST", paymentURL+"/api/payment", nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		statusCode = http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		statusCode = http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		statusCode = http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
		return
	}

	statusCode = resp.StatusCode
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}

func main() {
	ctx := context.Background()

	// Initialize telemetry
	var err error
	telemetry, err = initTelemetry(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		if err := telemetry.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down telemetry: %v", err)
		}
	}()

	// Setup HTTP handlers with instrumentation
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/products", handleProducts)
	mux.HandleFunc("/checkout", handleCheckout)

	handler := telemetry.HTTPMiddleware(mux, "frontend")

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		log.Println("Frontend service starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
