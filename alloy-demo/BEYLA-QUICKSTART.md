# Beyla Quick Start Guide

## What is Beyla?

**Grafana Beyla** adalah zero-code instrumentation tool yang menggunakan **eBPF** (extended Berkeley Packet Filter) untuk secara otomatis menangkap telemetry dari aplikasi tanpa perlu modifikasi code.

## Why Beyla?

✅ **No Code Changes** - Deploy dan langsung dapat telemetry  
✅ **Language Agnostic** - Bekerja dengan Go, Java, Python, Node.js, dll  
✅ **Low Overhead** - eBPF berjalan di kernel dengan minimal impact  
✅ **Network Visibility** - Menangkap semua network traffic  
✅ **Production Safe** - Verified eBPF programs, aman untuk production  

## Quick Start

### 1. Start the Stack

```bash
cd alloy-demo
docker compose up -d
```

Ini akan start:
- 3 aplikasi (frontend, backend, payment)
- 3 Beyla sidecars (auto-instrumentation)
- Observability stack (Alloy, Tempo, Loki, Prometheus, Grafana)

### 2. Verify Beyla is Running

```bash
# Check Beyla containers
docker ps | grep beyla

# Should show:
# beyla-frontend
# beyla-backend
# beyla-payment
```

### 3. Check Beyla Logs

```bash
# Frontend Beyla
docker logs beyla-frontend

# Look for:
# "eBPF programs loaded successfully"
# "Starting OTLP exporter"
# "Instrumenting service: frontend"
```

### 4. Generate Traffic

```bash
./load-generator.sh
```

### 5. View Beyla Data

#### In Grafana (http://localhost:3000)

**Service Graph:**
- Go to: http://localhost:3000/d/service-graph
- You'll see services with `-beyla` suffix:
  - `frontend-beyla`
  - `backend-beyla`
  - `payment-beyla`

**Traces in Tempo:**
1. Go to Explore → Select "Tempo"
2. Search by service: `{service.name="frontend-beyla"}`
3. You'll see HTTP traces automatically captured

**Metrics in Prometheus:**
1. Go to Explore → Select "Prometheus"
2. Query Beyla metrics:
```promql
# HTTP request duration
http_server_request_duration_seconds{service_name="frontend-beyla"}

# Request rate
rate(http_server_request_duration_seconds_count{service_name="frontend-beyla"}[5m])

# Network flow bytes
beyla_network_flow_bytes_total{src_name="frontend-beyla"}
```

## What Beyla Captures

### 1. HTTP Traces
- ✅ Request method, path, status code
- ✅ Request/response duration
- ✅ Error tracking
- ✅ Distributed trace context

### 2. Network Metrics
- ✅ Source/destination IPs and ports
- ✅ Bytes sent/received
- ✅ Connection tracking
- ✅ Service-to-service flows

### 3. Application Metrics
- ✅ Request rate (RPS)
- ✅ Error rate
- ✅ Latency (p50, p95, p99)
- ✅ Active connections

## Beyla vs Manual Instrumentation

This setup uses **BOTH**:

| Feature | Manual (otel-lib) | Auto (Beyla) |
|---------|-------------------|--------------|
| Code changes | Required | None |
| Business context | Rich | Limited |
| Network visibility | Limited | Excellent |
| Setup time | Medium | Instant |
| Maintenance | Code updates | Config only |

**Best of both worlds!** 🎉

## Configuration Files

### Beyla Configs
- `config/beyla/beyla-frontend.yml` - Frontend Beyla config
- `config/beyla/beyla-backend.yml` - Backend Beyla config
- `config/beyla/beyla-payment.yml` - Payment Beyla config

### Docker Compose
Each app has a Beyla sidecar:

```yaml
# Application
frontend:
  build: ./apps/frontend
  pid: "host"  # Required for eBPF

# Beyla sidecar
beyla-frontend:
  image: grafana/beyla:latest
  privileged: true  # Required for eBPF
  pid: "service:frontend"  # Share PID namespace
  network_mode: "service:frontend"  # Share network
  volumes:
    - ./config/beyla/beyla-frontend.yml:/etc/beyla/config.yml
```

## Troubleshooting

### Beyla Not Starting?

**Check logs:**
```bash
docker logs beyla-frontend
```

**Common issues:**
1. **Kernel too old** - Need Linux kernel >= 5.8
   ```bash
   uname -r  # Check kernel version
   ```

2. **Not privileged** - Beyla needs `privileged: true`
   ```bash
   docker inspect beyla-frontend | grep Privileged
   ```

3. **PID namespace issue** - Check `pid: "service:frontend"`

### No Telemetry Data?

**Check OTLP endpoint:**
```bash
# From Beyla container
docker exec beyla-frontend curl http://alloy:4318/v1/traces
```

**Verify app is running:**
```bash
docker ps | grep frontend
curl http://localhost:8080/
```

**Check Alloy is receiving data:**
```bash
docker logs alloy | grep beyla
```

### High CPU Usage?

Beyla should use <1% CPU. If high:

1. **Increase metrics interval:**
   Edit `config/beyla/beyla-frontend.yml`:
   ```yaml
   otel_metrics_export:
     interval: 30s  # Increase from 10s
   ```

2. **Disable network telemetry:**
   ```yaml
   network:
     enable: false
   ```

## Viewing Beyla Metrics Directly

Each Beyla exposes Prometheus metrics:

```bash
# Frontend Beyla metrics
curl http://localhost:9400/metrics

# Backend Beyla metrics
curl http://localhost:9401/metrics

# Payment Beyla metrics
curl http://localhost:9402/metrics
```

## Example Queries

### Prometheus

```promql
# Request rate by service
rate(http_server_request_duration_seconds_count[5m])

# P95 latency
histogram_quantile(0.95, 
  rate(http_server_request_duration_seconds_bucket[5m])
)

# Error rate
rate(http_server_request_duration_seconds_count{http_status_code=~"5.."}[5m])

# Network bytes by service
sum by (src_name, dst_name) (
  rate(beyla_network_flow_bytes_total[5m])
)
```

### Tempo (TraceQL)

```traceql
# All frontend traces
{service.name="frontend-beyla"}

# Slow requests (>100ms)
{service.name="frontend-beyla" && duration > 100ms}

# Errors
{service.name="frontend-beyla" && status=error}

# Specific endpoint
{service.name="frontend-beyla" && http.target="/products"}
```

## Comparison: Before vs After

### Before (Manual Only)
```
✅ Application traces with business context
✅ Custom metrics
❌ No network visibility
❌ Requires code changes
❌ Language-specific instrumentation
```

### After (Manual + Beyla)
```
✅ Application traces with business context
✅ Custom metrics
✅ Network visibility (NEW!)
✅ Zero-code backup instrumentation (NEW!)
✅ Service mesh-like observability (NEW!)
✅ Works even if manual instrumentation fails (NEW!)
```

## Performance Impact

### Resource Usage (per Beyla)
- **Memory**: ~50-100 MB
- **CPU**: <1% steady state
- **Network**: Minimal (OTLP export only)

### Application Impact
- **Latency**: <1ms added
- **Throughput**: No measurable impact
- **Code**: Zero changes

## Security

### Why Privileged?

Beyla needs `privileged: true` for eBPF:

✅ **Safe because:**
- eBPF programs are verified by kernel
- Beyla only reads, doesn't modify
- No access to application memory
- Runs in isolated container

### Data Privacy

Beyla captures:
- ✅ HTTP metadata (method, path, status)
- ✅ Network metadata (IPs, ports)
- ✅ Request/response sizes
- ❌ Request/response bodies (NOT captured)

## Next Steps

### 1. Explore the Data

- **Service Graph**: http://localhost:3000/d/service-graph
- **Traces**: Grafana → Explore → Tempo
- **Metrics**: Grafana → Explore → Prometheus

### 2. Compare Telemetry

Compare data from:
- Manual instrumentation (`frontend`, `backend`, `payment`)
- Auto instrumentation (`frontend-beyla`, `backend-beyla`, `payment-beyla`)

### 3. Customize Configuration

Edit Beyla configs in `config/beyla/`:
- Add route patterns
- Configure sampling
- Adjust metrics interval
- Enable/disable network telemetry

### 4. Read Full Documentation

- `BEYLA-INTEGRATION.md` - Complete Beyla documentation
- `MIGRATION-TO-OTEL-LIB.md` - Manual instrumentation guide
- `README.md` - Overall project documentation

## Useful Commands

```bash
# Start stack
docker compose up -d

# Stop stack
docker compose down

# View Beyla logs
docker logs -f beyla-frontend
docker logs -f beyla-backend
docker logs -f beyla-payment

# Restart Beyla (if config changed)
docker compose restart beyla-frontend
docker compose restart beyla-backend
docker compose restart beyla-payment

# Check Beyla metrics
curl http://localhost:9400/metrics  # frontend
curl http://localhost:9401/metrics  # backend
curl http://localhost:9402/metrics  # payment

# Generate traffic
./load-generator.sh

# View all containers
docker compose ps
```

## Summary

🎉 **You now have:**
- ✅ Zero-code instrumentation with Beyla
- ✅ Manual instrumentation with otel-lib
- ✅ Complete observability coverage
- ✅ Network-level visibility
- ✅ Production-ready setup

🚀 **Start exploring:**
1. Generate traffic: `./load-generator.sh`
2. Open Grafana: http://localhost:3000
3. View service graph with Beyla data
4. Compare manual vs auto instrumentation
5. Enjoy complete observability! 🎊
