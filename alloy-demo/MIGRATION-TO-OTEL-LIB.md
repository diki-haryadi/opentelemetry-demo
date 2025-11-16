# Migration to otel-lib Wrapper

## Overview

Semua aplikasi (frontend, backend, payment) telah diupdate untuk menggunakan `otel-lib` wrapper yang menyederhanakan inisialisasi dan penggunaan OpenTelemetry.

## Perubahan yang Dilakukan

### 1. Structure

```
apps/
├── otel-lib/              # Shared OpenTelemetry wrapper library
│   ├── builder.go         # Telemetry builder
│   ├── config.go          # Configuration
│   ├── http.go            # HTTP helpers
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── main.go            # ✅ Updated to use otel-lib
│   ├── go.mod             # ✅ Updated dependencies
│   └── Dockerfile
├── backend/
│   ├── main.go            # ✅ Updated to use otel-lib
│   ├── go.mod             # ✅ Updated dependencies
│   └── Dockerfile
└── payment/
    ├── main.go            # ✅ Updated to use otel-lib
    ├── go.mod             # ✅ Updated dependencies
    └── Dockerfile
```

### 2. go.mod Changes

**Sebelum:**
```go
module frontend

go 1.21

require (
    go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1
    go.opentelemetry.io/otel v1.21.0
    go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
    // ... many more dependencies
)
```

**Sesudah:**
```go
module frontend

go 1.23.0

require (
    github.com/rctiplus/otel-lib v0.0.0
    go.opentelemetry.io/otel v1.38.0
    go.opentelemetry.io/otel/attribute v1.38.0
    go.opentelemetry.io/otel/codes v1.38.0
    go.opentelemetry.io/otel/trace v1.38.0
)

replace github.com/rctiplus/otel-lib => ../otel-lib
```

### 3. Code Changes

#### Sebelum (Manual Setup)

```go
// Banyak boilerplate code
func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithInsecure(),
        otlptracegrpc.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
    )
    // ... 30+ lines of setup code
}

func initMeter(ctx context.Context) (*sdkmetric.MeterProvider, error) {
    // ... 20+ lines of setup code
}

func initMetrics() error {
    // ... manual metric creation
}

func main() {
    // Initialize tracer
    tp, err := initTracer(ctx)
    // ... error handling
    
    // Initialize meter
    mp, err := initMeter(ctx)
    // ... error handling
    
    // Initialize metrics
    if err := initMetrics(); err != nil {
        // ... error handling
    }
    
    // Manual HTTP client setup
    client := http.Client{
        Transport: otelhttp.NewTransport(http.DefaultTransport),
    }
}
```

#### Sesudah (Menggunakan otel-lib)

```go
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

func main() {
    ctx := context.Background()

    // Initialize telemetry - single call!
    var err error
    telemetry, err = initTelemetry(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize telemetry: %v", err)
    }
    defer telemetry.Shutdown(ctx)

    // Use instrumented HTTP client
    client := telemetry.HTTPClient()
    
    // Use instrumented HTTP handler
    handler := telemetry.HTTPMiddleware(mux, "frontend")
}
```

### 4. Metrics Recording

#### Sebelum

```go
defer func() {
    duration := float64(time.Since(start).Milliseconds())
    requestDuration.Record(ctx, duration,
        metric.WithAttributes(
            attribute.String("service_name", "frontend"),
            attribute.String("http_method", r.Method),
            attribute.String("http_route", "/products"),
            attribute.Int("status_code", statusCode),
        ),
    )
    requestCounter.Add(ctx, 1,
        metric.WithAttributes(
            attribute.String("service_name", "frontend"),
            attribute.String("http_method", r.Method),
            attribute.String("http_route", "/products"),
            attribute.Int("status_code", statusCode),
        ),
    )
}()
```

#### Sesudah

```go
defer func() {
    duration := float64(time.Since(start).Milliseconds())
    telemetry.RequestDuration().Record(ctx, duration)
    telemetry.RecordMetrics(ctx, r.Method, "/products", statusCode)
}()
```

## Benefits

### 1. **Simplified Code**
- Reduced boilerplate dari ~100 lines menjadi ~10 lines
- Single initialization call
- Automatic resource detection

### 2. **Consistent Configuration**
- Semua services menggunakan configuration yang sama
- Mudah untuk update configuration secara global
- Centralized defaults

### 3. **Better Maintainability**
- Perubahan di otel-lib otomatis apply ke semua services
- Easier to upgrade OpenTelemetry versions
- Single source of truth

### 4. **Pre-built Helpers**
- `HTTPClient()` - Instrumented HTTP client
- `HTTPMiddleware()` - Instrumented HTTP handler
- `RecordMetrics()` - Simplified metrics recording
- `RequestDuration()`, `RequestCounter()`, `ActiveRequests()` - Pre-built metrics

### 5. **Flexible Configuration**
- Builder pattern untuk easy customization
- Environment variable support
- Custom histogram boundaries
- Configurable propagators

## otel-lib Features

### Configuration Options

```go
config := otellib.DefaultConfig().
    WithServiceName("my-service").
    WithServiceVersion("2.0.0").
    WithEnvironment("staging").
    WithAttributes(
        attribute.String("team", "platform"),
        attribute.String("region", "us-west"),
    ).
    WithTraceEndpoint("alloy:4317").
    WithMetricEndpoint("alloy:4317").
    WithMetricInterval(30 * time.Second).
    WithHistogramBoundaries([]float64{0, 10, 50, 100, 500, 1000}).
    WithPropagators("tracecontext", "baggage")
```

### Pre-built Metrics

1. **http.server.request.count** - Request counter
2. **http.server.request.duration** - Request duration histogram (ms)
3. **http.server.active_requests** - Active requests gauge

### Helper Methods

```go
// HTTP instrumentation
client := telemetry.HTTPClient()
handler := telemetry.HTTPMiddleware(mux, "service-name")

// Metrics
telemetry.RecordMetrics(ctx, method, route, statusCode)
telemetry.RequestDuration().Record(ctx, duration)
telemetry.RequestCounter().Add(ctx, 1)

// Tracing
ctx, span := telemetry.StartSpan(ctx, "operation-name")
telemetry.SetSpanAttributes(ctx, attribute.String("key", "value"))
telemetry.RecordSpanError(ctx, err)

// Access raw components
tracer := telemetry.Tracer()
meter := telemetry.Meter()
```

## Testing

### Build Applications

```bash
cd alloy-demo

# Build all services
docker compose build

# Or build individually
docker compose build frontend
docker compose build backend
docker compose build payment
```

### Run Stack

```bash
# Start all services
docker compose up -d

# Check logs
docker compose logs -f frontend
docker compose logs -f backend
docker compose logs -f payment

# Generate traffic
./load-generator.sh
```

### Verify Telemetry

1. **Grafana**: http://localhost:3000
   - Service Graph: http://localhost:3000/d/service-graph
   - Metrics: http://localhost:3000/d/app-metrics

2. **Alloy UI**: http://localhost:12345

3. **Prometheus**: http://localhost:9090
   - Query: `http_server_request_count`
   - Query: `http_server_request_duration`

## Troubleshooting

### Build Errors

If you get import errors during build:

```bash
# In each app directory
cd apps/frontend
go mod tidy
go mod download

cd ../backend
go mod tidy
go mod download

cd ../payment
go mod tidy
go mod download
```

### Missing Metrics

Jika metrics tidak muncul:
1. Pastikan telemetry initialized dengan benar
2. Check Alloy logs: `docker compose logs alloy`
3. Verify OTLP endpoint configuration
4. Generate traffic dengan load generator

### Service Graph Not Showing

1. Tunggu 2-3 menit setelah startup
2. Generate traffic antar services
3. Check Tempo metrics: `curl http://localhost:3200/metrics | grep service_graph`

## Migration Checklist

- [x] Create otel-lib wrapper
- [x] Update frontend/go.mod
- [x] Update frontend/main.go
- [x] Update backend/go.mod
- [x] Update backend/main.go
- [x] Update payment/go.mod
- [x] Update payment/main.go
- [ ] Test docker build
- [ ] Test runtime behavior
- [ ] Verify metrics collection
- [ ] Verify trace propagation
- [ ] Verify service graph generation

## Next Steps

1. **Test the build**: `docker compose build`
2. **Run the stack**: `docker compose up -d`
3. **Generate traffic**: `./load-generator.sh`
4. **Verify dashboards**: Check Grafana
5. **Monitor logs**: `docker compose logs -f`

## Rollback Plan

Jika ada masalah, file original disimpan sebagai `main_old.go`:

```bash
# Rollback frontend
cd apps/frontend
mv main.go main_new.go
mv main_old.go main.go

# Rollback backend
cd ../backend
mv main.go main_new.go
mv main_old.go main.go

# Rollback payment
cd ../payment
mv main.go main_new.go
mv main_old.go main.go
```

## Summary

✅ **Completed:**
- otel-lib wrapper created with full functionality
- All 3 services migrated to use otel-lib
- Dependencies updated to latest versions
- Code simplified significantly (60-70% reduction)
- Consistent configuration across services

🔄 **Next:**
- Build and test with Docker
- Verify telemetry data flow
- Confirm service graph generation
- Performance testing

📚 **Documentation:**
- otel-lib/README.md - Library documentation
- otel-lib/USAGE.md - Usage examples
- This file - Migration guide
