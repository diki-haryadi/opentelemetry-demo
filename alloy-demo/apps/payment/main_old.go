package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"math/rand"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/attribute"
// 	"go.opentelemetry.io/otel/codes"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
// 	"go.opentelemetry.io/otel/metric"
// 	"go.opentelemetry.io/otel/propagation"
// 	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
// 	"go.opentelemetry.io/otel/sdk/resource"
// 	sdktrace "go.opentelemetry.io/otel/sdk/trace"
// 	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
// 	"go.opentelemetry.io/otel/trace"
// )

// var (
// 	tracer trace.Tracer
// 	meter  metric.Meter

// 	// Metrics
// 	requestCounter  metric.Int64Counter
// 	requestDuration metric.Float64Histogram
// 	paymentAmount   metric.Float64Histogram
// )

// func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
// 	exporter, err := otlptracegrpc.New(ctx,
// 		otlptracegrpc.WithInsecure(),
// 		otlptracegrpc.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := resource.New(ctx,
// 		resource.WithAttributes(
// 			semconv.ServiceName("payment"),
// 			semconv.ServiceVersion("1.0.0"),
// 		),
// 		resource.WithFromEnv(),
// 		resource.WithProcess(),
// 		resource.WithHost(),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	tp := sdktrace.NewTracerProvider(
// 		sdktrace.WithBatcher(exporter),
// 		sdktrace.WithResource(res),
// 		sdktrace.WithSampler(sdktrace.AlwaysSample()),
// 	)

// 	otel.SetTracerProvider(tp)
// 	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
// 		propagation.TraceContext{},
// 		propagation.Baggage{},
// 	))

// 	return tp, nil
// }

// func initMeter(ctx context.Context) (*sdkmetric.MeterProvider, error) {
// 	exporter, err := otlpmetricgrpc.New(ctx,
// 		otlpmetricgrpc.WithInsecure(),
// 		otlpmetricgrpc.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := resource.New(ctx,
// 		resource.WithAttributes(
// 			semconv.ServiceName("payment"),
// 			semconv.ServiceVersion("1.0.0"),
// 		),
// 		resource.WithFromEnv(),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	mp := sdkmetric.NewMeterProvider(
// 		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
// 		sdkmetric.WithResource(res),
// 	)

// 	otel.SetMeterProvider(mp)
// 	return mp, nil
// }

// func initMetrics() error {
// 	var err error
// 	requestCounter, err = meter.Int64Counter(
// 		"calls_total",
// 		metric.WithDescription("Total number of HTTP requests"),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	requestDuration, err = meter.Float64Histogram(
// 		"duration_milliseconds",
// 		metric.WithDescription("HTTP request duration in milliseconds"),
// 		metric.WithUnit("ms"),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	paymentAmount, err = meter.Float64Histogram(
// 		"payment_amount",
// 		metric.WithDescription("Payment amount processed"),
// 		metric.WithUnit("USD"),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// type PaymentRequest struct {
// 	Amount   float64 `json:"amount"`
// 	Currency string  `json:"currency"`
// }

// type PaymentResponse struct {
// 	TransactionID string  `json:"transaction_id"`
// 	Status        string  `json:"status"`
// 	Amount        float64 `json:"amount"`
// 	Currency      string  `json:"currency"`
// }

// func handlePayment(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	span := trace.SpanFromContext(ctx)
// 	span.SetAttributes(attribute.String("http.route", "/api/payment"))

// 	start := time.Now()
// 	statusCode := http.StatusOK
// 	defer func() {
// 		duration := float64(time.Since(start).Milliseconds())
// 		requestDuration.Record(ctx, duration,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "payment"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/api/payment"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 		requestCounter.Add(ctx, 1,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "payment"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/api/payment"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 	}()

// 	log.Printf("Payment: Processing payment request")

// 	// Simulate payment processing time
// 	processingTime := time.Duration(rand.Intn(100)+50) * time.Millisecond
// 	time.Sleep(processingTime)

// 	// Simulate occasional payment failures (3% chance)
// 	if rand.Float32() < 0.03 {
// 		statusCode = http.StatusPaymentRequired
// 		span.SetStatus(codes.Error, "Payment declined")
// 		response := PaymentResponse{
// 			TransactionID: generateTransactionID(),
// 			Status:        "declined",
// 			Amount:        99.99,
// 			Currency:      "USD",
// 		}
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(statusCode)
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	amount := 99.99 + rand.Float64()*900.01 // Random amount between 99.99 and 1000
// 	paymentAmount.Record(ctx, amount,
// 		metric.WithAttributes(
// 			attribute.String("currency", "USD"),
// 			attribute.String("status", "approved"),
// 		),
// 	)

// 	span.SetAttributes(
// 		attribute.Float64("payment.amount", amount),
// 		attribute.String("payment.currency", "USD"),
// 		attribute.String("payment.status", "approved"),
// 	)

// 	response := PaymentResponse{
// 		TransactionID: generateTransactionID(),
// 		Status:        "approved",
// 		Amount:        amount,
// 		Currency:      "USD",
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	json.NewEncoder(w).Encode(response)
// }

// func handlePricing(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	span := trace.SpanFromContext(ctx)
// 	span.SetAttributes(attribute.String("http.route", "/api/pricing"))

// 	start := time.Now()
// 	statusCode := http.StatusOK
// 	defer func() {
// 		duration := float64(time.Since(start).Milliseconds())
// 		requestDuration.Record(ctx, duration,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "payment"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/api/pricing"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 		requestCounter.Add(ctx, 1,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "payment"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/api/pricing"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 	}()

// 	log.Printf("Payment: Fetching pricing information")

// 	// Simulate pricing lookup
// 	time.Sleep(time.Duration(rand.Intn(30)+10) * time.Millisecond)

// 	response := map[string]interface{}{
// 		"service":      "payment",
// 		"pricing_tier": "standard",
// 		"fee_percent":  2.9,
// 		"fixed_fee":    0.30,
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }

// func handleHealth(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	start := time.Now()
// 	statusCode := http.StatusOK
// 	defer func() {
// 		duration := float64(time.Since(start).Milliseconds())
// 		requestDuration.Record(ctx, duration,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "payment"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/health"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 		requestCounter.Add(ctx, 1,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "payment"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/health"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 	}()

// 	response := map[string]string{
// 		"status":  "healthy",
// 		"service": "payment",
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }

// func generateTransactionID() string {
// 	return fmt.Sprintf("TXN-%d-%d", time.Now().Unix(), rand.Intn(10000))
// }

// func main() {
// 	ctx := context.Background()

// 	// Initialize tracer
// 	tp, err := initTracer(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize tracer: %v", err)
// 	}
// 	defer func() {
// 		if err := tp.Shutdown(ctx); err != nil {
// 			log.Printf("Error shutting down tracer provider: %v", err)
// 		}
// 	}()

// 	tracer = otel.Tracer("payment")

// 	// Initialize meter
// 	mp, err := initMeter(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize meter: %v", err)
// 	}
// 	defer func() {
// 		if err := mp.Shutdown(ctx); err != nil {
// 			log.Printf("Error shutting down meter provider: %v", err)
// 		}
// 	}()

// 	meter = otel.Meter("payment")

// 	// Initialize metrics
// 	if err := initMetrics(); err != nil {
// 		log.Fatalf("Failed to initialize metrics: %v", err)
// 	}

// 	// Setup HTTP handlers with instrumentation
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/api/payment", handlePayment)
// 	mux.HandleFunc("/api/pricing", handlePricing)
// 	mux.HandleFunc("/health", handleHealth)

// 	handler := otelhttp.NewHandler(mux, "payment")

// 	server := &http.Server{
// 		Addr:    ":8082",
// 		Handler: handler,
// 	}

// 	// Graceful shutdown
// 	go func() {
// 		log.Println("Payment service starting on :8082")
// 		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			log.Fatalf("Server failed: %v", err)
// 		}
// 	}()

// 	// Wait for interrupt signal
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit

// 	log.Println("Shutting down server...")
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	if err := server.Shutdown(ctx); err != nil {
// 		log.Fatalf("Server forced to shutdown: %v", err)
// 	}

// 	log.Println("Server exited")
// }
