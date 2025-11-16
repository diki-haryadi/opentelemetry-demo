#!/bin/bash
# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}OpenTelemetry Demo - Terraform Destroy Script${NC}"
echo "=================================================="
echo ""

# Check if terraform is initialized
if [ ! -d ".terraform" ]; then
    echo -e "${RED}Error: Terraform not initialized${NC}"
    echo "Please run 'terraform init' first or use the deploy.sh script"
    exit 1
fi

# Get current context
CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "unknown")
echo -e "Current Kubernetes context: ${YELLOW}${CURRENT_CONTEXT}${NC}"
echo ""

# Warning message
echo -e "${RED}WARNING: This will destroy all OpenTelemetry Demo resources!${NC}"
echo ""

# Ask for confirmation
read -p "Are you sure you want to destroy the deployment? (yes/no): " -r
echo ""
if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
    echo "Destroy cancelled"
    exit 0
fi

# Run terraform destroy
echo "Destroying OpenTelemetry Demo resources..."
terraform destroy

echo ""
echo -e "${GREEN}Destruction complete!${NC}"
echo ""
