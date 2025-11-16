# OpenTelemetry Demo - Terraform Deployment

This directory contains Terraform configuration to deploy the OpenTelemetry Demo to a Kubernetes cluster using Helm.

## Prerequisites

Before you begin, ensure you have the following installed:

1. **Terraform** (>= 1.0): [Install Terraform](https://developer.hashicorp.com/terraform/downloads)
2. **kubectl**: [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
3. **A Kubernetes cluster** - Choose one:
   - [Docker Desktop](https://www.docker.com/products/docker-desktop/) (easiest for local development)
   - [minikube](https://minikube.sigs.k8s.io/docs/start/)
   - [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
   - Cloud-managed cluster (EKS, GKE, AKS)

## Quick Start

### 1. Verify Kubernetes Cluster

Ensure your Kubernetes cluster is running and accessible:

```bash
kubectl cluster-info
kubectl get nodes
```

### 2. Initialize Terraform

```bash
cd terraform
terraform init
```

### 3. Review the Plan

```bash
terraform plan
```

### 4. Deploy the Demo

```bash
terraform apply
```

Type `yes` when prompted to confirm the deployment.

### 5. Access the Demo

After deployment completes, Terraform will output access instructions. The easiest way to access the demo is via port-forwarding:

```bash
kubectl port-forward -n otel-demo svc/opentelemetry-demo-frontendproxy 8080:8080
```

Then open your browser to:
- **Demo UI**: http://localhost:8080
- **Jaeger UI**: http://localhost:8080/jaeger/ui
- **Grafana**: http://localhost:8080/grafana/
- **Load Generator UI**: http://localhost:8080/loadgen/
- **Feature Flags**: http://localhost:8080/feature/

## Configuration Options

### Using a terraform.tfvars file

Create a `terraform.tfvars` file to customize your deployment:

```hcl
# Kubernetes configuration
namespace       = "otel-demo"
release_name    = "opentelemetry-demo"
service_type    = "LoadBalancer"  # ClusterIP, NodePort, or LoadBalancer

# Resource limits
resource_limits_enabled = true
default_memory_limit    = "512Mi"
default_cpu_limit       = "500m"

# Custom settings
custom_settings = {
  "default.replicas" = "2"
}
```

### Using command-line variables

```bash
terraform apply -var="namespace=my-otel-demo" -var="service_type=LoadBalancer"
```

### Using custom Helm values

For advanced configuration, you can provide a custom Helm values file:

```bash
terraform apply -var-file=custom-values.tfvars
```

Example `custom-values.tfvars`:

```hcl
helm_values = <<-EOT
  default:
    enabled: true
    replicas: 2

  jaeger:
    enabled: true

  prometheus:
    enabled: true
EOT
```

## Available Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `kubeconfig_path` | Path to kubeconfig file | `~/.kube/config` |
| `kube_context` | Kubernetes context to use | `""` (current context) |
| `namespace` | Kubernetes namespace | `otel-demo` |
| `release_name` | Helm release name | `opentelemetry-demo` |
| `chart_version` | Helm chart version | `""` (latest) |
| `service_type` | Service type for frontend | `ClusterIP` |
| `image_pull_policy` | Image pull policy | `IfNotPresent` |
| `resource_limits_enabled` | Enable resource limits | `false` |
| `default_memory_limit` | Default memory limit | `500Mi` |
| `default_cpu_limit` | Default CPU limit | `500m` |
| `helm_values` | Additional Helm values (YAML) | `""` |
| `custom_settings` | Map of custom settings | `{}` |

## Platform-Specific Setup

### Docker Desktop Kubernetes

1. Enable Kubernetes in Docker Desktop settings
2. Verify: `kubectl config current-context` should show `docker-desktop`
3. Deploy: `terraform apply`

### minikube

```bash
# Start minikube with sufficient resources
minikube start --cpus=4 --memory=8192 --disk-size=20g

# Deploy
terraform apply -var="kube_context=minikube"

# Access via minikube service
minikube service opentelemetry-demo-frontendproxy -n otel-demo
```

### kind (Kubernetes in Docker)

```bash
# Create a kind cluster
kind create cluster --name otel-demo

# Deploy
terraform apply -var="kube_context=kind-otel-demo"
```

### Cloud Providers (EKS, GKE, AKS)

For cloud deployments, you can use `service_type = "LoadBalancer"` to expose the demo:

```bash
terraform apply -var="service_type=LoadBalancer"
```

Then get the external IP:

```bash
kubectl get svc -n otel-demo opentelemetry-demo-frontendproxy
```

## Troubleshooting

### Check deployment status

```bash
kubectl get pods -n otel-demo
kubectl get svc -n otel-demo
```

### View logs for a specific pod

```bash
kubectl logs -n otel-demo <pod-name>
```

### View Helm release status

```bash
helm list -n otel-demo
helm status opentelemetry-demo -n otel-demo
```

### Pods not starting (insufficient resources)

If running locally and pods are pending or crashing, you may need to:
- Increase Docker Desktop memory allocation (recommended: 8GB+)
- Reduce resource requests by disabling some components
- Enable resource limits: `terraform apply -var="resource_limits_enabled=false"`

### Connection refused when accessing services

Ensure port-forwarding is active:

```bash
kubectl port-forward -n otel-demo svc/opentelemetry-demo-frontendproxy 8080:8080
```

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

Type `yes` when prompted to confirm.

## Advanced Configuration

### Enabling/Disabling Components

Create a custom values file to disable certain components:

```hcl
helm_values = <<-EOT
  components:
    accountingService:
      enabled: true
    adService:
      enabled: true
    cartService:
      enabled: true
    checkoutService:
      enabled: true
    currencyService:
      enabled: true
    emailService:
      enabled: true
    featureflagService:
      enabled: true
    frauddetectionService:
      enabled: true
    frontend:
      enabled: true
    frontendProxy:
      enabled: true
    loadgenerator:
      enabled: true
    paymentService:
      enabled: true
    productCatalogService:
      enabled: true
    quoteService:
      enabled: true
    recommendationService:
      enabled: true
    shippingService:
      enabled: true

  # Observability backends
  jaeger:
    enabled: true
  grafana:
    enabled: true
  prometheus:
    enabled: true
  opensearch:
    enabled: true
EOT
```

### Using with CI/CD

Example GitHub Actions workflow:

```yaml
name: Deploy OpenTelemetry Demo

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          kubeconfig: ${{ secrets.KUBECONFIG }}

      - name: Terraform Init
        run: terraform init
        working-directory: ./terraform

      - name: Terraform Apply
        run: terraform apply -auto-approve
        working-directory: ./terraform
```

## Additional Resources

- [OpenTelemetry Demo Documentation](https://opentelemetry.io/docs/demo/)
- [Helm Chart Repository](https://github.com/open-telemetry/opentelemetry-helm-charts)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)
- [Terraform Helm Provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs)
