# Kubernetes Deployment untuk RCTI Plus

Dokumentasi ini menjelaskan cara deploy RCTI Plus observability stack ke Kubernetes menggunakan Terraform dan Helm.

## Prerequisites

- Kubernetes cluster yang sudah running
- `kubectl` terkonfigurasi dengan akses ke cluster
- `terraform` terinstall
- `helm` terinstall (opsional, jika menggunakan Helm chart)

## Struktur File

```
kubernetes/
├── namespace.yaml              # Namespace definition
├── tempo-deployment.yaml      # Tempo deployment
├── tempo-service.yaml          # Tempo service
├── tempo-configmap.yaml        # Tempo configuration
├── alloy-deployment.yaml       # Alloy deployment
├── alloy-service.yaml          # Alloy service
├── alloy-configmap.yaml        # Alloy configuration
├── prometheus-deployment.yaml  # Prometheus deployment
├── prometheus-service.yaml     # Prometheus service
├── prometheus-configmap.yaml   # Prometheus configuration
├── mimir-deployment.yaml       # Mimir deployment
├── mimir-service.yaml          # Mimir service
├── mimir-configmap.yaml        # Mimir configuration
├── loki-deployment.yaml        # Loki deployment
├── loki-service.yaml           # Loki service
├── loki-configmap.yaml         # Loki configuration
├── jaeger-deployment.yaml      # Jaeger deployment
├── jaeger-service.yaml         # Jaeger service
├── grafana-deployment.yaml     # Grafana deployment
├── grafana-service.yaml        # Grafana service
├── grafana-datasources-configmap.yaml    # Grafana datasources
├── grafana-dashboards-configmap.yaml      # Grafana dashboards config
├── grafana-dashboard-files-configmap.yaml # Grafana dashboard files
├── app-deployment.yaml         # Application deployment
├── app-service.yaml            # Application service
├── values.yaml                 # Helm values
└── helm-chart/                 # Helm chart (optional)
```

## Deployment Methods

### Method 1: Menggunakan Terraform (Recommended)

Terraform akan membuat:
- Namespace
- ConfigMaps untuk semua konfigurasi
- Persistent Volume Claims untuk data persistence

1. **Initialize Terraform**:
   ```bash
   cd terraform
   terraform init
   ```

2. **Review dan Apply**:
   ```bash
   terraform plan
   terraform apply
   ```

3. **Deploy Kubernetes Resources**:
   Setelah Terraform selesai membuat ConfigMaps dan PVCs, deploy resources menggunakan Helm atau kubectl:
   
   **Menggunakan Helm**:
   ```bash
   helm install rctiplus ./kubernetes/helm-chart \
     --namespace rctiplus \
     --create-namespace \
     -f ./kubernetes/values.yaml
   ```
   
   **Menggunakan kubectl** (setelah template dengan Helm):
   ```bash
   # Template dengan Helm terlebih dahulu
   helm template rctiplus ./kubernetes/helm-chart \
     --namespace rctiplus \
     -f ./kubernetes/values.yaml | kubectl apply -f -
   ```

### Method 2: Menggunakan kubectl langsung

Jika Anda ingin menggunakan kubectl langsung tanpa Terraform:

1. **Create Namespace**:
   ```bash
   kubectl apply -f kubernetes/namespace.yaml
   ```

2. **Create ConfigMaps** (atau gunakan yang dibuat Terraform):
   ```bash
   kubectl apply -f kubernetes/tempo-configmap.yaml
   kubectl apply -f kubernetes/alloy-configmap.yaml
   kubectl apply -f kubernetes/prometheus-configmap.yaml
   kubectl apply -f kubernetes/mimir-configmap.yaml
   kubectl apply -f kubernetes/loki-configmap.yaml
   kubectl apply -f kubernetes/grafana-datasources-configmap.yaml
   kubectl apply -f kubernetes/grafana-dashboards-configmap.yaml
   kubectl apply -f kubernetes/grafana-dashboard-files-configmap.yaml
   ```

3. **Create PVCs** (atau gunakan yang dibuat Terraform):
   ```bash
   # Create PVCs manually or use Terraform
   ```

4. **Deploy Services**:
   ```bash
   # Template YAML files dengan Helm values terlebih dahulu
   helm template rctiplus ./kubernetes/helm-chart \
     --namespace rctiplus \
     -f ./kubernetes/values.yaml | kubectl apply -f -
   ```

## Service Graph Configuration

Konfigurasi service graph sudah termasuk:

1. **Tempo**: 
   - Metrics generator dengan `service-graphs` processor
   - Remote write ke Prometheus untuk service graph metrics
   - ConfigMap: `tempo-config`

2. **Grafana**:
   - Service Graph Dashboard sudah tersedia
   - Datasource Tempo dikonfigurasi dengan serviceMap enabled
   - ConfigMap: `grafana-dashboard-files` (service-tree.json)

3. **Jaeger**:
   - Service graph built-in di Jaeger UI
   - Deployment dan Service sudah dikonfigurasi

## Access Services

Setelah deployment, akses services menggunakan port-forward atau Ingress:

```bash
# Grafana
kubectl port-forward -n rctiplus svc/grafana 3000:3000
# Access: http://localhost:3000

# Jaeger UI
kubectl port-forward -n rctiplus svc/jaeger 16686:16686
# Access: http://localhost:16686

# Prometheus
kubectl port-forward -n rctiplus svc/prometheus 9090:9090
# Access: http://localhost:9090

# Tempo
kubectl port-forward -n rctiplus svc/tempo 3200:3200
# Access: http://localhost:3200

# Alloy UI
kubectl port-forward -n rctiplus svc/alloy 12345:12345
# Access: http://localhost:12345
```

## Verifikasi Deployment

1. **Check Pods**:
   ```bash
   kubectl get pods -n rctiplus
   ```

2. **Check Services**:
   ```bash
   kubectl get svc -n rctiplus
   ```

3. **Check ConfigMaps**:
   ```bash
   kubectl get configmaps -n rctiplus
   ```

4. **Check PVCs**:
   ```bash
   kubectl get pvc -n rctiplus
   ```

## Troubleshooting

### Pods tidak running

```bash
# Check pod logs
kubectl logs -n rctiplus <pod-name>

# Check pod events
kubectl describe pod -n rctiplus <pod-name>
```

### ConfigMap tidak ditemukan

Pastikan ConfigMaps sudah dibuat oleh Terraform atau manual:
```bash
kubectl get configmap -n rctiplus
```

### Service Graph tidak muncul

1. Pastikan Tempo metrics generator sudah enabled dengan service-graphs processor
2. Pastikan Prometheus menerima remote write dari Tempo
3. Generate traffic ke aplikasi untuk menghasilkan traces
4. Tunggu beberapa menit untuk metrics ter-generate

### PVC tidak ter-bind

Pastikan StorageClass tersedia di cluster:
```bash
kubectl get storageclass
```

## Cleanup

### Menggunakan Terraform

```bash
cd terraform
terraform destroy
```

### Menggunakan Helm

```bash
helm uninstall rctiplus -n rctiplus
```

### Menggunakan kubectl

```bash
kubectl delete namespace rctiplus
```

## Notes

- File YAML menggunakan Helm templates (`{{ .Values.namespace }}`), jadi perlu di-template dengan Helm sebelum apply
- ConfigMaps dibuat oleh Terraform dari file di `infrastructure/` directory
- Semua konfigurasi sudah include service graph support
- Tempo remote_write sudah dikonfigurasi untuk mengirim service graph metrics ke Prometheus

