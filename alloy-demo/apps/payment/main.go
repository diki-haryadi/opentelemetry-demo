package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	telemetry     *otellib.Telemetry
	paymentAmount metric.Float64Histogram
)

func initTelemetry(ctx context.Context) (*otellib.Telemetry, error) {
	config := otellib.DefaultConfig().
		WithServiceName("payment").
		WithServiceVersion("1.0.0").
		WithEnvironment("production")

	return otellib.NewBuilder().
		WithConfig(config).
		Build(ctx)
}

type PaymentRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentResponse struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

func handlePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.route", "/api/payment"))

	start := time.Now()
	statusCode := http.StatusOK
	defer func() {
		duration := float64(time.Since(start).Milliseconds())
		telemetry.RequestDuration().Record(ctx, duration)
		telemetry.RecordMetrics(ctx, r.Method, "/api/payment", statusCode)
	}()

	log.Printf("Payment: Processing payment request")

	// Simulate payment processing time
	processingTime := time.Duration(rand.Intn(100)+50) * time.Millisecond
	time.Sleep(processingTime)

	// Simulate occasional payment failures (3% chance)
	if rand.Float32() < 0.03 {
		statusCode = http.StatusPaymentRequired
		span.SetStatus(codes.Error, "Payment declined")
		response := PaymentResponse{
			TransactionID: generateTransactionID(),
			Status:        "declined",
			Amount:        99.99,
			Currency:      "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
		return
	}

	amount := 99.99 + rand.Float64()*900.01 // Random amount between 99.99 and 1000
	if paymentAmount != nil {
		paymentAmount.Record(ctx, amount,
			metric.WithAttributes(
				attribute.String("currency", "USD"),
				attribute.String("status", "approved"),
			),
		)
	}

	span.SetAttributes(
		attribute.Float64("payment.amount", amount),
		attribute.String("payment.currency", "USD"),
		attribute.String("payment.status", "approved"),
	)

	response := PaymentResponse{
		TransactionID: generateTransactionID(),
		Status:        "approved",
		Amount:        amount,
		Currency:      "USD",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func handlePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.route", "/api/pricing"))

	start := time.Now()
	statusCode := http.StatusOK
	defer func() {
		duration := float64(time.Since(start).Milliseconds())
		telemetry.RequestDuration().Record(ctx, duration)
		telemetry.RecordMetrics(ctx, r.Method, "/api/pricing", statusCode)
	}()

	log.Printf("Payment: Fetching pricing information")

	// Simulate pricing lookup
	time.Sleep(time.Duration(rand.Intn(30)+10) * time.Millisecond)

	response := map[string]interface{}{
		"service":      "payment",
		"pricing_tier": "standard",
		"fee_percent":  2.9,
		"fixed_fee":    0.30,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		"service": "payment",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateTransactionID() string {
	return fmt.Sprintf("TXN-%d-%d", time.Now().Unix(), rand.Intn(10000))
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

	// Create custom payment amount metric
	paymentAmount, err = telemetry.Meter().Float64Histogram(
		"payment_amount",
		metric.WithDescription("Payment amount processed"),
		metric.WithUnit("USD"),
	)
	if err != nil {
		log.Printf("Failed to create payment amount metric: %v", err)
	}

	// Setup HTTP handlers with instrumentation
	mux := http.NewServeMux()
	mux.HandleFunc("/api/payment", handlePayment)
	mux.HandleFunc("/api/pricing", handlePricing)
	mux.HandleFunc("/health", handleHealth)

	handler := telemetry.HTTPMiddleware(mux, "payment")

	server := &http.Server{
		Addr:    ":8082",
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		log.Println("Payment service starting on :8082")
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
