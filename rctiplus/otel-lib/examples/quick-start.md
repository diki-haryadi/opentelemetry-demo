# Quick Start Guide

Panduan cepat untuk menjalankan examples.

## Prerequisites

1. Go 1.23+ installed
2. Docker Compose running (untuk Tempo, Prometheus, Grafana)
3. OTLP endpoint accessible (default: `http://localhost:4317`)

## Setup dan Run

### Option 1: Menggunakan Setup Script

```bash
# Setup example
cd examples
./setup.sh basic

# Run example
cd basic
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

### Option 2: Manual Setup

```bash
# Navigate ke example directory
cd examples/basic

# Initialize go module
go mod init example-basic

# Add local library
go mod edit -replace github.com/rctiplus/otel-lib=../../..

# Install dependencies
go mod tidy

# Run
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

## Test Examples

### Basic Example

```bash
# Terminal 1: Run example
cd examples/basic
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go

# Terminal 2: Test
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/users
```

### Advanced Example

```bash
# Run
cd examples/advanced
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go

# Test
curl http://localhost:8080/api/external
```

### Custom Metrics Example

```bash
# Run
cd examples/custom-metrics
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go

# Test
curl http://localhost:8080/api/process
curl http://localhost:8080/api/cache
```

### Microservice Example

```bash
# Run
cd examples/microservice
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go

# Test
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "product_id": 100, "quantity": 2}'
```

## Verify di Grafana

1. Buka Grafana: http://localhost:3000
2. Navigate ke **Explore** → **Tempo**
3. Pilih service name dari example (misal: `example-service`, `advanced-example-service`)
4. Klik **Run Query** untuk melihat traces
5. Untuk service graph, buka dashboard **Service Graph Dashboard**

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | - | OTLP endpoint untuk traces dan metrics |
| `PORT` | `8080` | Port untuk HTTP server |
| `OTEL_SERVICE_NAME` | - | Service name (override dengan config) |

## Troubleshooting

### Error: "cannot find module"

```bash
# Pastikan replace directive sudah ditambahkan
go mod edit -replace github.com/rctiplus/otel-lib=../../..
go mod tidy
```

### Traces tidak muncul

1. Check Tempo is running: `docker ps | grep tempo`
2. Check endpoint: `echo $OTEL_EXPORTER_OTLP_ENDPOINT`
3. Check logs: `docker logs tempo`

### Port already in use

```bash
# Gunakan port berbeda
PORT=8081 OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

