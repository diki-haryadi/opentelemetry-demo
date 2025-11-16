# Grafana Beyla Integration - Zero-Code Instrumentation

## Overview

Grafana Beyla telah diintegrasikan sebagai **sidecar containers** untuk memberikan **zero-code instrumentation** menggunakan **eBPF** (extended Berkeley Packet Filter). Beyla secara otomatis menangkap traces, metrics, dan network telemetry tanpa perlu modifikasi code aplikasi.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Container                     │
│  ┌────────────┐                                             │
│  │   App      │ ← No code changes needed!                   │
│  │ (Go/Java/  │                                             │
│  │  Python)   │                                             │
│  └────────────┘                                             │
│        ↑                                                     │
│        │ eBPF probes attach here                            │
│        ↓                                                     │
│  ┌────────────────────────────────────────┐                │
│  │      Beyla Sidecar Container           │                │
│  │  ┌──────────────────────────────────┐  │                │
│  │  │  eBPF Instrumentation Engine     │  │                │
│  │  │  - HTTP/HTTPS traffic capture    │  │                │
│  │  │  - gRPC traffic capture          │  │                │
│  │  │  - Database calls                │  │                │
│  │  │  - Network metrics               │  │                │
│  │  └──────────────────────────────────┘  │                │
│  │              ↓                          │                │
│  │  ┌──────────────────────────────────┐  │                │
│  │  │  OTLP Exporter                   │  │                │
│  │  │  → Traces to Alloy               │  │                │
│  │  │  → Metrics to Alloy              │  │                │
│  │  └──────────────────────────────────┘  │                │
│  └────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────┘
                         ↓
                   ┌──────────┐
                   │  Alloy   │
                   └──────────┘
                         ↓
        ┌────────────────┼────────────────┐
        ↓                ↓                ↓
   ┌────────┐      ┌────────┐      ┌──────────┐
   │ Tempo  │      │  Loki  │      │Prometheus│
   └────────┘      └────────┘      └──────────┘
```

## What is Beyla?

Beyla adalah **eBPF-based auto-instrumentation tool** dari Grafana Labs yang:

✅ **Zero-code** - Tidak perlu modifikasi aplikasi  
✅ **Language agnostic** - Bekerja dengan Go, Java, Python, Node.js, Ruby, Rust, dll  
✅ **Low overhead** - eBPF berjalan di kernel space dengan minimal overhead  
✅ **Comprehensive** - Menangkap HTTP, gRPC, database calls, network metrics  
✅ **Production-ready** - Aman untuk production environments  

## Implementation Details

### Services with Beyla Sidecars

Setiap aplikasi memiliki Beyla sidecar:

1. **frontend** → **beyla-frontend**
2. **backend** → **beyla-backend**
3. **payment** → **beyla-payment**

### Docker Compose Configuration

#### Sidecar Pattern

```yaml
# Application container
frontend:
  build: ./apps/frontend
  ports:
    - "8080:8080"
  pid: "host"  # Required for eBPF
  networks:
    - alloy-demo

# Beyla sidecar
beyla-frontend:
  image: grafana/beyla:latest
  privileged: true  # Required for eBPF
  pid: "service:frontend"  # Share PID namespace
  network_mode: "service:frontend"  # Share network namespace
  volumes:
    - ./config/beyla/beyla-frontend.yml:/etc/beyla/config.yml
    - /sys/kernel/debug:/sys/kernel/debug:ro
  environment:
    - BEYLA_OPEN_PORT=8080
    - OTEL_EXPORTER_OTLP_ENDPOINT=http://alloy:4318
```

### Key Configuration Elements

#### 1. **Privileged Mode**
```yaml
privileged: true
```
Required untuk eBPF programs yang perlu akses ke kernel.

#### 2. **PID Namespace Sharing**
```yaml
pid: "service:frontend"
```
Beyla perlu akses ke process tree aplikasi untuk attach eBPF probes.

#### 3. **Network Namespace Sharing**
```yaml
network_mode: "service:frontend"
```
Beyla menangkap network traffic dari aplikasi.

#### 4. **Kernel Debug Filesystem**
```yaml
volumes:
  - /sys/kernel/debug:/sys/kernel/debug:ro
```
eBPF memerlukan akses ke kernel debug filesystem.

## Beyla Configuration Files

### Frontend Configuration (`beyla-frontend.yml`)

```yaml
# OTLP exporters
otel_metrics_export:
  endpoint: http://alloy:4318
  protocol: http/protobuf
  interval: 10s

otel_traces_export:
  endpoint: http://alloy:4318
  protocol: http/protobuf

# Service discovery
discovery:
  services:
    - name: frontend
      namespace: alloy-demo
      open_port: 8080

# Route patterns
routes:
  unmatch: heuristic
  patterns:
    - /
    - /products
    - /checkout

# Network telemetry
network:
  enable: true
  allowed_attributes:
    - src.address
    - dst.address
    - src.name
    - dst.name
```

## Telemetry Collected by Beyla

### 1. **HTTP Traces**

Beyla automatically captures:
- Request/Response traces
- HTTP method, path, status code
- Request duration
- Error tracking
- Distributed tracing context propagation

**Metrics:**
- `http_server_request_duration_seconds`
- `http_server_request_body_size_bytes`
- `http_server_response_body_size_bytes`

### 2. **Network Metrics**

- Network flow data (src/dst IP, ports)
- Bytes sent/received
- Connection tracking
- Service-to-service communication

**Metrics:**
- `beyla_network_flow_bytes_total`
- `beyla_network_flow_packets_total`

### 3. **Application Metrics**

- Request rate (RPS)
- Error rate
- Latency percentiles (p50, p95, p99)
- Active connections

### 4. **Service Graph Data**

Beyla contributes to service graph by capturing:
- Client → Server relationships
- Request counts between services
- Latency between services

## Dual Instrumentation Strategy

Setup ini menggunakan **dual instrumentation**:

### 1. **Manual Instrumentation** (otel-lib)
- ✅ Rich business context
- ✅ Custom metrics (payment_amount, etc)
- ✅ Custom spans with attributes
- ✅ Application-specific logging

### 2. **Auto Instrumentation** (Beyla)
- ✅ Zero-code overhead
- ✅ Network-level visibility
- ✅ Backup telemetry jika manual instrumentation gagal
- ✅ Infrastructure-level metrics

## Benefits of Dual Approach

| Aspect | Manual (otel-lib) | Auto (Beyla) |
|--------|-------------------|--------------|
| **Setup Effort** | Medium | Minimal |
| **Code Changes** | Required | None |
| **Business Context** | Rich | Limited |
| **Network Visibility** | Limited | Excellent |
| **Overhead** | Low | Very Low |
| **Coverage** | Application code | All network traffic |
| **Maintenance** | Code updates | Config only |

**Combined Benefits:**
- 🎯 Complete observability coverage
- 🔍 Deep application insights + network visibility
- 🛡️ Redundancy - if one fails, other still works
- 📊 Correlation between app metrics and network metrics

## Viewing Beyla Data in Grafana

### 1. **Service Graph Dashboard**

Beyla data akan muncul di service graph dengan suffix `-beyla`:
- `frontend-beyla` → `backend-beyla`
- `backend-beyla` → `payment-beyla`

### 2. **Traces in Tempo**

Filter by service name:
```
{service.name="frontend-beyla"}
{service.name="backend-beyla"}
{service.name="payment-beyla"}
```

### 3. **Metrics in Prometheus**

Query Beyla metrics:
```promql
# HTTP request duration
http_server_request_duration_seconds{service_name="frontend-beyla"}

# Network flow bytes
beyla_network_flow_bytes_total{src_name="frontend-beyla"}

# Request rate
rate(http_server_request_duration_seconds_count[5m])
```

### 4. **Beyla-specific Metrics**

Access Beyla's own metrics:
- Frontend Beyla: http://localhost:9400/metrics
- Backend Beyla: http://localhost:9401/metrics
- Payment Beyla: http://localhost:9402/metrics

## Comparison: Manual vs Beyla Telemetry

### Example: Frontend Service

#### Manual Instrumentation (otel-lib)
```
Service: frontend
Spans:
  - GET /products
    - attributes: http.route=/products, custom.user_id=123
    - child span: call_backend
      - attributes: backend.url=http://backend:8081
```

#### Beyla Auto-Instrumentation
```
Service: frontend-beyla
Spans:
  - HTTP GET
    - attributes: http.method=GET, http.target=/products
    - network: src.address=172.18.0.5, dst.address=172.18.0.6
    - metrics: bytes_sent=1024, bytes_received=4096
```

**Together:** Complete picture of application behavior + network communication!

## Troubleshooting

### Beyla Not Starting

**Check kernel version:**
```bash
uname -r  # Should be >= 5.8 for eBPF
```

**Check privileged mode:**
```bash
docker inspect beyla-frontend | grep Privileged
# Should show "Privileged": true
```

### No Telemetry Data

**Check Beyla logs:**
```bash
docker logs beyla-frontend
docker logs beyla-backend
docker logs beyla-payment
```

**Verify eBPF programs loaded:**
```bash
docker exec beyla-frontend bpftool prog list
```

**Check OTLP endpoint:**
```bash
# From Beyla container
docker exec beyla-frontend curl -v http://alloy:4318/v1/traces
```

### High CPU Usage

Beyla biasanya menggunakan <1% CPU, jika tinggi:

1. Reduce metrics interval:
```yaml
otel_metrics_export:
  interval: 30s  # Increase from 10s
```

2. Disable network telemetry if not needed:
```yaml
network:
  enable: false
```

3. Limit route patterns:
```yaml
routes:
  patterns:
    - /api/*  # Use wildcards instead of listing all routes
```

## Performance Impact

### Resource Usage (per Beyla instance)

- **Memory**: ~50-100 MB
- **CPU**: <1% in steady state
- **Network**: Minimal (only OTLP export)
- **Disk**: None (no local storage)

### Application Impact

- **Latency**: <1ms added latency
- **Throughput**: No measurable impact
- **Code**: Zero changes required

## Security Considerations

### Privileged Containers

Beyla requires `privileged: true` for eBPF:

✅ **Safe because:**
- Beyla only reads data, doesn't modify
- eBPF programs are verified by kernel
- No access to application memory
- Runs in separate container

⚠️ **Best practices:**
- Use in trusted environments
- Monitor Beyla logs
- Keep Beyla image updated
- Use network policies to limit Beyla's network access

### Data Privacy

Beyla captures:
- ✅ HTTP headers (can be filtered)
- ✅ Request/response sizes
- ✅ Network metadata
- ❌ Request/response bodies (not captured by default)

Configure sensitive data filtering:
```yaml
attributes:
  select:
    exclude:
      - http.request.header.authorization
      - http.request.header.cookie
```

## Testing Beyla

### 1. Start the Stack

```bash
cd alloy-demo
docker compose up -d
```

### 2. Verify Beyla Containers

```bash
docker ps | grep beyla
# Should show 3 Beyla containers running
```

### 3. Check Beyla Logs

```bash
docker logs beyla-frontend --tail=50
# Look for: "eBPF programs loaded successfully"
```

### 4. Generate Traffic

```bash
./load-generator.sh
```

### 5. View in Grafana

- Service Graph: http://localhost:3000/d/service-graph
  - Look for `*-beyla` services
- Explore → Tempo → Search for `service.name="frontend-beyla"`
- Explore → Prometheus → Query `http_server_request_duration_seconds`

## Advanced Configuration

### Custom eBPF Programs

Beyla supports custom eBPF programs for specific use cases:

```yaml
ebpf:
  functions:
    - name: "my_custom_function"
      namespace: "my_app"
```

### Sampling

Reduce data volume with sampling:

```yaml
otel_traces_export:
  sampler:
    name: "traceidratio"
    arg: "0.1"  # Sample 10% of traces
```

### Multiple Ports

If app listens on multiple ports:

```yaml
discovery:
  services:
    - name: frontend
      open_port: 8080
    - name: frontend-admin
      open_port: 9090
```

## Migration Path

### Phase 1: Beyla Only (Current)
- ✅ Zero-code instrumentation active
- ✅ Manual instrumentation also active
- ✅ Dual telemetry for comparison

### Phase 2: Evaluate (Recommended)
- Compare Beyla vs manual instrumentation
- Identify gaps in Beyla coverage
- Decide which to keep long-term

### Phase 3: Optimize
- **Option A**: Keep both (recommended)
  - Rich app context + network visibility
- **Option B**: Beyla only
  - Remove otel-lib, simplify code
- **Option C**: Manual only
  - Remove Beyla, full control

## Resources

- [Beyla Documentation](https://grafana.com/docs/beyla/)
- [Beyla GitHub](https://github.com/grafana/beyla)
- [eBPF Introduction](https://ebpf.io/)
- [Beyla Blog Post](https://grafana.com/blog/2023/11/14/grafana-beyla-1.0-release-zero-code-instrumentation-for-application-telemetry-using-ebpf/)

## Summary

✅ **Implemented:**
- 3 Beyla sidecars (frontend, backend, payment)
- eBPF-based zero-code instrumentation
- OTLP export to Alloy
- Network telemetry collection
- HTTP/gRPC tracing

🎯 **Benefits:**
- No code changes required
- Complete network visibility
- Backup telemetry layer
- Production-safe eBPF
- Low overhead (<1% CPU)

🔄 **Next Steps:**
1. Start stack: `docker compose up -d`
2. Generate traffic: `./load-generator.sh`
3. View Beyla data in Grafana
4. Compare with manual instrumentation
5. Optimize based on your needs
