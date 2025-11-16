# OpenTelemetry Library Examples

Koleksi contoh penggunaan OpenTelemetry Library dengan berbagai skenario.

## Examples

### 1. Basic Example (`basic/`)

Contoh penggunaan dasar library dengan konfigurasi minimal.

**Features:**
- Basic HTTP server dengan OpenTelemetry middleware
- Simple span dan metrics recording
- Graceful shutdown

**Run:**
```bash
cd basic
go mod init example-basic
go mod edit -replace github.com/rctiplus/otel-lib=../../..
go mod tidy
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

### 2. Advanced Example (`advanced/`)

Contoh penggunaan dengan konfigurasi advanced dan custom attributes.

**Features:**
- Custom resource attributes (team, region, deployment)
- Custom histogram boundaries
- Custom metric interval
- HTTP client instrumentation untuk external calls
- Error recording

**Run:**
```bash
cd advanced
go mod init example-advanced
go mod edit -replace github.com/rctiplus/otel-lib=../../..
go mod tidy
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

### 3. Custom Metrics (`custom-metrics/`)

Contoh membuat custom metrics selain default HTTP metrics.

**Features:**
- Custom counter untuk business operations
- Custom gauge untuk cache size
- Custom histogram untuk processing duration
- Menggunakan meter dari library

**Run:**
```bash
cd custom-metrics
go mod init example-custom-metrics
go mod edit -replace github.com/rctiplus/otel-lib=../../..
go mod tidy
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

### 4. Microservice Example (`microservice/`)

Contoh implementasi di microservice dengan multiple service calls.

**Features:**
- Service-to-service communication dengan trace propagation
- Multiple spans dalam satu request
- Peer service attributes
- Error handling dan recording

**Run:**
```bash
cd microservice
go mod init example-microservice
go mod edit -replace github.com/rctiplus/otel-lib=../../..
go mod tidy
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

## Testing Examples

### Test dengan cURL

```bash
# Basic example
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/users

# Advanced example
curl http://localhost:8080/api/external

# Custom metrics
curl http://localhost:8080/api/process
curl http://localhost:8080/api/cache

# Microservice
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "product_id": 100, "quantity": 2}'
```

### Verify di Grafana/Tempo

1. Buka Grafana: http://localhost:3000
2. Navigate ke Explore → Tempo
3. Query traces dengan service name yang sesuai
4. Lihat service graph untuk microservice example

## Common Patterns

### Pattern 1: Basic Handler

```go
func handleEndpoint(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Set attributes
    telemetry.SetSpanAttributes(ctx,
        attribute.String("endpoint", "my-endpoint"),
    )
    
    // Record metrics
    telemetry.RecordMetrics(ctx, r.Method, "/api/endpoint", 200)
    
    // Your logic here
}
```

### Pattern 2: Handler dengan Span

```go
func handleEndpoint(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ctx, span := telemetry.StartSpan(ctx, "handleEndpoint")
    defer span.End()
    
    // Set attributes
    telemetry.SetSpanAttributes(ctx,
        attribute.String("endpoint", "my-endpoint"),
        attribute.String("method", r.Method),
    )
    
    // Your logic here
    
    // Record metrics
    telemetry.RecordMetrics(ctx, r.Method, "/api/endpoint", 200)
}
```

### Pattern 3: External Service Call

```go
func callExternalService(ctx context.Context, url string) error {
    ctx, span := telemetry.StartSpan(ctx, "callExternalService")
    defer span.End()
    
    telemetry.SetSpanAttributes(ctx,
        attribute.String("http.url", url),
        attribute.String("peer.service", "external-service"),
    )
    
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
    )
    
    return nil
}
```

### Pattern 4: Custom Metrics

```go
// Create custom metrics
meter := telemetry.Meter()

customCounter, _ := meter.Int64Counter(
    "custom.operations.total",
    metric.WithDescription("Total operations"),
)

// Use in handler
customCounter.Add(ctx, 1, metric.WithAttributes(
    attribute.String("operation.type", "data_processing"),
))
```

## Best Practices

1. **Always use context**: Pass context through all function calls
2. **Set meaningful attributes**: Use semantic conventions when possible
3. **Record errors**: Always record errors using `RecordSpanError()`
4. **Use descriptive span names**: Use operation names like "handleUsers", "validateOrder"
5. **Set peer.service**: When calling external services, set `peer.service` attribute
6. **Record metrics consistently**: Use same endpoint patterns for metrics
7. **Graceful shutdown**: Always call `Shutdown()` on application exit

## Troubleshooting

### Traces tidak muncul di Tempo

1. Check `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable
2. Verify Tempo is running and accessible
3. Check network connectivity
4. Look for errors in application logs

### Metrics tidak muncul di Prometheus/Mimir

1. Check metric endpoint configuration
2. Verify metric interval is not too long
3. Check if metrics are being recorded (use debug logs)
4. Verify Prometheus/Mimir is scraping correctly

### Spans tidak connected

1. Ensure context is passed through all function calls
2. Use `HTTPClient()` from library for HTTP calls
3. Check propagators configuration
4. Verify trace context headers are being sent

