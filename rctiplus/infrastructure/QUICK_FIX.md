# Quick Fix: "No service graph data found"

## Langkah Cepat untuk Memperbaiki

### Step 1: Generate Cross-Service Traffic

Service graph membutuhkan **cross-service calls** untuk generate data. Pastikan ada traffic antara services:

```bash
# Generate traffic yang akan call rctiplus-api -> user-profile-service
curl http://localhost:8080/api/users-with-profile/1

# Tunggu 1-2 menit untuk Tempo memproses traces dan generate metrics
sleep 60
```

### Step 2: Check Metrics dengan Script

Jalankan script diagnostic:

```bash
cd infrastructure
./check-service-graph.sh
```

Script ini akan check:
- ✅ Tempo metrics generator
- ✅ Prometheus/Mimir metrics
- ✅ Traces di Tempo
- ✅ Service names

### Step 3: Verify di Grafana

1. **Check Explore → Tempo → Service Graph:**
   - Buka: http://localhost:3000/explore
   - Pilih datasource: **Tempo**
   - Query type: **Service Graph**
   - Time range: **Last 1 hour** (atau lebih lama)
   - Klik **Run Query**

2. **Check Metrics di Prometheus/Mimir:**
   - Buka: http://localhost:3000/explore
   - Pilih datasource: **Mimir** (atau **Prometheus**)
   - Query: `traces_service_graph_request_total`
   - Harus ada data dengan labels `client` dan `server`

### Step 4: Check Dashboard Time Range

Pastikan time range di dashboard cukup lama:

1. Buka dashboard: **Service Graph Dashboard**
2. Klik time picker di kanan atas
3. Pilih: **Last 1 hour** atau **Last 6 hours**
4. Klik **Apply**

### Step 5: Restart Services (Jika Masih Tidak Ada)

```bash
# Restart Tempo untuk memastikan metrics generator aktif
docker-compose restart tempo

# Tunggu 30 detik
sleep 30

# Generate traffic lagi
curl http://localhost:8080/api/users-with-profile/1

# Tunggu 2 menit
sleep 120
```

## Checklist

- [ ] Services running: `docker-compose ps`
- [ ] Generated cross-service traffic
- [ ] Waited 1-2 minutes after traffic
- [ ] Time range di dashboard cukup lama (min 1 hour)
- [ ] Check metrics di Prometheus/Mimir ada data
- [ ] Check traces di Tempo ada data

## Common Causes

### 1. Belum Ada Traffic
**Solution:** Generate traffic dengan `curl http://localhost:8080/api/users-with-profile/1`

### 2. Time Range Terlalu Pendek
**Solution:** Set time range ke "Last 1 hour" atau lebih lama

### 3. Metrics Belum Di-generate
**Solution:** Tunggu 1-2 menit setelah generate traffic, atau restart Tempo

### 4. Wrong Datasource
**Solution:** Pastikan `serviceMap.datasourceUid` di `datasources.yml` mengarah ke 'mimir' atau 'prometheus' yang benar

### 5. Tidak Ada Cross-Service Calls
**Solution:** Pastikan ada calls antara services (bukan hanya single service)

## Expected Result

Setelah semua langkah, Anda harus melihat:

1. **Di Explore → Tempo → Service Graph:**
   - Node graph dengan services (rctiplus-api, user-profile-service)
   - Edges menunjukkan connections
   - Numbers menunjukkan request rate

2. **Di Dashboard:**
   - Service graph visualization
   - Metrics panels menampilkan data
   - Request rate, latency, error rate charts

## Still Not Working?

Jalankan diagnostic lengkap:

```bash
cd infrastructure
./check-service-graph.sh
```

Dan lihat detail di: `TROUBLESHOOTING.md`

