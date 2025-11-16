package main

// import (
// 	"context"
// 	"encoding/json"
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
// 			semconv.ServiceName("backend"),
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
// 			semconv.ServiceName("backend"),
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

// 	return nil
// }

// type Product struct {
// 	ID          string  `json:"id"`
// 	Name        string  `json:"name"`
// 	Description string  `json:"description"`
// 	Price       float64 `json:"price"`
// 	Stock       int     `json:"stock"`
// }

// var products = []Product{
// 	{ID: "1", Name: "Laptop", Description: "High-performance laptop", Price: 1299.99, Stock: 15},
// 	{ID: "2", Name: "Mouse", Description: "Wireless mouse", Price: 29.99, Stock: 50},
// 	{ID: "3", Name: "Keyboard", Description: "Mechanical keyboard", Price: 89.99, Stock: 30},
// 	{ID: "4", Name: "Monitor", Description: "4K monitor", Price: 499.99, Stock: 20},
// 	{ID: "5", Name: "Headphones", Description: "Noise-cancelling headphones", Price: 199.99, Stock: 25},
// }

// func handleProducts(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	span := trace.SpanFromContext(ctx)
// 	span.SetAttributes(attribute.String("http.route", "/api/products"))

// 	start := time.Now()
// 	statusCode := http.StatusOK
// 	defer func() {
// 		duration := float64(time.Since(start).Milliseconds())
// 		requestDuration.Record(ctx, duration,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "backend"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/api/products"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 		requestCounter.Add(ctx, 1,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "backend"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/api/products"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 	}()

// 	log.Printf("Backend: Fetching products")

// 	// Simulate some processing time
// 	time.Sleep(time.Duration(rand.Intn(50)+10) * time.Millisecond)

// 	// Simulate occasional errors (5% chance)
// 	if rand.Float32() < 0.05 {
// 		statusCode = http.StatusInternalServerError
// 		span.SetStatus(codes.Error, "Internal server error")
// 		http.Error(w, "Internal server error", statusCode)
// 		return
// 	}

// 	// Call payment service to check pricing
// 	paymentURL := os.Getenv("PAYMENT_URL")
// 	if paymentURL == "" {
// 		paymentURL = "http://payment:8082"
// 	}

// 	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
// 	req, err := http.NewRequestWithContext(ctx, "GET", paymentURL+"/api/pricing", nil)
// 	if err != nil {
// 		log.Printf("Error creating request to payment service: %v", err)
// 	} else {
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			log.Printf("Error calling payment service: %v", err)
// 		} else {
// 			resp.Body.Close()
// 			log.Printf("Payment service pricing check completed")
// 		}
// 	}

// 	span.SetAttributes(attribute.Int("product.count", len(products)))

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	json.NewEncoder(w).Encode(products)
// }

// func handleHealth(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	start := time.Now()
// 	statusCode := http.StatusOK
// 	defer func() {
// 		duration := float64(time.Since(start).Milliseconds())
// 		requestDuration.Record(ctx, duration,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "backend"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/health"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 		requestCounter.Add(ctx, 1,
// 			metric.WithAttributes(
// 				attribute.String("service_name", "backend"),
// 				attribute.String("http_method", r.Method),
// 				attribute.String("http_route", "/health"),
// 				attribute.Int("status_code", statusCode),
// 			),
// 		)
// 	}()

// 	response := map[string]string{
// 		"status":  "healthy",
// 		"service": "backend",
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
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

// 	tracer = otel.Tracer("backend")

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

// 	meter = otel.Meter("backend")

// 	// Initialize metrics
// 	if err := initMetrics(); err != nil {
// 		log.Fatalf("Failed to initialize metrics: %v", err)
// 	}

// 	// Setup HTTP handlers with instrumentation
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/api/products", handleProducts)
// 	mux.HandleFunc("/health", handleHealth)

// 	handler := otelhttp.NewHandler(mux, "backend")

// 	server := &http.Server{
// 		Addr:    ":8081",
// 		Handler: handler,
// 	}

// 	// Graceful shutdown
// 	go func() {
// 		log.Println("Backend service starting on :8081")
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
