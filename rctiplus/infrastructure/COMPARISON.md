# Perbandingan Infrastructure: Demo vs RCTI Plus

## Overview

Dokumen ini membandingkan setup infrastructure antara OpenTelemetry Demo original dengan RCTI Plus setup.

## Stack Comparison

| Komponen | Demo Original | RCTI Plus | Catatan |
|----------|---------------|-----------|---------|
| **Tracing** | Jaeger | Tempo | Tempo lebih modern dan scalable |
| **Metrics** | Prometheus | Prometheus + Mimir | Mimir untuk long-term storage |
| **Logs** | OpenSearch | Loki | Loki lebih ringan dan terintegrasi dengan Grafana |
| **Collector** | OTel Collector | Grafana Alloy | Alloy adalah fork OTel Collector dengan Prometheus pipelines |
| **Service Graph** | Jaeger dependencyGraph | Tempo serviceMap | Keduanya menghasilkan service graph |

## Service Graph Implementation

### Demo Original (Jaeger)

**Konfigurasi:**
```yaml
# Grafana Datasource
- name: Jaeger
  type: jaeger
  jsonData:
    tracesToLogsV2:
      datasourceUid: webstore-logs
```

**Dashboard Query:**
```json
{
  "queryType": "dependencyGraph",
  "datasource": {
    "type": "jaeger",
    "uid": "webstore-traces"
  }
}
```

**Cara Kerja:**
- Jaeger secara otomatis generate service graph dari traces
- Menggunakan internal dependency graph algorithm
- Tidak perlu metrics generator terpisah

### RCTI Plus (Tempo)

**Konfigurasi:**
```yaml
# Tempo Config
metrics_generator:
  processors: [service-graphs, span-metrics, local-blocks]
  remote_write:
    - url: http://prometheus:9090/api/v1/write
    - url: http://mimir:9009/api/v1/push

# Grafana Datasource
- name: Tempo
  jsonData:
    serviceMap:
      datasourceUid: 'mimir'
    nodeGraph:
      enabled: true
```

**Dashboard Query:**
```json
{
  "queryType": "serviceMap",
  "datasource": {
    "type": "tempo",
    "uid": "tempo"
  }
}
```

**Cara Kerja:**
- Tempo metrics generator memproses traces
- Generate service graph metrics ke Prometheus/Mimir
- Grafana membaca metrics dari Prometheus/Mimir untuk visualisasi

## Metrics Generation

### Demo Original

**OTel Collector Config:**
```yaml
connectors:
  spanmetrics:

service:
  pipelines:
    traces:
      processors: [spanmetrics]
      exporters: [otlp, spanmetrics]
    metrics:
      receivers: [spanmetrics]
      exporters: [otlphttp/prometheus]
```

**Metrics yang Dihasilkan:**
- `traces_span_metrics_calls_total`
- `traces_span_metrics_duration_milliseconds_bucket`
- Service graph dari Jaeger internal

### RCTI Plus

**Tempo Config:**
```yaml
metrics_generator:
  processors: [service-graphs, span-metrics, local-blocks]
  remote_write:
    - url: http://prometheus:9090/api/v1/write
```

**Metrics yang Dihasilkan:**
- `traces_service_graph_request_total`
- `traces_service_graph_request_failed_total`
- `traces_service_graph_request_server_seconds_bucket`
- `traces_spanmetrics_calls_total`
- `traces_spanmetrics_latency_bucket`

## Keuntungan Masing-masing Approach

### Demo Original (Jaeger)

✅ **Pros:**
- Service graph langsung dari Jaeger, tidak perlu metrics generator
- Lebih sederhana untuk setup basic
- Dependency graph sudah built-in

❌ **Cons:**
- Jaeger kurang scalable untuk production besar
- Tidak ada long-term storage built-in
- Service graph tidak bisa di-query sebagai metrics

### RCTI Plus (Tempo)

✅ **Pros:**
- Tempo lebih scalable dan performant
- Service graph sebagai metrics, bisa di-query dan alert
- Long-term storage dengan Mimir
- Terintegrasi dengan Grafana ecosystem (LGTM stack)
- Bisa generate multiple metrics dari traces

❌ **Cons:**
- Setup sedikit lebih kompleks (perlu metrics generator)
- Perlu Prometheus/Mimir untuk service graph

## Migration Guide

Jika ingin migrate dari Demo ke RCTI Plus setup:

### 1. Update Datasources

**Dari:**
```yaml
- name: Jaeger
  type: jaeger
  jsonData:
    tracesToLogsV2:
      datasourceUid: webstore-logs
```

**Ke:**
```yaml
- name: Tempo
  type: tempo
  jsonData:
    serviceMap:
      datasourceUid: 'mimir'
    tracesToLogs:
      datasourceUid: 'loki'
```

### 2. Update Dashboard Queries

**Dari:**
```json
{
  "queryType": "dependencyGraph",
  "datasource": {"type": "jaeger"}
}
```

**Ke:**
```json
{
  "queryType": "serviceMap",
  "datasource": {"type": "tempo"}
}
```

### 3. Update Metrics Queries

**Dari:**
```promql
traces_span_metrics_calls_total
```

**Ke:**
```promql
traces_service_graph_request_total
# atau
traces_spanmetrics_calls_total
```

## Best Practices

### Untuk Demo/Development
- Gunakan setup sederhana (Jaeger + OTel Collector)
- Focus pada functionality, bukan scalability

### Untuk Production
- Gunakan Tempo + Mimir untuk scalability
- Enable service graph metrics untuk observability
- Setup long-term retention untuk compliance

## References

- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [Grafana Service Graph](https://grafana.com/docs/grafana/latest/datasources/tempo/service-graph/)
- [OpenTelemetry Demo](https://github.com/open-telemetry/opentelemetry-demo)

