# Cara Menggunakan OpenTelemetry Library

## Setup sebagai Local Module

Karena library ini adalah local module, Anda perlu menambahkan replace directive di `go.mod` service yang akan menggunakannya.

### 1. Update go.mod di service

Tambahkan di `go.mod` service (app atau user-profile-service):

```go
module rctiplus-api

go 1.23.0

require (
    github.com/rctiplus/otel-lib v0.0.0
    // ... dependencies lainnya
)

replace github.com/rctiplus/otel-lib => ../otel-lib
```

Atau jalankan command:
```bash
go mod edit -replace github.com/rctiplus/otel-lib=../otel-lib
go mod tidy
```

### 2. Contoh Implementasi di app/main.go

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    
    "github.com/rctiplus/otel-lib"
)

var telemetry *otellib.Telemetry

func main() {
    ctx := context.Background()
    
    // Initialize OpenTelemetry dengan builder pattern
    var err error
    telemetry, err = otellib.NewBuilder().
        WithConfig(otellib.DefaultConfig().
            WithServiceName("rctiplus-api").
            WithServiceVersion("1.0.0").
            WithEnvironment("demo")).
        Build(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
    }
    defer telemetry.Shutdown(ctx)
    
    // Setup HTTP routes
    mux := http.NewServeMux()
    mux.HandleFunc("/", handleHome)
    
    // Wrap dengan OpenTelemetry middleware
    handler := telemetry.HTTPMiddleware(mux, "rctiplus-api")
    
    http.ListenAndServe(":8080", handler)
}
```

### 3. Contoh di Handler

```go
func handleHome(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Set span attributes
    telemetry.SetSpanAttributes(ctx,
        attribute.String("endpoint", "home"),
    )
    
    // Record metrics
    telemetry.RecordMetrics(ctx, "GET", "/", 200)
    
    // Your business logic
    w.Write([]byte("Hello"))
}
```

### 4. Contoh HTTP Client

```go
func callExternalService(ctx context.Context) error {
    // Gunakan HTTP client dari library (sudah di-instrument)
    client := telemetry.HTTPClient()
    
    req, err := http.NewRequestWithContext(ctx, "GET", 
        "http://other-service:8080/api/data", nil)
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

## Migration dari Kode Lama

### Sebelum (Manual Setup):
```go
func initOpenTelemetry(ctx context.Context) (func(context.Context) error, error) {
    res, err := resource.New(ctx, ...)
    traceExporter, err := otlptracegrpc.New(ctx, ...)
    tracerProvider := sdktrace.NewTracerProvider(...)
    // ... banyak kode boilerplate
}
```

### Sesudah (Dengan Library):
```go
telemetry, err := otellib.NewBuilder().
    WithConfig(otellib.DefaultConfig().
        WithServiceName("my-service")).
    Build(ctx)
```

## Benefits

1. **Kurang Boilerplate**: Dari ~100 baris menjadi ~10 baris
2. **Konsisten**: Semua service menggunakan konfigurasi yang sama
3. **Mudah Maintenance**: Update library sekali, semua service ter-update
4. **Type Safe**: Builder pattern memastikan konfigurasi yang valid
5. **Fleksibel**: Masih bisa customize jika diperlukan

