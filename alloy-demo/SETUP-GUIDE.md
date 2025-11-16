# Setup Guide - Alloy Demo

## 📁 Project Structure

```
alloy-demo/
├── apps/                          # Golang applications
│   ├── frontend/                  # Frontend service (port 8080)
│   │   ├── main.go               # Application code with OTel instrumentation
│   │   ├── go.mod                # Go dependencies
│   │   └── Dockerfile            # Multi-stage build
│   ├── backend/                   # Backend service (port 8081)
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── payment/                   # Payment service (port 8082)
│       ├── main.go
│       ├── go.mod
│       └── Dockerfile
├── config/                        # Configuration files
│   ├── alloy/
│   │   └── config.alloy          # Alloy collector configuration
│   ├── grafana/
│   │   ├── provisioning/
│   │   │   ├── datasources/      # Prometheus, Tempo, Loki datasources
│   │   │   └── dashboards/       # Dashboard provisioning
│   │   └── dashboards/
│   │       ├── service-graph.json       # Service graph dashboard
│   │       └── application-metrics.json # RED metrics dashboard
│   ├── tempo.yaml                # Tempo configuration with service graphs
│   ├── loki.yaml                 # Loki configuration
│   └── prometheus.yml            # Prometheus scrape config
├── docker-compose.yml            # Complete stack definition
├── start.sh                      # Quick start script
├── load-generator.sh             # Traffic generator
└── README.md                     # Full documentation
```

## 🚀 Quick Start (3 Steps)

### 1. Start the Stack

```bash
cd alloy-demo
./start.sh
```

Or manually:
```bash
docker-compose up -d
```

### 2. Generate Traffic

```bash
./load-generator.sh
```

### 3. View Dashboards

Open Grafana: http://localhost:3000

- **Service Graph Dashboard**: http://localhost:3000/d/service-graph
- **Application Metrics**: http://localhost:3000/d/app-metrics

## 🏗️ Architecture Components

### Applications (Golang + OpenTelemetry)

All three services are instrumented with:
- **OTLP gRPC exporter** → sends to Alloy
- **Automatic HTTP instrumentation** via `otelhttp`
- **Custom metrics**: `calls_total`, `duration_milliseconds`
- **Trace context propagation** (W3C Trace Context)
- **Resource attributes**: service name, version, namespace

**Service Flow:**
```
Frontend (8080)
  ├─→ Backend (8081)
  │     └─→ Payment (8082)
  └─→ Payment (8082)
```

### Grafana Alloy

**Purpose**: Unified telemetry collector

**Configuration** (`config/alloy/config.alloy`):
- OTLP receivers (gRPC :4317, HTTP :4318)
- Batch processor for efficiency
- **Span metrics connector** - generates RED metrics from traces
- **Service graph connector** - creates service dependency metrics
- Exporters:
  - Traces → Tempo (OTLP)
  - Metrics → Prometheus (remote write)
  - Logs → Loki (push API)

### Tempo

**Purpose**: Distributed tracing backend

**Key Features**:
- OTLP trace ingestion
- **Metrics generator** with service graphs enabled
- Generates metrics from traces:
  - `traces_service_graph_request_total`
  - `traces_service_graph_request_server_seconds`
  - `traces_service_graph_request_failed_total`
- Local storage backend

### Prometheus

**Purpose**: Metrics storage and querying

**Scrapes**:
- Alloy metrics
- Tempo metrics (including service graphs)
- Loki metrics
- Self-monitoring

**Features**:
- Exemplar storage enabled (links metrics to traces)
- Native histograms support

### Loki

**Purpose**: Log aggregation

**Features**:
- Receives logs from Alloy
- Trace correlation via derived fields
- TSDB schema for efficient storage

### Grafana

**Purpose**: Visualization platform

**Pre-configured**:
- 3 datasources (Prometheus, Tempo, Loki)
- 2 dashboards (Service Graph, Application Metrics)
- Trace-to-logs correlation
- Metrics-to-traces exemplars
- Anonymous access enabled (no login required)

## 📊 Dashboards Explained

### Service Graph Dashboard

**Panels**:
1. **Service Graph (Node Graph)** - Visual dependency map
   - Nodes = services
   - Edges = requests between services
   - Edge thickness = request rate
   - Color = latency/errors

2. **Request Rate by Service** - Time series of RPS per service

3. **P95 Latency by Service** - 95th percentile latency trends

4. **Error Rate by Service** - Error percentage over time

5. **Service Dependencies Table** - Matrix of client→server calls

**Metrics Used**:
- `traces_service_graph_request_total`
- `traces_service_graph_request_server_seconds_bucket`
- `traces_service_graph_request_failed_total`

### Application Metrics Dashboard

**Panels**:
1. **Total Request Rate** - System-wide throughput (stat)
2. **P95 Latency** - Overall latency (stat)
3. **Error Rate** - System error percentage (stat)
4. **Active Services** - Service count (stat)
5. **Request Rate by Service** - Per-service RPS (time series)
6. **Latency Percentiles** - p50, p95, p99 by service (time series)
7. **HTTP Status Codes** - Status distribution (stacked area)

**Metrics Used**:
- `calls_total` - Custom counter from apps
- `duration_milliseconds_bucket` - Custom histogram from apps

## 🔧 Configuration Details

### Alloy Connectors

**Span Metrics Connector**:
```hcl
otelcol.connector.spanmetrics "default" {
  histogram {
    explicit {
      buckets = [0.001, 0.002, 0.005, ..., 60.0]
    }
  }
  dimensions {
    name = "http.method"
  }
  dimensions {
    name = "http.status_code"
  }
}
```

**Service Graph Connector**:
```hcl
otelcol.connector.servicegraph "default" {
  latency_histogram_buckets = [0.001, ..., 60.0]
  dimensions = ["http.method", "http.status_code"]
  store {
    ttl = "2s"
    max_items = 10000
  }
}
```

### Tempo Metrics Generator

```yaml
metrics_generator:
  processor:
    service_graphs:
      dimensions:
        - http.method
        - http.status_code
      histogram_buckets: [0.1, 0.2, 0.4, ..., 12.8]
    span_metrics:
      dimensions:
        - http.method
        - http.status_code
        - http.route
```

## 🧪 Testing

### Manual Testing

```bash
# Home page
curl http://localhost:8080/

# Products (frontend → backend → payment)
curl http://localhost:8080/products

# Checkout (frontend → payment)
curl -X POST http://localhost:8080/checkout

# Health checks
curl http://localhost:8081/health
curl http://localhost:8082/health
```

### Load Testing

The `load-generator.sh` script sends:
- 50% GET /products
- 30% POST /checkout
- 20% GET /

Random delays between 100-500ms simulate realistic traffic.

## 🔍 Troubleshooting

### No Service Graph?

**Cause**: Not enough traffic or data not yet aggregated

**Solution**:
1. Run load generator for 2-3 minutes
2. Check Tempo is generating metrics:
   ```bash
   curl http://localhost:3200/metrics | grep service_graph
   ```
3. Verify Prometheus is scraping Tempo:
   ```
   http://localhost:9090/targets
   ```

### Applications Not Sending Data?

**Check Alloy**:
```bash
docker-compose logs alloy
curl http://localhost:12345/metrics
```

**Check app logs**:
```bash
docker-compose logs frontend
docker-compose logs backend
docker-compose logs payment
```

### Grafana Shows "No Data"?

1. Wait 2-3 minutes after startup
2. Generate traffic with load generator
3. Check time range in dashboard (last 15 minutes)
4. Verify datasources in Grafana settings

## 📈 Metrics Reference

### Application Metrics (Custom)

| Metric | Type | Description |
|--------|------|-------------|
| `calls_total` | Counter | Total HTTP requests |
| `duration_milliseconds` | Histogram | Request duration |
| `payment_amount` | Histogram | Payment amounts (payment service only) |

**Labels**: `service_name`, `http_method`, `http_route`, `status_code`

### Service Graph Metrics (Generated by Tempo)

| Metric | Type | Description |
|--------|------|-------------|
| `traces_service_graph_request_total` | Counter | Requests between services |
| `traces_service_graph_request_server_seconds` | Histogram | Latency between services |
| `traces_service_graph_request_failed_total` | Counter | Failed requests |

**Labels**: `client`, `server`, `connection_type`

## 🛠️ Customization

### Add New Service

1. Create app in `apps/new-service/`
2. Add OpenTelemetry instrumentation
3. Set `OTEL_EXPORTER_OTLP_ENDPOINT=http://alloy:4317`
4. Add to `docker-compose.yml`
5. Rebuild: `docker-compose up -d --build`

### Modify Dashboards

Edit JSON files in `config/grafana/dashboards/` or use Grafana UI.

### Change Retention

**Tempo**: Edit `config/tempo.yaml` → `block_retention`
**Loki**: Edit `config/loki.yaml` → `reject_old_samples_max_age`
**Prometheus**: Add to `docker-compose.yml` → `--storage.tsdb.retention.time=7d`

## 📚 Resources

- [Grafana Alloy Docs](https://grafana.com/docs/alloy/)
- [Tempo Service Graphs](https://grafana.com/docs/tempo/latest/metrics-generator/service-graphs/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [PromQL Guide](https://prometheus.io/docs/prometheus/latest/querying/basics/)

## 🎯 Next Steps

1. ✅ Start the stack: `./start.sh`
2. ✅ Generate traffic: `./load-generator.sh`
3. ✅ Explore Service Graph dashboard
4. ✅ View traces in Grafana → Explore → Tempo
5. ✅ Query logs in Grafana → Explore → Loki
6. ✅ Create custom dashboards
7. ✅ Add your own services

---

**Note on Lint Errors**: The IDE shows import errors for the Go applications because dependencies aren't downloaded locally. These will be resolved automatically when Docker builds the images using `go mod download && go mod tidy` in the Dockerfiles.
