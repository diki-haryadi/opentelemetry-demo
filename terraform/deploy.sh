#!/bin/bash
# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}OpenTelemetry Demo - Terraform Deployment Script${NC}"
echo "=================================================="
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check if terraform is installed
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}Error: terraform is not installed${NC}"
    echo "Please install Terraform: https://developer.hashicorp.com/terraform/downloads"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    echo "Please install kubectl: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

# Check if kubernetes cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Error: Cannot connect to Kubernetes cluster${NC}"
    echo "Please ensure your Kubernetes cluster is running and accessible"
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites met${NC}"
echo ""

# Get current context
CURRENT_CONTEXT=$(kubectl config current-context)
echo -e "Current Kubernetes context: ${YELLOW}${CURRENT_CONTEXT}${NC}"
echo ""

# Ask for confirmation
read -p "Deploy OpenTelemetry Demo to this cluster? (yes/no): " -r
echo ""
if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
    echo "Deployment cancelled"
    exit 0
fi

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Run terraform plan
echo ""
echo "Running Terraform plan..."
terraform plan

# Ask for confirmation to apply
echo ""
read -p "Apply this Terraform plan? (yes/no): " -r
echo ""
if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
    echo "Deployment cancelled"
    exit 0
fi

# Apply Terraform
echo "Deploying OpenTelemetry Demo..."
terraform apply -auto-approve

# Show access instructions
echo ""
echo -e "${GREEN}=================================================="
echo "Deployment Complete!"
echo -e "==================================================${NC}"
echo ""
echo "To access the demo, run the following command:"
echo ""
echo -e "${YELLOW}kubectl port-forward -n otel-demo svc/opentelemetry-demo-frontendproxy 8080:8080${NC}"
echo ""
echo "Then open your browser to:"
echo "  • Demo UI: http://localhost:8080"
echo "  • Jaeger UI: http://localhost:8080/jaeger/ui"
echo "  • Grafana: http://localhost:8080/grafana/"
echo "  • Load Generator: http://localhost:8080/loadgen/"
echo ""
echo "To check the status of your deployment:"
echo -e "${YELLOW}kubectl get pods -n otel-demo${NC}"
echo ""
