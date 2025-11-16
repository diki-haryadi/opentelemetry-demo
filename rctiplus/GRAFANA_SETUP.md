# Grafana Setup & Infrastructure Configuration Guide

## Table of Contents
1. [Overview](#overview)
2. [Infrastructure Architecture](#infrastructure-architecture)
3. [Grafana Setup](#grafana-setup)
4. [Infrastructure Configuration Analysis](#infrastructure-configuration-analysis)
5. [Verification Steps](#verification-steps)
6. [Troubleshooting](#troubleshooting)

## Overview

This document provides comprehensive instructions for setting up Grafana in the OpenTelemetry demo environment and analyzing the current infrastructure configuration. The demo uses a complete observability stack with distributed tracing, metrics, and logs.

## Infrastructure Architecture

The observability stack consists of:

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
│   Loki   │   │  Mimir/  │ (Metrics)
│  (Logs)  │   │Prometheus│
└──────────┘   └──────────┘
     │              │
     └──────┬───────┘
            │
            v
      ┌──────────┐
      │ Grafana  │ (Visualization)
      └──────────┘
```

### Components

- **App**: Golang REST API microservice with OpenTelemetry instrumentation
- **Alloy**: Grafana Alloy as collector for traces, metrics, and logs
- **Tempo**: Backend for storing traces
- **Mimir/Prometheus**: Backend for storing metrics
- **Loki**: Backend for storing logs
- **Grafana**: Dashboard for visualizing all telemetry data

## Grafana Setup

### 1. Starting the Infrastructure

```bash
# Navigate to project directory
cd rctiplus

# Start all services
docker-compose up -d
```

Wait for all services to start. You can check the status with:
```bash
docker-compose ps
```

### 2. Access Grafana

- **Grafana UI**: http://localhost:3000
- **Username**: No login required (anonymous login enabled)
- **Role**: Admin (pre-configured)

### 3. Data Sources Configuration

Grafana is pre-configured with the following data sources:

#### Tempo (Tracing)
- **Name**: Tempo
- **Type**: Tempo
- **URL**: http://tempo:3200
- **Features**:
  - Traces-to-logs correlation with Loki
  - Traces-to-metrics correlation with Prometheus
  - Service map visualization
  - Node graph visualization

#### Prometheus (Metrics)
- **Name**: Prometheus
- **Type**: Prometheus
- **URL**: http://prometheus:9090
- **Features**:
  - Exemplar trace ID linking to Tempo
  - Default metrics data source

#### Mimir (Metrics)
- **Name**: Mimir
- **Type**: Prometheus
- **URL**: http://mimir:9009/prometheus
- **Features**:
  - Exemplar trace ID linking to Tempo
  - Alternative metrics backend

#### Loki (Logs)
- **Name**: Loki
- **Type**: Loki
- **URL**: http://loki:3100
- **Features**:
  - Derived fields for trace correlation
  - Trace ID extraction from log entries

#### Jaeger (Tracing)
- **Name**: Jaeger
- **Type**: Jaeger
- **URL**: http://jaeger:16686
- **Alternative tracing backend**

### 4. Dashboard Setup

Grafana is pre-configured with custom dashboards:

- **Infrastructure Metrics**: System and container metrics
- **RED Metrics**: Request rate, error rate, and duration metrics
- **Service Tree**: Service dependency visualization

### 5. Using Grafana

#### Exploring Data

1. **Traces**: Navigate to **Explore** → Select "Tempo" → Search for traces using query like `{service.name="rctiplus-api"}`
2. **Metrics**: Navigate to **Explore** → Select "Prometheus" or "Mimir" → Run PromQL queries
3. **Logs**: Navigate to **Explore** → Select "Loki" → Search for logs using LogQL

#### Query Examples

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
```logql
{service_name="rctiplus-api"}
```

## Infrastructure Configuration Analysis

### 1. Grafana Configuration

**File**: `infrastructure/grafana/datasources.yml`
- **Status**: ✅ **CORRECT** - Properly configured with all necessary data sources
- **Correlation**: All data sources have proper linking for trace/log/metric correlation
- **Features**: Trace-to-logs, trace-to-metrics, exemplars are correctly configured

**File**: `infrastructure/grafana/dashboards.yml`
- **Status**: ✅ **CORRECT** - Properly configured to load dashboards from the dashboards directory
- **Functionality**: Auto-updates dashboards every 10 seconds

### 2. Alloy Configuration

**File**: `infrastructure/alloy/config.alloy`
- **Status**: ✅ **CORRECT** - Comprehensive collector configuration
- **OTLP Receiver**: Properly configured for both gRPC and HTTP protocols
- **Batch Processor**: Properly configured for performance optimization
- **Exporters**: Correctly configured to forward data to appropriate backends
- **Service Graphs**: Enabled for metrics generator in Tempo

### 3. Backend Services Configuration

**Tempo** (`infrastructure/tempo/tempo.yaml`):
- ✅ **CORRECT** - Properly configured with OTLP receivers
- ✅ **CORRECT** - Metrics generator enabled with service graphs and span metrics
- ✅ **CORRECT** - Storage configured for local development

**Loki** (`infrastructure/loki/loki.yaml`):
- ✅ **CORRECT** - HTTP port properly configured
- ✅ **CORRECT** - Storage paths configured
- ✅ **CORRECT** - Schema configuration for logs storage

**Mimir** (`infrastructure/mimir/mimir.yaml`):
- ✅ **CORRECT** - Storage properly configured
- ✅ **CORRECT** - Multi-tenancy disabled for single-tenant usage
- ✅ **CORRECT** - HTTP port properly configured

**Prometheus** (`infrastructure/mimir/prometheus.yml`):
- ✅ **CORRECT** - Basic configuration with Alloy integration
- ✅ **CORRECT** - Scraping configuration for services

### 4. Docker Compose Configuration

**File**: `docker-compose.yml`
- ✅ **CORRECT** - Proper networking configuration
- ✅ **CORRECT** - Dependencies correctly set up
- ✅ **CORRECT** - Port mappings appropriate
- ✅ **CORRECT** - Volume mounts for persistent data

### 5. Configuration Quality Assessment

**Strengths:**
1. **Complete OTLP Pipeline**: Proper end-to-end pipeline from application to visualization
2. **Correlation Features**: Full trace-log-metric correlation configured
3. **Multiple Backend Options**: Both Mimir and Prometheus for metrics redundancy
4. **Service Mesh Features**: Service graphs and topology visualization
5. **Auto-configuration**: Data sources and dashboards automatically provisioned
6. **Development-Friendly**: Anonymous login, easy startup

**Considerations:**
1. **Security**: Anonymous login enabled (appropriate for demo, not production)
2. **Storage**: Local filesystem storage (appropriate for demo, not production)
3. **Replication**: Single instance setup (appropriate for demo, not production)

## Verification Steps

### 1. Service Health Check
```bash
# Check all services are running
docker-compose ps

# Verify specific services
docker-compose logs grafana | head -10
docker-compose logs alloy | head -10
docker-compose logs tempo | head -10
docker-compose logs loki | head -10
docker-compose logs mimir | head -10
```

### 2. Initialization Time Considerations
⚠️ **Note**: Some services (especially Tempo and Loki) may take 1-2 minutes to fully initialize after starting. If services aren't immediately accessible, wait a minute and try again.

```bash
# Wait for services to be fully initialized
sleep 60
# Then check availability again
```

### 3. Grafana Data Source Verification
1. Open Grafana at http://localhost:3000
2. Navigate to **Configuration** → **Data Sources**
3. Verify all 5 data sources are present and working:
   - Tempo (should show "Data source is working")
   - Prometheus (should show "Data source is working")
   - Mimir (should show "Data source is working")
   - Loki (should show "Data source is working")
   - Jaeger (should show "Data source is working")

### 3. Generate Test Data
```bash
# Generate some traffic to create traces/metrics/logs
curl http://localhost:8080/api/users
curl http://localhost:8080/api/products
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1,"product_id": 1,"quantity": 1}'
```

### 4. Verify Data Flow
1. **Traces in Tempo**: Go to Explore → Tempo → Search for service `rctiplus-api`
2. **Metrics in Prometheus/Mimir**: Go to Explore → Prometheus → Run `http_server_request_count_total`
3. **Logs in Loki**: Go to Explore → Loki → Query `{service_name="rctiplus-api"}`

### 5. Dashboard Verification
1. Navigate to **Dashboards** in Grafana
2. Verify custom dashboards are loaded:
   - Infrastructure Metrics
   - RED Metrics
   - Service Tree
3. Check if dashboards display data correctly

## Troubleshooting

### Common Issues

#### Grafana Data Sources Not Working
```bash
# Check for configuration errors
docker-compose logs grafana

# Check if backend services are accessible
docker-compose exec grafana curl -v http://tempo:3200
docker-compose exec grafana curl -v http://prometheus:9090
docker-compose exec grafana curl -v http://loki:3100
docker-compose exec grafana curl -v http://mimir:9009
```

#### No Traces/Metrics/Logs Appearing
```bash
# Check Alloy collector
docker-compose logs -f alloy

# Check application logs
docker-compose logs -f app

# Verify OTLP export
docker-compose exec app env | grep OTEL
```

#### Service Dependencies Not Starting
```bash
# Check dependencies manually
docker-compose up -d alloy
docker-compose up -d tempo
docker-compose up -d loki
docker-compose up -d mimir
docker-compose up -d grafana
```

#### Reset All Data
```bash
# Stop and remove all containers and volumes
docker-compose down -v
docker-compose up -d
```

### Performance and Scaling Considerations

For production usage, consider:

1. **Authentication**: Implement proper authentication instead of anonymous access
2. **Storage**: Use cloud storage solutions instead of local filesystem
3. **High Availability**: Deploy multiple instances with replication
4. **Resource Limits**: Set appropriate CPU and memory limits
5. **SSL/TLS**: Enable encryption for all communication
6. **Monitoring**: Monitor the observability stack itself

## Additional Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Grafana Alloy Documentation](https://grafana.com/docs/alloy/latest/)
- [Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [Mimir Documentation](https://grafana.com/docs/mimir/latest/)
- [Loki Documentation](https://grafana.com/docs/loki/latest/)