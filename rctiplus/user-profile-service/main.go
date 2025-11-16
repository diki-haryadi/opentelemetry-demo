package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
	meter  metric.Meter

	// Metrics
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	activeRequests  metric.Int64UpDownCounter
)

// UserProfile represents a user profile entity
type UserProfile struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	IsActive     bool      `json:"is_active"`
	LastLogin    time.Time `json:"last_login"`
	Subscription string    `json:"subscription"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// In-memory storage for user profiles
var userProfiles = make(map[int]*UserProfile)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	shutdown, err := initOpenTelemetry(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	// Initialize sample data
	initSampleData()

	// Create router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/user-profiles/", handleUserProfileByUserID)

	// Wrap with OpenTelemetry middleware
	handler := otelhttp.NewHandler(mux, "user-profile-service")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Different port from main app
	}

	log.Printf("🚀 User Profile Service starting on port %s", port)
	log.Printf("📊 OpenTelemetry enabled - Sending to %s", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	log.Printf("📝 Available endpoints:")
	log.Printf("   GET  /")
	log.Printf("   GET  /health")
	log.Printf("   GET  /api/user-profiles/{user_id}")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initOpenTelemetry(ctx context.Context) (func(context.Context) error, error) {
	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("user-profile-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
			attribute.String("environment", "demo"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize Tracer
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tracerProvider.Tracer("user-profile-service")

	// Initialize Metrics
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{Name: "http.server.request.duration"},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
				},
			},
		)),
	)
	otel.SetMeterProvider(meterProvider)

	meter = meterProvider.Meter("user-profile-service")

	// Create metrics
	requestCounter, err = meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	requestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request duration histogram: %w", err)
	}

	activeRequests, err = meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active requests counter: %w", err)
	}

	// Return shutdown function
	return func(ctx context.Context) error {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}, nil
}

func initSampleData() {
	// Initialize user profiles
	baseTime := time.Now()
	userProfiles[1] = &UserProfile{
		ID:           1,
		UserID:       1,
		Phone:        "+6281234567890",
		Address:      "Jakarta, Indonesia",
		DateOfBirth:  time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		IsActive:     true,
		LastLogin:    baseTime.Add(-2 * time.Hour),
		Subscription: "Premium",
		CreatedAt:    baseTime.Add(-365 * 24 * time.Hour),
		UpdatedAt:    baseTime.Add(-1 * time.Hour),
	}

	userProfiles[2] = &UserProfile{
		ID:           2,
		UserID:       2,
		Phone:        "+6282234567890",
		Address:      "Bandung, Indonesia",
		DateOfBirth:  time.Date(1985, 11, 22, 0, 0, 0, 0, time.UTC),
		IsActive:     true,
		LastLogin:    baseTime.Add(-5 * time.Hour),
		Subscription: "Family",
		CreatedAt:    baseTime.Add(-180 * 24 * time.Hour),
		UpdatedAt:    baseTime.Add(-3 * time.Hour),
	}

	userProfiles[3] = &UserProfile{
		ID:           3,
		UserID:       3,
		Phone:        "+6283234567890",
		Address:      "Surabaya, Indonesia",
		DateOfBirth:  time.Date(1992, 8, 7, 0, 0, 0, 0, time.UTC),
		IsActive:     false,
		LastLogin:    baseTime.Add(-48 * time.Hour),
		Subscription: "Basic",
		CreatedAt:    baseTime.Add(-90 * 24 * time.Hour),
		UpdatedAt:    baseTime.Add(-24 * time.Hour),
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("endpoint", "home"))

	recordMetrics(ctx, "GET", "/", 200)

	response := Response{
		Success: true,
		Message: "Welcome to User Profile Service",
		Data: map[string]interface{}{
			"version": "1.0.0",
			"service": "user-profile-service",
			"endpoints": []string{
				"GET /health",
				"GET /api/user-profiles/{user_id}",
			},
		},
	}

	respondJSON(w, http.StatusOK, response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("endpoint", "health"))

	recordMetrics(ctx, "GET", "/health", 200)

	response := Response{
		Success: true,
		Message: "Service is healthy",
		Data: map[string]interface{}{
			"status":    "UP",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	respondJSON(w, http.StatusOK, response)
}

func handleUserProfileByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "handleUserProfileByUserID")
	defer span.End()

	// Extract user ID from URL
	userID := extractUserID(r.URL.Path, "/api/user-profiles/")
	span.SetAttributes(attribute.Int("user.id", userID))

	// Simulate processing delay
	simulateWork(ctx, 20, 100)

	profile, exists := userProfiles[userID]
	if !exists {
		recordMetrics(ctx, r.Method, "/api/user-profiles/{user_id}", 404)
		span.SetAttributes(attribute.Bool("profile.found", false))
		log.Printf("User profile not found for user ID: %d", userID)

		response := Response{
			Success: false,
			Message: "User profile not found",
		}
		respondJSON(w, http.StatusNotFound, response)
		return
	}

	span.SetAttributes(attribute.Bool("profile.found", true))
	recordMetrics(ctx, r.Method, "/api/user-profiles/{user_id}", 200)
	log.Printf("Retrieved profile for user: %d", profile.UserID)

	response := Response{
		Success: true,
		Data:    profile,
	}

	respondJSON(w, http.StatusOK, response)
}

func simulateWork(ctx context.Context, minMs, maxMs int) {
	_, span := tracer.Start(ctx, "simulateWork")
	defer span.End()

	duration := time.Duration(rand.Intn(maxMs-minMs)+minMs) * time.Millisecond
	span.SetAttributes(attribute.Int64("delay.ms", duration.Milliseconds()))
	time.Sleep(duration)
}

func recordMetrics(ctx context.Context, method, endpoint string, statusCode int) {
	attrs := metric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", endpoint),
		attribute.Int("http.status_code", statusCode),
	)

	requestCounter.Add(ctx, 1, attrs)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func extractUserID(path, prefix string) int {
	idStr := path[len(prefix):]
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	return id
}