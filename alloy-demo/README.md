# Alloy Demo - Observability Stack

Complete observability stack using Grafana Alloy, Tempo, Loki, Prometheus, and Grafana with service graph visualization.

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

### Applications (Golang)
- **Frontend** (`:8080`) - Main entry point, calls backend and payment services
- **Backend** (`:8081`) - Product catalog service, calls payment for pricing
- **Payment** (`:8082`) - Payment processing service

### Zero-Code Instrumentation (NEW! 🎉)
- **Beyla Frontend** - eBPF-based auto-instrumentation for frontend
- **Beyla Backend** - eBPF-based auto-instrumentation for backend
- **Beyla Payment** - eBPF-based auto-instrumentation for payment

### Observability Stack
- **Grafana Alloy** (`:12345`) - Collects traces, metrics, and logs from applications
- **Tempo** (`:3200`) - Distributed tracing backend with service graph generation
- **Loki** (`:3100`) - Log aggregation system
- **Prometheus** (`:9090`) - Metrics storage and querying
- **Grafana** (`:3000`) - Visualization and dashboards

## Features

✅ **Service Graph Dashboard** - Visualize service dependencies and performance  
✅ **Distributed Tracing** - End-to-end request tracing across services  
✅ **Metrics Collection** - RED metrics (Rate, Errors, Duration)  
✅ **Log Aggregation** - Centralized logging with trace correlation  
✅ **Exemplars** - Link metrics to traces for deep debugging  
✅ **Zero-Code Instrumentation** - Grafana Beyla eBPF auto-instrumentation (NEW!)  
✅ **Dual Instrumentation** - Manual (otel-lib) + Auto (Beyla) for complete coverage  

## Quick Start

### Prerequisites
- Docker and Docker Compose
- 8GB+ RAM recommended

### Start the Stack

```bash
cd alloy-demo
docker-compose up -d
```

### Access the Services

- **Grafana**: http://localhost:3000
  - Service Graph Dashboard: http://localhost:3000/d/service-graph
  - Application Metrics: http://localhost:3000/d/app-metrics
- **Alloy UI**: http://localhost:12345
- **Prometheus**: http://localhost:9090
- **Frontend App**: http://localhost:8080

### Generate Traffic

Use the included load generator script:

```bash
./load-generator.sh
```

Or manually test endpoints:

```bash
# Home page
curl http://localhost:8080/

# Get products (calls backend -> payment)
curl http://localhost:8080/products

# Checkout (calls payment)
curl http://localhost:8080/checkout
```

## Dashboards

### Service Graph Dashboard
- **Node Graph**: Visual representation of service dependencies
- **Request Rate**: Requests per second by service
- **P95 Latency**: 95th percentile latency by service
- **Error Rate**: Error percentage by service
- **Service Dependencies**: Matrix view of service calls

### Application Metrics Dashboard
- **Total Request Rate**: Overall system throughput
- **P95 Latency**: System-wide latency
- **Error Rate**: System-wide error percentage
- **Active Services**: Number of running services
- **Request Rate by Service**: Per-service throughput
- **Latency Percentiles**: p50, p95, p99 by service
- **HTTP Status Codes**: Status code distribution

## Configuration Files

### Alloy Configuration
- `config/alloy/config.alloy` - Main Alloy configuration
  - OTLP receivers (gRPC and HTTP)
  - Batch processing
  - Span metrics connector for service graphs
  - Service graph connector
  - Exporters to Tempo, Prometheus, and Loki

### Tempo Configuration
- `config/tempo.yaml` - Tempo configuration
  - OTLP receivers
  - Metrics generator with service graphs
  - Span metrics generation
  - Local storage backend

### Grafana Configuration
- `config/grafana/provisioning/datasources/` - Data source configurations
  - Prometheus with exemplar support
  - Tempo with trace-to-logs correlation
  - Loki with derived fields for traces
- `config/grafana/dashboards/` - Pre-built dashboards
  - Service Graph Dashboard
  - Application Metrics Dashboard

## Application Instrumentation

All Golang applications use OpenTelemetry SDK with:
- **Traces**: Automatic HTTP instrumentation with `otelhttp`
- **Metrics**: Custom metrics for request count and duration
- **Propagation**: W3C Trace Context and Baggage
- **Resource Attributes**: Service name, version, namespace

### Key Metrics

- `calls_total` - Total HTTP requests counter
- `duration_milliseconds` - Request duration histogram
- `traces_service_graph_request_total` - Service graph request counter
- `traces_service_graph_request_server_seconds` - Service graph latency histogram
- `traces_service_graph_request_failed_total` - Service graph error counter

## Monitoring

### View Traces
1. Open Grafana at http://localhost:3000
2. Navigate to Explore
3. Select "Tempo" data source
4. Search for traces or use TraceQL queries

### View Metrics
1. Open Grafana at http://localhost:3000
2. Navigate to Explore
3. Select "Prometheus" data source
4. Query metrics like `rate(calls_total[5m])`

### View Logs
1. Open Grafana at http://localhost:3000
2. Navigate to Explore
3. Select "Loki" data source
4. Use LogQL queries like `{service_name="frontend"}`

## Troubleshooting

### Check Service Health

```bash
# Check all containers
docker-compose ps

# View logs
docker-compose logs -f [service-name]

# Check Alloy status
curl http://localhost:12345/metrics
```

### No Data in Grafana?

1. Wait 2-3 minutes for metrics to populate
2. Generate traffic using the load generator
3. Check Alloy logs: `docker-compose logs alloy`
4. Verify data sources in Grafana settings

### Service Graph Not Showing?

1. Ensure traffic is flowing between services
2. Check Tempo metrics generator is enabled
3. Verify Prometheus is scraping Tempo metrics
4. Wait for data to accumulate (2-5 minutes)

## Stopping the Stack

```bash
docker-compose down

# Remove volumes (clean slate)
docker-compose down -v
```

## Architecture Details

### Data Flow

1. **Applications** send OTLP data (traces, metrics, logs) to **Alloy**
2. **Alloy** processes and routes data:
   - Traces → Tempo (also generates span metrics)
   - Metrics → Prometheus
   - Logs → Loki
3. **Tempo** generates service graph metrics from traces
4. **Prometheus** stores all metrics including service graphs
5. **Grafana** queries all backends for visualization

### Service Graph Generation

Service graphs are generated from distributed traces using two methods:

1. **Alloy Span Metrics Connector**: Generates RED metrics from spans
2. **Tempo Metrics Generator**: Creates service graph metrics with client/server relationships

Both methods produce metrics that power the service graph visualization in Grafana.

## Customization

### Add More Services

1. Create new Go application in `apps/[service-name]/`
2. Add OpenTelemetry instrumentation
3. Add service to `docker-compose.yml`
4. Configure OTLP endpoint to Alloy

### Modify Dashboards

Dashboards are in `config/grafana/dashboards/` and can be edited:
- Directly in JSON files
- Through Grafana UI (changes persist in volume)

### Adjust Retention

Edit configuration files:
- **Tempo**: `config/tempo.yaml` - `block_retention`
- **Loki**: `config/loki.yaml` - `reject_old_samples_max_age`
- **Prometheus**: `config/prometheus.yml` - add `--storage.tsdb.retention.time`

## Resources

- [Grafana Alloy Documentation](https://grafana.com/docs/alloy/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Service Graph View](https://grafana.com/docs/tempo/latest/metrics-generator/service-graphs/)
