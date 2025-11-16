package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
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

// UserProfile represents a user profile entity from the user-profile-service
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

// User represents a user entity
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserWithProfile represents a user with profile information
type UserWithProfile struct {
	User    *User        `json:"user"`
	Profile *UserProfile `json:"profile,omitempty"`
}

// Product represents a product entity
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// In-memory storage
var (
	users    = make(map[int]*User)
	products = make(map[int]*Product)
)

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
	mux.HandleFunc("/api/users", handleUsers)
	mux.HandleFunc("/api/users/", handleUserByID)
	mux.HandleFunc("/api/users-with-profile/", handleUserWithProfileByID)
	mux.HandleFunc("/api/products", handleProducts)
	mux.HandleFunc("/api/products/", handleProductByID)
	mux.HandleFunc("/api/orders", handleOrders)

	// Wrap with OpenTelemetry middleware
	handler := otelhttp.NewHandler(mux, "rctiplus-api")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 RCTI Plus API Server starting on port %s", port)
	log.Printf("📊 OpenTelemetry enabled - Sending to %s", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	log.Printf("📝 Available endpoints:")
	log.Printf("   GET  /")
	log.Printf("   GET  /health")
	log.Printf("   GET  /api/users")
	log.Printf("   GET  /api/users/{id}")
	log.Printf("   GET  /api/users-with-profile/{id}")
	log.Printf("   GET  /api/products")
	log.Printf("   GET  /api/products/{id}")
	log.Printf("   POST /api/orders")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initOpenTelemetry(ctx context.Context) (func(context.Context) error, error) {
	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("rctiplus-api"),
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

	tracer = tracerProvider.Tracer("rctiplus-api")

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

	meter = meterProvider.Meter("rctiplus-api")

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
	// Initialize users
	users[1] = &User{ID: 1, Name: "John Doe", Email: "john@rctiplus.id", CreatedAt: time.Now()}
	users[2] = &User{ID: 2, Name: "Jane Smith", Email: "jane@rctiplus.id", CreatedAt: time.Now()}
	users[3] = &User{ID: 3, Name: "Bob Wilson", Email: "bob@rctiplus.id", CreatedAt: time.Now()}

	// Initialize products
	products[1] = &Product{ID: 1, Name: "RCTI+ Premium", Description: "Premium streaming subscription", Price: 49000, Stock: 100}
	products[2] = &Product{ID: 2, Name: "RCTI+ Sport", Description: "Sports streaming package", Price: 39000, Stock: 100}
	products[3] = &Product{ID: 3, Name: "RCTI+ Family", Description: "Family entertainment package", Price: 59000, Stock: 100}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("endpoint", "home"))

	recordMetrics(ctx, "GET", "/", 200)

	response := Response{
		Success: true,
		Message: "Welcome to RCTI Plus API",
		Data: map[string]interface{}{
			"version": "1.0.0",
			"endpoints": []string{
				"GET /health",
				"GET /api/users",
				"GET /api/users/{id}",
				"GET /api/products",
				"GET /api/products/{id}",
				"POST /api/orders",
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

func handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "handleUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint", "users"),
		attribute.String("method", r.Method),
	)

	// Simulate processing delay
	simulateWork(ctx, 50, 150)

	userList := make([]*User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}

	span.SetAttributes(attribute.Int("users.count", len(userList)))
	recordMetrics(ctx, r.Method, "/api/users", 200)

	log.Printf("Retrieved %d users", len(userList))

	response := Response{
		Success: true,
		Data:    userList,
	}

	respondJSON(w, http.StatusOK, response)
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "handleUserByID")
	defer span.End()

	// Extract ID from URL
	id := extractID(r.URL.Path, "/api/users/")
	span.SetAttributes(attribute.Int("user.id", id))

	// Simulate database query
	simulateWork(ctx, 20, 80)

	user, exists := users[id]
	if !exists {
		recordMetrics(ctx, r.Method, "/api/users/{id}", 404)
		span.SetAttributes(attribute.Bool("user.found", false))
		log.Printf("User not found: %d", id)

		response := Response{
			Success: false,
			Message: "User not found",
		}
		respondJSON(w, http.StatusNotFound, response)
		return
	}

	span.SetAttributes(attribute.Bool("user.found", true))
	recordMetrics(ctx, r.Method, "/api/users/{id}", 200)
	log.Printf("Retrieved user: %d - %s", user.ID, user.Name)

	response := Response{
		Success: true,
		Data:    user,
	}

	respondJSON(w, http.StatusOK, response)
}

func handleUserWithProfileByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "handleUserWithProfileByID")
	defer span.End()

	// Extract ID from URL
	userID := extractID(r.URL.Path, "/api/users-with-profile/")
	span.SetAttributes(attribute.Int("user.id", userID))

	// Simulate database query for user
	simulateWork(ctx, 20, 80)

	user, exists := users[userID]
	if !exists {
		recordMetrics(ctx, r.Method, "/api/users-with-profile/{id}", 404)
		span.SetAttributes(attribute.Bool("user.found", false))
		log.Printf("User not found: %d", userID)

		response := Response{
			Success: false,
			Message: "User not found",
		}
		respondJSON(w, http.StatusNotFound, response)
		return
	}

	span.SetAttributes(attribute.Bool("user.found", true))

	// Call user profile service to get profile details
	profile, err := getUserProfile(ctx, userID)
	if err != nil {
		log.Printf("Error getting profile for user %d: %v", userID, err)
		// Continue without profile if service is unavailable
	}

	userWithProfile := &UserWithProfile{
		User:    user,
		Profile: profile,
	}

	span.SetAttributes(attribute.Bool("profile.found", profile != nil))
	recordMetrics(ctx, r.Method, "/api/users-with-profile/{id}", 200)
	log.Printf("Retrieved user with profile: %d - %s", user.ID, user.Name)

	response := Response{
		Success: true,
		Data:    userWithProfile,
	}

	respondJSON(w, http.StatusOK, response)
}

// getUserProfile calls the user-profile-service to get profile information
func getUserProfile(ctx context.Context, userID int) (*UserProfile, error) {
	ctx, span := tracer.Start(ctx, "getUserProfile")
	defer span.End()

	// Add attributes to the span
	span.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", fmt.Sprintf("http://user-profile-service:8081/api/user-profiles/%d", userID)),
		attribute.Int("user.id", userID),
	)

	// Create HTTP request with context to propagate trace
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://user-profile-service:8081/api/user-profiles/%d", userID), nil)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Use otelhttp to instrument the HTTP client
	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Tambahkan atribut peer.service
	span.SetAttributes(attribute.String("peer.service", "user-profile-service"))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			log.Printf("User profile not found for user ID: %d", userID)
			return nil, nil // Return nil for not found, which is not an error
		}
		err := fmt.Errorf("user profile service returned status: %d", resp.StatusCode)
		span.RecordError(err)
		return nil, err
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Unmarshal response
	var userProfileResponse Response
	if err := json.Unmarshal(body, &userProfileResponse); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Extract profile data from response
	if userProfile, ok := userProfileResponse.Data.(*UserProfile); ok {
		return userProfile, nil
	}

	// If unmarshaling directly doesn't work, try unmarshaling to a generic map
	var dataMap map[string]interface{}
	if err := json.Unmarshal(body, &dataMap); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check if response has success field
	if success, ok := dataMap["success"]; ok && success == true {
		if userData, ok := dataMap["data"]; ok {
			jsonData, _ := json.Marshal(userData)
			var profile UserProfile
			if err := json.Unmarshal(jsonData, &profile); err != nil {
				span.RecordError(err)
				return nil, err
			}
			return &profile, nil
		}
	}

	return nil, nil
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "handleProducts")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint", "products"),
		attribute.String("method", r.Method),
	)

	// Simulate processing delay
	simulateWork(ctx, 30, 100)

	productList := make([]*Product, 0, len(products))
	for _, product := range products {
		productList = append(productList, product)
	}

	span.SetAttributes(attribute.Int("products.count", len(productList)))
	recordMetrics(ctx, r.Method, "/api/products", 200)

	log.Printf("Retrieved %d products", len(productList))

	response := Response{
		Success: true,
		Data:    productList,
	}

	respondJSON(w, http.StatusOK, response)
}

func handleProductByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "handleProductByID")
	defer span.End()

	// Extract ID from URL
	id := extractID(r.URL.Path, "/api/products/")
	span.SetAttributes(attribute.Int("product.id", id))

	// Simulate database query
	simulateWork(ctx, 20, 80)

	product, exists := products[id]
	if !exists {
		recordMetrics(ctx, r.Method, "/api/products/{id}", 404)
		span.SetAttributes(attribute.Bool("product.found", false))
		log.Printf("Product not found: %d", id)

		response := Response{
			Success: false,
			Message: "Product not found",
		}
		respondJSON(w, http.StatusNotFound, response)
		return
	}

	span.SetAttributes(attribute.Bool("product.found", true))
	recordMetrics(ctx, r.Method, "/api/products/{id}", 200)
	log.Printf("Retrieved product: %d - %s", product.ID, product.Name)

	response := Response{
		Success: true,
		Data:    product,
	}

	respondJSON(w, http.StatusOK, response)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "handleOrders")
	defer span.End()

	if r.Method != http.MethodPost {
		recordMetrics(ctx, r.Method, "/api/orders", 405)
		response := Response{
			Success: false,
			Message: "Method not allowed",
		}
		respondJSON(w, http.StatusMethodNotAllowed, response)
		return
	}

	var order struct {
		UserID    int `json:"user_id"`
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		recordMetrics(ctx, r.Method, "/api/orders", 400)
		span.RecordError(err)
		log.Printf("Invalid request body: %v", err)

		response := Response{
			Success: false,
			Message: "Invalid request body",
		}
		respondJSON(w, http.StatusBadRequest, response)
		return
	}

	span.SetAttributes(
		attribute.Int("order.user_id", order.UserID),
		attribute.Int("order.product_id", order.ProductID),
		attribute.Int("order.quantity", order.Quantity),
	)

	// Validate user
	ctx, userSpan := tracer.Start(ctx, "validateUser")
	user, userExists := users[order.UserID]
	simulateWork(ctx, 10, 30)
	userSpan.End()

	if !userExists {
		recordMetrics(ctx, r.Method, "/api/orders", 400)
		span.SetAttributes(attribute.Bool("validation.user", false))
		log.Printf("Invalid user ID: %d", order.UserID)

		response := Response{
			Success: false,
			Message: "Invalid user ID",
		}
		respondJSON(w, http.StatusBadRequest, response)
		return
	}

	// Validate product
	ctx, productSpan := tracer.Start(ctx, "validateProduct")
	product, productExists := products[order.ProductID]
	simulateWork(ctx, 10, 30)
	productSpan.End()

	if !productExists {
		recordMetrics(ctx, r.Method, "/api/orders", 400)
		span.SetAttributes(attribute.Bool("validation.product", false))
		log.Printf("Invalid product ID: %d", order.ProductID)

		response := Response{
			Success: false,
			Message: "Invalid product ID",
		}
		respondJSON(w, http.StatusBadRequest, response)
		return
	}

	// Check stock
	if product.Stock < order.Quantity {
		recordMetrics(ctx, r.Method, "/api/orders", 400)
		span.SetAttributes(attribute.Bool("validation.stock", false))
		log.Printf("Insufficient stock for product %d", order.ProductID)

		response := Response{
			Success: false,
			Message: "Insufficient stock",
		}
		respondJSON(w, http.StatusBadRequest, response)
		return
	}

	// Process order
	ctx, processSpan := tracer.Start(ctx, "processOrder")
	simulateWork(ctx, 100, 300)
	product.Stock -= order.Quantity
	processSpan.SetAttributes(
		attribute.Int("product.remaining_stock", product.Stock),
	)
	processSpan.End()

	totalPrice := product.Price * float64(order.Quantity)
	span.SetAttributes(
		attribute.Float64("order.total_price", totalPrice),
		attribute.Bool("order.success", true),
	)

	recordMetrics(ctx, r.Method, "/api/orders", 201)
	log.Printf("Order created - User: %s, Product: %s, Qty: %d, Total: %.2f",
		user.Name, product.Name, order.Quantity, totalPrice)

	response := Response{
		Success: true,
		Message: "Order created successfully",
		Data: map[string]interface{}{
			"order_id":    rand.Intn(10000),
			"user":        user.Name,
			"product":     product.Name,
			"quantity":    order.Quantity,
			"total_price": totalPrice,
			"created_at":  time.Now().Format(time.RFC3339),
		},
	}

	respondJSON(w, http.StatusCreated, response)
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

func extractID(path, prefix string) int {
	idStr := path[len(prefix):]
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	return id
}
