# OpenTelemetry Library

Library wrapper untuk OpenTelemetry dengan pattern Factory Builder yang memudahkan implementasi di banyak service.

## Features

- ✅ Factory Builder Pattern untuk konfigurasi yang fleksibel
- ✅ Auto-initialization untuk tracer dan meter
- ✅ Pre-built HTTP metrics (counter, histogram, up-down counter)
- ✅ HTTP middleware dan client wrapper
- ✅ Helper functions untuk tracing dan metrics
- ✅ Konfigurasi melalui builder atau environment variables

## Installation

```bash
go get github.com/rctiplus/otel-lib
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/rctiplus/otel-lib"
)

func main() {
    ctx := context.Background()
    
    // Build OpenTelemetry dengan builder
    telemetry, err := otellib.NewBuilder().
        WithConfig(otellib.DefaultConfig().
            WithServiceName("my-service").
            WithServiceVersion("1.0.0").
            WithEnvironment("production")).
        Build(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
    }
    defer telemetry.Shutdown(ctx)
    
    // Wrap HTTP handler
    mux := http.NewServeMux()
    mux.HandleFunc("/", handleHome)
    
    handler := telemetry.HTTPMiddleware(mux, "my-service")
    
    http.ListenAndServe(":8080", handler)
}
```

### Advanced Configuration

```go
config := otellib.DefaultConfig().
    WithServiceName("my-service").
    WithServiceVersion("2.0.0").
    WithEnvironment("staging").
    WithAttributes(
        attribute.String("team", "backend"),
        attribute.String("region", "us-east-1"),
    ).
    WithMetricInterval(10 * time.Second).
    WithHistogramBoundaries([]float64{0, 10, 50, 100, 500, 1000}).
    WithPropagators("tracecontext", "baggage")

telemetry, err := otellib.NewBuilder().
    WithConfig(config).
    Build(ctx)
```

### Using Tracer

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Start span
    ctx, span := telemetry.StartSpan(ctx, "handleRequest")
    defer span.End()
    
    // Set attributes
    telemetry.SetSpanAttributes(ctx,
        attribute.String("user.id", "123"),
        attribute.String("operation", "get_user"),
    )
    
    // Your business logic here
    
    // Record metrics
    telemetry.RecordMetrics(ctx, "GET", "/api/users", 200)
}
```

### Using HTTP Client

```go
func callExternalAPI(ctx context.Context) error {
    client := telemetry.HTTPClient()
    
    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    if err != nil {
        return err
    }
    
    resp, err := client.Do(req)
    if err != nil {
        telemetry.RecordSpanError(ctx, err)
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

## Configuration Options

| Option | Method | Default | Description |
|--------|--------|---------|-------------|
| Service Name | `WithServiceName()` | `"unknown-service"` | Service name for resource |
| Service Version | `WithServiceVersion()` | `"1.0.0"` | Service version |
| Environment | `WithEnvironment()` | `"production"` | Environment name |
| Attributes | `WithAttributes()` | `[]` | Custom resource attributes |
| Trace Enabled | `WithTraceEnabled()` | `true` | Enable/disable tracing |
| Trace Endpoint | `WithTraceEndpoint()` | Env var | OTLP trace endpoint |
| Metric Enabled | `WithMetricEnabled()` | `true` | Enable/disable metrics |
| Metric Endpoint | `WithMetricEndpoint()` | Env var | OTLP metric endpoint |
| Metric Interval | `WithMetricInterval()` | `15s` | Metric export interval |
| Histogram Boundaries | `WithHistogramBoundaries()` | Default buckets | Custom histogram buckets |
| Propagators | `WithPropagators()` | `["tracecontext", "baggage"]` | Propagation formats |

## Environment Variables

Library juga mendukung environment variables standar OpenTelemetry:

- `OTEL_EXPORTER_OTLP_ENDPOINT` - OTLP endpoint untuk traces dan metrics
- `OTEL_SERVICE_NAME` - Service name (akan di-override oleh config)
- `OTEL_RESOURCE_ATTRIBUTES` - Resource attributes tambahan

## Default Metrics

Library secara otomatis membuat metrics berikut:

- `http.server.request.count` - Counter untuk jumlah HTTP requests
- `http.server.request.duration` - Histogram untuk durasi HTTP requests
- `http.server.active_requests` - UpDownCounter untuk active requests

## Examples

Lihat contoh implementasi di:
- `../app/main.go` - Main application
- `../user-profile-service/main.go` - User profile service

