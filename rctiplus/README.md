# RCTI Plus - OpenTelemetry Demo

Demo aplikasi microservice REST API dengan Golang yang terintegrasi dengan full observability stack menggunakan OpenTelemetry.

## Architecture

```
┌─────────┐
│  Apps   │
│ (Golang)│
└────┬────┘
     │
     v
┌─────────┐        ┌──────────┐
│  Alloy  │───────>│  Tempo   │ (Traces)
│(Collect)│        └──────────┘
└─────────┘              │
     │                   │
     ├──────────────┐    │
     │              │    │
     v              v    v
┌──────────┐   ┌──────────┐
│   Loki   │   │Prometheus│ (Metrics)
│  (Logs)  │   │          │
└──────────┘   └──────────┘
     │              │
     └──────┬───────┘
            │
            v
      ┌──────────┐
      │ Grafana  │ (Visualization)
      └──────────┘
```

## Components

- **App**: Golang REST API microservice dengan OpenTelemetry instrumentation
- **Alloy**: Grafana Alloy sebagai collector untuk traces, metrics, dan logs
- **Tempo**: Backend untuk menyimpan traces
- **Mimir/Prometheus**: Backend untuk menyimpan metrics
- **Loki**: Backend untuk menyimpan logs
- **Grafana**: Dashboard untuk visualisasi semua telemetry data

## Prerequisites

- Docker & Docker Compose
- Go 1.21+ (jika ingin development lokal)

## Quick Start

### 1. Clone & Navigate

```bash
cd rctiplus
```

### 2. Start Infrastructure

```bash
docker-compose up -d
```

Tunggu beberapa saat hingga semua services running. Cek status:

```bash
docker-compose ps
```

### 3. Access Services

- **API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (auto-login enabled)
  - Service Graph Dashboard available in Dashboards
  - **Loki logs**: Access via Grafana Explore → Select "Loki" datasource
- **Jaeger UI**: http://localhost:16686
  - Service Graph available in Dependencies tab
- **Alloy UI**: http://localhost:12345
- **Prometheus**: http://localhost:9090
- **Tempo**: http://localhost:3200
- **Loki API**: http://localhost:3100 (API endpoint only, no web UI)
  - Use Grafana Explore to query logs, or use LogCLI tool

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Get All Users
```bash
curl http://localhost:8080/api/users
```

### Get User by ID
```bash
curl http://localhost:8080/api/users/1
```

### Get All Products
```bash
curl http://localhost:8080/api/products
```

### Get Product by ID
```bash
curl http://localhost:8080/api/products/1
```

### Create Order
```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "product_id": 1,
    "quantity": 2
  }'
```

## Load Testing

Generate traffic untuk melihat traces & metrics:

```bash
# Install hey (HTTP load generator)
# macOS: brew install hey
# Linux: go install github.com/rakyll/hey@latest

# Generate load
hey -n 1000 -c 10 http://localhost:8080/api/users
hey -n 500 -c 5 http://localhost:8080/api/products
```

Atau gunakan script sederhana:

```bash
# Bash script
for i in {1..100}; do
  curl -s http://localhost:8080/api/users > /dev/null
  curl -s http://localhost:8080/api/products > /dev/null
  curl -s -X POST http://localhost:8080/api/orders \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":$((RANDOM%3+1)),\"product_id\":$((RANDOM%3+1)),\"quantity\":$((RANDOM%5+1))}" > /dev/null
  echo "Request $i completed"
  sleep 0.1
done
```

### Generate Traffic for Service Dependencies

Untuk melihat service dependencies di Jaeger, generate traffic yang melakukan cross-service calls:

```bash
# Generate cross-service traffic (rctiplus-api -> user-profile-service)
for i in {1..30}; do
  curl -s http://localhost:8080/api/users-with-profile/1 > /dev/null
  echo "Cross-service request $i completed"
  sleep 0.3
done

# Wait 1-2 minutes, then check Jaeger UI System Architecture tab
```

## Grafana Setup

Grafana has been pre-configured with all necessary data sources and dashboards. Here's how to get started:

### Access Grafana
1. Open Grafana: http://localhost:3000
2. No login required (anonymous access enabled with Admin role)
3. Navigate to **Dashboards** to see pre-loaded dashboards or **Explore** to query data directly

### Available Data Sources
- **Tempo**: Distributed traces with service graph support
- **Prometheus**: Metrics (default) - receives service graph metrics from Tempo
- **Mimir**: Alternative metrics backend
- **Loki**: Logs
- **Jaeger**: Alternative trace backend with built-in service graph

### Quick Navigation
- **Dashboards**: Pre-configured dashboards for infrastructure metrics, RED metrics, and service topology
  - **Service Graph Dashboard**: Visual service dependency map with metrics from Tempo
- **Explore**: Direct querying of data sources
- **Alerting**: Configure alerts based on metrics (if needed)

### Service Graph Features

#### Grafana Service Graph
1. **Generate Cross-Service Traffic** (required first):
   ```bash
   # Generate traffic that calls user-profile-service
   for i in {1..30}; do
     curl -s http://localhost:8080/api/users-with-profile/1 > /dev/null
     sleep 0.3
   done
   ```

2. **Access Service Graph Dashboard**: 
   - Navigate to **Dashboards** → **Service Graph Dashboard**
   - View interactive service map showing dependencies between services
   - Service Map mengambil data langsung dari Tempo (tidak perlu Prometheus metrics)
   - Tunggu 1-2 menit setelah generate traffic untuk Service Map muncul

3. **Service Map in Explore**:
   - Go to **Explore** → Select **Tempo** datasource
   - Use query type **Service Map** to visualize service dependencies
   - Service graph di-generate oleh Tempo metrics generator dari traces

#### Jaeger Service Graph
1. **Access Jaeger UI**: http://localhost:16686
2. **Generate Cross-Service Traffic** (required for service dependencies):
   ```bash
   # Generate traffic that calls user-profile-service
   for i in {1..20}; do
     curl -s http://localhost:8080/api/users-with-profile/1 > /dev/null
     sleep 0.5
   done
   ```
3. **View Service Dependencies**:
   - Wait 1-2 minutes for Jaeger to process dependencies
   - Click on **System Architecture** tab in Jaeger UI
   - Service dependencies will show relationships between services
   - Filter by time range (select last 15-30 minutes)
4. **Verify Services**:
   - Check that both `rctiplus-api` and `user-profile-service` appear in the services list
   - Service dependencies require cross-service calls (e.g., `/api/users-with-profile/{id}`)

### Query Examples

**Tempo (Traces):**
- Search traces by service: `{service.name="rctiplus-api"}`
- Filter by duration: duration > 100ms

**Prometheus (Metrics):**
```promql
# Request rate
rate(http_server_request_count_total[5m])

# Request duration p95
histogram_quantile(0.95, rate(http_server_request_duration_bucket[5m]))

# Active requests
http_server_active_requests
```

**Loki (Logs):**
- Access via Grafana: Navigate to **Explore** → Select "Loki" datasource
- Query examples:
  ```logql
  {service_name="rctiplus-api"}
  {service_name="rctiplus-api"} |= "error"
  {service_name="rctiplus-api"} | json | level="error"
  ```
- Or use LogCLI (install: `brew install logcli` or download from [Loki releases](https://github.com/grafana/loki/releases)):
  ```bash
  logcli query '{service_name="rctiplus-api"}' --addr=http://localhost:3100
  ```

**Note:** Loki doesn't have a built-in web UI. Use Grafana Explore or LogCLI to query logs.

## Correlation Features

Grafana is configured for complete observability correlation:

- **Traces → Logs**: Click on a trace to see related logs
- **Traces → Metrics**: Navigate from traces to related metrics
- **Metrics → Traces**: View exemplar traces from metric queries
- **Service Maps**: Visual representation of service dependencies

## Development

### Local Development (tanpa Docker)

1. Install dependencies:
```bash
cd app
go mod download
```

2. Set environment variables:
```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_SERVICE_NAME=rctiplus-api
export PORT=8080
```

3. Run the app:
```bash
go run main.go
```

### Rebuild App Container

```bash
docker-compose up -d --build app
```

## Project Structure

```
rctiplus/
├── app/                          # Golang application
│   ├── main.go                   # Main application with OTEL
│   ├── go.mod                    # Go dependencies
│   ├── go.sum                    # Go checksums
│   └── Dockerfile                # App container
├── infrastructure/
│   ├── alloy/
│   │   └── config.alloy         # Alloy collector config
│   ├── tempo/
│   │   └── tempo.yaml           # Tempo config
│   ├── mimir/
│   │   ├── mimir.yaml           # Mimir config
│   │   └── prometheus.yml       # Prometheus config
│   ├── loki/
│   │   └── loki.yaml            # Loki config
│   └── grafana/
│       ├── datasources.yml      # Grafana datasources
│       ├── dashboards.yml       # Dashboard provisioning
│       └── dashboards/          # Custom dashboards
├── docker-compose.yml            # Main orchestration
└── README.md                     # This file
```

## Features Implemented

### OpenTelemetry Instrumentation

- **Traces**: Distributed tracing dengan context propagation
- **Metrics**: Custom metrics (counter, histogram, up-down counter)
- **Logs**: Structured logging dengan correlation
- **Automatic Instrumentation**: HTTP server & client
- **Manual Instrumentation**: Custom spans untuk business logic

### Application Features

- REST API dengan multiple endpoints
- In-memory data storage (users, products)
- Order processing dengan validation
- Error handling & logging
- Simulated processing delays untuk demo traces

### Observability

- Full OTLP integration
- Grafana Alloy sebagai collector
- Multi-backend support (Tempo, Mimir, Prometheus, Loki)
- Pre-configured datasources dengan correlation
- Exemplar support

## Troubleshooting

### Services tidak running

```bash
# Check logs
docker-compose logs -f [service_name]

# Restart specific service
docker-compose restart [service_name]
```

### Data tidak muncul di Grafana

1. Check Alloy logs:
```bash
docker-compose logs -f alloy
```

2. Verify OTLP connection:
```bash
docker-compose logs -f app
```

3. Check backend services:
```bash
docker-compose logs -f tempo mimir loki
```

### Reset semua data

```bash
docker-compose down -v
docker-compose up -d
```

### Infrastructure Verification

To verify the entire infrastructure is running correctly, you can use the verification script:

```bash
# Run the verification script
./check_infra_fixed.sh
```

This script will:
- Check the status of all containers
- Test accessibility of all services
- Generate sample traffic to the API
- Provide a summary of the infrastructure status

## Stop & Cleanup

```bash
# Stop services
docker-compose down

# Remove volumes (delete all data)
docker-compose down -v
```

## Performance Tips

- Untuk production, gunakan batch processor dengan size & timeout yang sesuai
- Configure sampling untuk high-traffic services
- Enable compression untuk OTLP export
- Use resource detection untuk automatic attributes

## Kubernetes Deployment

Infrastruktur ini juga dapat di-deploy ke Kubernetes menggunakan Terraform dan Helm.

### Prerequisites
- Kubernetes cluster
- `kubectl` configured
- `terraform` installed
- `helm` installed (optional)

### Quick Start

1. **Deploy dengan Terraform**:
   ```bash
   cd terraform
   terraform init
   terraform plan
   terraform apply
   ```

2. **Deploy Kubernetes Resources**:
   ```bash
   # Menggunakan Helm
   helm install rctiplus ./kubernetes/helm-chart \
     --namespace rctiplus \
     --create-namespace \
     -f ./kubernetes/values.yaml
   ```

3. **Access Services** (port-forward):
   ```bash
   kubectl port-forward -n rctiplus svc/grafana 3000:3000
   kubectl port-forward -n rctiplus svc/jaeger 16686:16686
   ```

Untuk detail lengkap, lihat [kubernetes/README.md](./kubernetes/README.md).

### Service Graph di Kubernetes

Semua konfigurasi service graph sudah termasuk:
- ✅ Tempo metrics generator dengan service-graphs processor
- ✅ Remote write ke Prometheus untuk service graph metrics
- ✅ Grafana Service Graph Dashboard
- ✅ Jaeger service graph built-in

## Learn More

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Grafana Alloy Documentation](https://grafana.com/docs/alloy/latest/)
- [Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [Mimir Documentation](https://grafana.com/docs/mimir/latest/)
- [Loki Documentation](https://grafana.com/docs/loki/latest/)

## License

Demo project - MIT License
