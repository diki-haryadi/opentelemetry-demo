# OpenTelemetry Demo - Quick Start Guide

Get the OpenTelemetry Demo running in minutes!

## Prerequisites

1. **Kubernetes cluster** - Choose one:
   - Docker Desktop (Enable Kubernetes in settings) - Recommended for Mac/Windows
   - minikube: `minikube start --cpus=4 --memory=8192`
   - kind: `kind create cluster --name otel-demo`

2. **Terraform** - [Install here](https://developer.hashicorp.com/terraform/downloads)

3. **kubectl** - [Install here](https://kubernetes.io/docs/tasks/tools/)

## Deploy in 3 Steps

### Step 1: Verify Kubernetes is running

```bash
kubectl cluster-info
```

### Step 2: Run the deployment script

```bash
cd terraform
./deploy.sh
```

The script will:
- Check prerequisites
- Initialize Terraform
- Show you the plan
- Deploy the demo after confirmation

### Step 3: Access the demo

Once deployed, run:

```bash
kubectl port-forward -n otel-demo svc/opentelemetry-demo-frontendproxy 8080:8080
```

Then open: http://localhost:8080

## What You'll See

- **Demo UI** (http://localhost:8080) - Online shop application
- **Jaeger** (http://localhost:8080/jaeger/ui) - Distributed tracing
- **Grafana** (http://localhost:8080/grafana/) - Metrics dashboards
- **Load Generator** (http://localhost:8080/loadgen/) - Generate traffic
- **Feature Flags** (http://localhost:8080/feature/) - Toggle features

## Manual Deployment (Alternative)

If you prefer manual control:

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

## Check Deployment Status

```bash
# View all pods
kubectl get pods -n otel-demo

# View all services
kubectl get svc -n otel-demo

# View logs for a specific service
kubectl logs -n otel-demo -l app.kubernetes.io/component=frontend
```

## Cleanup

To remove everything:

```bash
./destroy.sh
```

Or manually:

```bash
terraform destroy
```

## Troubleshooting

### Pods stuck in "Pending" state

**Docker Desktop Users:**
- Increase memory to at least 8GB in Docker Desktop settings
- Restart Docker Desktop

**minikube Users:**
```bash
minikube delete
minikube start --cpus=4 --memory=8192 --disk-size=20g
```

### Port 8080 already in use

Use a different port:
```bash
kubectl port-forward -n otel-demo svc/opentelemetry-demo-frontendproxy 9090:8080
```
Then access at http://localhost:9090

### Cannot connect to Kubernetes

Check your context:
```bash
kubectl config get-contexts
kubectl config use-context <context-name>
```

## Next Steps

- Explore the [full README](README.md) for advanced configuration
- Check out example configurations in [examples/](examples/)
- Learn more at [OpenTelemetry Demo Docs](https://opentelemetry.io/docs/demo/)

## Common Customizations

### Use LoadBalancer (cloud providers)

```bash
terraform apply -var="service_type=LoadBalancer"
```

### Change namespace

```bash
terraform apply -var="namespace=my-otel-demo"
```

### Enable resource limits

```bash
terraform apply -var="resource_limits_enabled=true"
```

## Support

- Report issues: https://github.com/open-telemetry/opentelemetry-demo/issues
- Join Slack: https://cloud-native.slack.com/archives/C03B4CWV4DA
