# Infrastructure Setup

Infrastructure setup untuk RCTI Plus OpenTelemetry Demo, diadaptasi dari [OpenTelemetry Demo](https://github.com/open-telemetry/opentelemetry-demo).

## Perbedaan dengan Demo Original

### Demo Original
- **Tracing**: Jaeger dengan `dependencyGraph` query
- **Metrics**: OpenTelemetry Collector dengan `spanmetrics` connector
- **Logs**: OpenSearch
- **Collector**: OpenTelemetry Collector

### RCTI Plus Setup
- **Tracing**: Tempo dengan `serviceMap` query
- **Metrics**: Tempo Metrics Generator dengan `service-graphs` processor
- **Logs**: Loki
- **Collector**: Grafana Alloy

## Komponen Infrastructure

### 1. Tempo (Tracing Backend)
- **Port**: 3200 (HTTP), 4317 (OTLP gRPC), 4318 (OTLP HTTP)
- **Config**: `tempo/tempo.yaml`
- **Features**:
  - Metrics Generator dengan `service-graphs` processor
  - Remote write ke Prometheus dan Mimir
  - Service graph metrics generation

### 2. Prometheus (Metrics Backend)
- **Port**: 9090
- **Config**: `mimir/prometheus.yml`
- **Features**:
  - Remote write receiver enabled
  - Scraping Tempo metrics
  - Scraping application metrics

### 3. Mimir (Metrics Backend - Alternative)
- **Port**: 9009
- **Config**: `mimir/mimir.yaml`
- **Features**:
  - Prometheus-compatible API
  - Long-term storage
  - Remote write receiver

### 4. Loki (Logs Backend)
- **Port**: 3100
- **Config**: `loki/loki.yaml`
- **Features**:
  - Log aggregation
  - Integration dengan Tempo untuk trace-to-logs

### 5. Grafana Alloy (Collector)
- **Port**: 12345 (UI), 4317 (OTLP gRPC), 4318 (OTLP HTTP)
- **Config**: `alloy/config.alloy`
- **Features**:
  - OTLP receiver untuk traces, metrics, logs
  - Forward ke Tempo, Prometheus/Mimir, Loki
  - Self-monitoring

### 6. Grafana (Visualization)
- **Port**: 3000
- **Config**: `grafana/datasources.yml`, `grafana/dashboards.yml`
- **Features**:
  - Tempo datasource dengan serviceMap
  - Prometheus/Mimir datasources
  - Loki datasource
  - Service Graph dashboard

## Service Graph Setup

### 1. Tempo Configuration
Tempo dikonfigurasi untuk generate service graph metrics:

```yaml
metrics_generator:
  processors: [service-graphs, span-metrics, local-blocks]
  remote_write:
    - url: http://prometheus:9090/api/v1/write
    - url: http://mimir:9009/api/v1/push
```

### 2. Grafana Datasource
Tempo datasource dikonfigurasi dengan serviceMap:

```yaml
serviceMap:
  datasourceUid: 'mimir'  # atau 'prometheus'
nodeGraph:
  enabled: true
```

### 3. Dashboard
Service Graph dashboard menggunakan:
- **Query Type**: `serviceMap`
- **Panel Type**: `nodeGraph`
- **Datasource**: Tempo

## Metrics yang Dihasilkan

Tempo metrics generator menghasilkan metrics berikut:

### Service Graph Metrics
- `traces_service_graph_request_total` - Total requests antar services
- `traces_service_graph_request_failed_total` - Failed requests
- `traces_service_graph_request_server_seconds_bucket` - Latency histogram

### Span Metrics
- `traces_spanmetrics_calls_total` - Total span calls
- `traces_spanmetrics_latency_bucket` - Span latency histogram

## Verifikasi Setup

### 1. Check Tempo Metrics
```bash
curl http://localhost:3200/metrics | grep traces_service_graph
```

### 2. Check Prometheus Metrics
```bash
curl http://localhost:9090/api/v1/query?query=traces_service_graph_request_total
```

### 3. Check Grafana Service Graph
1. Buka Grafana: http://localhost:3000
2. Navigate ke **Explore** → **Tempo**
3. Pilih query type: **Service Graph**
4. Klik **Run Query**

### 4. Check Dashboard
1. Buka dashboard: **Service Graph Dashboard**
2. Service graph harus muncul di panel pertama
3. Metrics panels harus menampilkan data

## Troubleshooting

### Service Graph Tidak Muncul

1. **Check Tempo Metrics Generator**
   ```bash
   docker logs tempo | grep "metrics_generator"
   ```

2. **Check Remote Write**
   ```bash
   docker logs tempo | grep "remote_write"
   ```

3. **Check Prometheus Metrics**
   ```bash
   curl http://localhost:9090/api/v1/query?query=traces_service_graph_request_total
   ```

4. **Check Grafana Datasource**
   - Pastikan `serviceMap.datasourceUid` mengarah ke Prometheus atau Mimir
   - Pastikan datasource UID sesuai

5. **Generate Traffic**
   ```bash
   curl http://localhost:8080/api/users-with-profile/1
   ```

### Metrics Tidak Muncul di Prometheus

1. **Check Prometheus Remote Write Receiver**
   ```bash
   docker logs prometheus | grep "remote_write"
   ```

2. **Check Tempo Remote Write**
   ```bash
   docker logs tempo | grep "remote_write"
   ```

3. **Check Network Connectivity**
   ```bash
   docker exec tempo ping prometheus
   ```

## Best Practices dari Demo

1. **Metrics Export Interval**: 15 detik (sesuai dengan Prometheus scrape interval)
2. **Trace Retention**: 1 jam untuk demo (bisa disesuaikan)
3. **Service Graph**: Generate dari traces menggunakan processor
4. **Exemplars**: Enable untuk correlation antara metrics dan traces

## Migration dari Demo

Jika ingin migrate dari demo original:

1. **Replace Jaeger dengan Tempo**
   - Update datasource dari Jaeger ke Tempo
   - Update query type dari `dependencyGraph` ke `serviceMap`

2. **Replace OTel Collector dengan Alloy**
   - Alloy sudah support spanmetrics connector
   - Atau gunakan Tempo metrics generator

3. **Replace OpenSearch dengan Loki**
   - Update datasource configuration
   - Update trace-to-logs configuration

## References

- [Tempo Service Graph Documentation](https://grafana.com/docs/tempo/latest/metrics-generator/service_graphs/)
- [Grafana Service Graph Documentation](https://grafana.com/docs/grafana/latest/datasources/tempo/service-graph/)
- [OpenTelemetry Demo](https://github.com/open-telemetry/opentelemetry-demo)

