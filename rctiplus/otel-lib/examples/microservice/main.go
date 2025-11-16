package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rctiplus/otel-lib"
	"go.opentelemetry.io/otel/attribute"
)

var telemetry *otellib.Telemetry

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry untuk microservice
	var err error
	telemetry, err = otellib.NewBuilder().
		WithConfig(otellib.DefaultConfig().
			WithServiceName("order-service").
			WithServiceVersion("1.0.0").
			WithEnvironment("production").
			WithAttributes(
				attribute.String("service.type", "microservice"),
				attribute.String("domain", "e-commerce"),
			)).
		Build(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		telemetry.Shutdown(ctx)
	}()

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/api/orders", handleCreateOrder)
	mux.HandleFunc("/api/orders/", handleGetOrder)

	handler := telemetry.HTTPMiddleware(mux, "order-service")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Order Service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	telemetry.SetSpanAttributes(ctx, attribute.String("endpoint", "home"))
	telemetry.RecordMetrics(ctx, "GET", "/", 200)

	respondJSON(w, http.StatusOK, map[string]string{
		"service": "order-service",
		"version": "1.0.0",
	})
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := telemetry.StartSpan(ctx, "handleCreateOrder")
	defer span.End()

	if r.Method != http.MethodPost {
		telemetry.RecordMetrics(ctx, r.Method, "/api/orders", 405)
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		telemetry.RecordSpanError(ctx, err)
		telemetry.RecordMetrics(ctx, r.Method, "/api/orders", 400)
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	telemetry.SetSpanAttributes(ctx,
		attribute.Int("order.user_id", order.UserID),
		attribute.Int("order.product_id", order.ProductID),
		attribute.Int("order.quantity", order.Quantity),
	)

	// Validate user by calling user service
	ctx, validateSpan := telemetry.StartSpan(ctx, "validateUser")
	if err := validateUser(ctx, order.UserID); err != nil {
		validateSpan.End()
		telemetry.RecordSpanError(ctx, err)
		telemetry.RecordMetrics(ctx, r.Method, "/api/orders", 400)
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user"})
		return
	}
	validateSpan.End()

	// Validate product by calling product service
	ctx, productSpan := telemetry.StartSpan(ctx, "validateProduct")
	if err := validateProduct(ctx, order.ProductID); err != nil {
		productSpan.End()
		telemetry.RecordSpanError(ctx, err)
		telemetry.RecordMetrics(ctx, r.Method, "/api/orders", 400)
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid product"})
		return
	}
	productSpan.End()

	// Create order
	order.ID = 12345
	order.CreatedAt = time.Now()

	telemetry.SetSpanAttributes(ctx,
		attribute.Int("order.id", order.ID),
		attribute.Bool("order.created", true),
	)

	telemetry.RecordMetrics(ctx, r.Method, "/api/orders", 201)
	respondJSON(w, http.StatusCreated, order)
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := telemetry.StartSpan(ctx, "handleGetOrder")
	defer span.End()

	orderID := 12345 // In real app, extract from URL
	telemetry.SetSpanAttributes(ctx,
		attribute.Int("order.id", orderID),
	)

	// Simulate database query
	time.Sleep(20 * time.Millisecond)

	order := Order{
		ID:        orderID,
		UserID:    1,
		ProductID: 100,
		Quantity:  2,
		CreatedAt: time.Now(),
	}

	telemetry.RecordMetrics(ctx, r.Method, "/api/orders/{id}", 200)
	respondJSON(w, http.StatusOK, order)
}

// validateUser calls user service to validate user
func validateUser(ctx context.Context, userID int) error {
	ctx, span := telemetry.StartSpan(ctx, "validateUser")
	defer span.End()

	telemetry.SetSpanAttributes(ctx,
		attribute.Int("user.id", userID),
		attribute.String("peer.service", "user-service"),
	)

	// Use instrumented HTTP client
	client := telemetry.HTTPClient()
	url := fmt.Sprintf("http://user-service:8081/api/users/%d", userID)

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
	)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user validation failed: status %d", resp.StatusCode)
	}

	return nil
}

// validateProduct calls product service to validate product
func validateProduct(ctx context.Context, productID int) error {
	ctx, span := telemetry.StartSpan(ctx, "validateProduct")
	defer span.End()

	telemetry.SetSpanAttributes(ctx,
		attribute.Int("product.id", productID),
		attribute.String("peer.service", "product-service"),
	)

	// Use instrumented HTTP client
	client := telemetry.HTTPClient()
	url := fmt.Sprintf("http://product-service:8082/api/products/%d", productID)

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
	)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("product validation failed: status %d", resp.StatusCode)
	}

	return nil
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
