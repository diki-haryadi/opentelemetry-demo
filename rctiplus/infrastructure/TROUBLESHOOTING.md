# Troubleshooting Service Graph

## Error: "No service graph data found"

Error ini muncul ketika Grafana tidak menemukan service graph metrics di Prometheus/Mimir.

## Langkah Troubleshooting

### 1. Verifikasi Tempo Metrics Generator

**Check Tempo logs:**
```bash
docker logs tempo | grep -i "metrics_generator\|service_graph"
```

**Check Tempo metrics endpoint:**
```bash
curl http://localhost:3200/metrics | grep traces_service_graph
```

**Expected output:**
```
traces_service_graph_request_total{client="rctiplus-api",server="user-profile-service"} 10
traces_service_graph_request_failed_total{client="rctiplus-api",server="user-profile-service"} 0
```

### 2. Verifikasi Remote Write ke Prometheus

**Check Tempo logs untuk remote write:**
```bash
docker logs tempo | grep -i "remote_write\|prometheus"
```

**Check apakah Prometheus menerima data:**
```bash
curl http://localhost:9090/api/v1/query?query=traces_service_graph_request_total
```

**Expected output:**
```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "client": "rctiplus-api",
          "server": "user-profile-service"
        },
        "value": [1234567890, "10"]
      }
    ]
  }
}
```

### 3. Verifikasi Traffic dan Traces

**Generate traffic:**
```bash
# Generate cross-service traffic
curl http://localhost:8080/api/users-with-profile/1

# Tunggu beberapa detik untuk Tempo memproses
sleep 10
```

**Check traces di Tempo:**
```bash
curl http://localhost:3200/api/search?limit=10
```

**Check traces di Grafana:**
1. Buka Grafana: http://localhost:3000
2. Navigate ke **Explore** → **Tempo**
3. Query type: **Search** atau **TraceQL**
4. Run query: `{ service.name = "rctiplus-api" }`

### 4. Verifikasi Konfigurasi Tempo

**Pastikan metrics_generator enabled:**
```yaml
metrics_generator:
  processors: [service-graphs, span-metrics, local-blocks]
  remote_write:
    - url: http://prometheus:9090/api/v1/write
      send_exemplars: true
```

**Pastikan overrides:**
```yaml
overrides:
  defaults:
    metrics_generator:
      processors: [service-graphs, span-metrics, local-blocks]
```

### 5. Verifikasi Grafana Datasource

**Check serviceMap configuration:**
```yaml
serviceMap:
  datasourceUid: 'mimir'  # atau 'prometheus'
```

**Test query di Grafana:**
1. Buka **Explore** → **Mimir** (atau **Prometheus**)
2. Query: `traces_service_graph_request_total`
3. Harus ada data

### 6. Check Time Range

Service graph metrics membutuhkan waktu untuk di-generate:
- **Minimum**: 1-2 menit setelah traffic pertama
- **Recommended**: Tunggu 5 menit setelah generate traffic

**Pastikan time range di dashboard:**
- From: `now-1h` atau lebih lama
- To: `now`

### 7. Verifikasi Service Names

**Pastikan service names konsisten:**
- Service name di aplikasi: `OTEL_SERVICE_NAME=rctiplus-api`
- Service name harus sama di semua traces
- Check di Tempo: `{ service.name = "rctiplus-api" }`

### 8. Check Network Connectivity

**Tempo ke Prometheus:**
```bash
docker exec tempo ping -c 3 prometheus
docker exec tempo curl -v http://prometheus:9090/api/v1/write
```

**Grafana ke Mimir/Prometheus:**
```bash
docker exec grafana ping -c 3 mimir
docker exec grafana ping -c 3 prometheus
```

## Common Issues

### Issue 1: Metrics Generator Tidak Aktif

**Symptoms:**
- Tidak ada metrics di `/metrics` endpoint
- Logs tidak menunjukkan metrics_generator activity

**Solution:**
1. Restart Tempo: `docker-compose restart tempo`
2. Check logs: `docker logs tempo | grep metrics_generator`
3. Pastikan config file ter-mount dengan benar

### Issue 2: Remote Write Gagal

**Symptoms:**
- Metrics ada di Tempo tapi tidak di Prometheus
- Error di Tempo logs tentang remote_write

**Solution:**
1. Check Prometheus remote write receiver enabled:
   ```bash
   docker-compose ps prometheus
   # Check command: --web.enable-remote-write-receiver
   ```
2. Check network: `docker exec tempo curl http://prometheus:9090/api/v1/write`
3. Check Prometheus logs: `docker logs prometheus | grep remote_write`

### Issue 3: Tidak Ada Cross-Service Traffic

**Symptoms:**
- Traces ada tapi tidak ada service graph
- Service graph kosong

**Solution:**
1. Pastikan ada cross-service calls:
   ```bash
   # Ini akan call rctiplus-api yang kemudian call user-profile-service
   curl http://localhost:8080/api/users-with-profile/1
   ```
2. Check traces memiliki multiple services:
   ```bash
   curl http://localhost:3200/api/search?limit=10 | jq '.traces[].rootServiceName'
   ```

### Issue 4: Wrong Datasource UID

**Symptoms:**
- Error "datasource not found"
- Service graph tidak load

**Solution:**
1. Check datasource UID di Grafana:
   - Buka **Configuration** → **Data Sources**
   - Check UID untuk Mimir/Prometheus
2. Update `datasources.yml`:
   ```yaml
   serviceMap:
     datasourceUid: 'mimir'  # Pastikan UID sesuai
   ```
3. Restart Grafana: `docker-compose restart grafana`

## Quick Diagnostic Commands

```bash
# 1. Check Tempo metrics
curl http://localhost:3200/metrics | grep traces_service_graph

# 2. Check Prometheus metrics
curl "http://localhost:9090/api/v1/query?query=traces_service_graph_request_total"

# 3. Check Mimir metrics
curl "http://localhost:9009/prometheus/api/v1/query?query=traces_service_graph_request_total"

# 4. Generate traffic
curl http://localhost:8080/api/users-with-profile/1

# 5. Check Tempo traces
curl http://localhost:3200/api/search?limit=5

# 6. Check logs
docker logs tempo | tail -50
docker logs prometheus | tail -50
```

## Expected Metrics

Setelah setup benar, Anda harus melihat metrics berikut:

### Service Graph Metrics
- `traces_service_graph_request_total{client, server}`
- `traces_service_graph_request_failed_total{client, server}`
- `traces_service_graph_request_server_seconds_bucket{client, server, le}`

### Span Metrics
- `traces_spanmetrics_calls_total{service_name, span_name}`
- `traces_spanmetrics_latency_bucket{service_name, span_name, le}`

## Verification Checklist

- [ ] Tempo metrics_generator enabled
- [ ] service-graphs processor aktif
- [ ] Remote write ke Prometheus/Mimir berhasil
- [ ] Prometheus remote write receiver enabled
- [ ] Ada cross-service traffic
- [ ] Traces muncul di Tempo
- [ ] Metrics muncul di Prometheus/Mimir
- [ ] Grafana datasource UID benar
- [ ] Time range di dashboard cukup lama
- [ ] Service names konsisten

## Still Not Working?

1. **Check semua logs:**
   ```bash
   docker-compose logs tempo | grep -i error
   docker-compose logs prometheus | grep -i error
   docker-compose logs grafana | grep -i error
   ```

2. **Restart semua services:**
   ```bash
   docker-compose restart tempo prometheus grafana
   ```

3. **Verify configuration:**
   ```bash
   cd infrastructure
   ./verify-setup.sh
   ```

4. **Check documentation:**
   - [Tempo Service Graphs](https://grafana.com/docs/tempo/latest/metrics-generator/service_graphs/)
   - [Grafana Service Graph](https://grafana.com/docs/grafana/latest/datasources/tempo/service-graph/)

