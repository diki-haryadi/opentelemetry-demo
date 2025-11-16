package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	otellib "github.com/diki-haryadi/otel-lib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	telemetry *otellib.Telemetry
)

func initTelemetry(ctx context.Context) (*otellib.Telemetry, error) {
	config := otellib.DefaultConfig().
		WithServiceName("backend").
		WithServiceVersion("1.0.0").
		WithEnvironment("production")

	return otellib.NewBuilder().
		WithConfig(config).
		Build(ctx)
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

var products = []Product{
	{ID: "1", Name: "Laptop", Description: "High-performance laptop", Price: 1299.99, Stock: 15},
	{ID: "2", Name: "Mouse", Description: "Wireless mouse", Price: 29.99, Stock: 50},
	{ID: "3", Name: "Keyboard", Description: "Mechanical keyboard", Price: 89.99, Stock: 30},
	{ID: "4", Name: "Monitor", Description: "4K monitor", Price: 499.99, Stock: 20},
	{ID: "5", Name: "Headphones", Description: "Noise-cancelling headphones", Price: 199.99, Stock: 25},
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.route", "/api/products"))

	start := time.Now()
	statusCode := http.StatusOK
	defer func() {
		duration := float64(time.Since(start).Milliseconds())
		telemetry.RequestDuration().Record(ctx, duration)
		telemetry.RecordMetrics(ctx, r.Method, "/api/products", statusCode)
	}()

	log.Printf("Backend: Fetching products")

	// Simulate some processing time
	time.Sleep(time.Duration(rand.Intn(50)+10) * time.Millisecond)

	// Simulate occasional errors (5% chance)
	if rand.Float32() < 0.05 {
		statusCode = http.StatusInternalServerError
		span.SetStatus(codes.Error, "Internal server error")
		http.Error(w, "Internal server error", statusCode)
		return
	}

	// Call payment service to check pricing
	paymentURL := os.Getenv("PAYMENT_URL")
	if paymentURL == "" {
		paymentURL = "http://payment:8082"
	}

	client := telemetry.HTTPClient()
	req, err := http.NewRequestWithContext(ctx, "GET", paymentURL+"/api/pricing", nil)
	if err != nil {
		log.Printf("Error creating request to payment service: %v", err)
	} else {
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error calling payment service: %v", err)
		} else {
			resp.Body.Close()
			log.Printf("Payment service pricing check completed")
		}
	}

	span.SetAttributes(attribute.Int("product.count", len(products)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(products)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()
	statusCode := http.StatusOK
	defer func() {
		duration := float64(time.Since(start).Milliseconds())
		telemetry.RequestDuration().Record(ctx, duration)
		telemetry.RecordMetrics(ctx, r.Method, "/health", statusCode)
	}()

	response := map[string]string{
		"status":  "healthy",
		"service": "backend",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	mux.HandleFunc("/api/products", handleProducts)
	mux.HandleFunc("/health", handleHealth)

	handler := telemetry.HTTPMiddleware(mux, "backend")

	server := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		log.Println("Backend service starting on :8081")
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
