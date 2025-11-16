#!/bin/bash

# Script untuk verifikasi setup infrastructure
# Berdasarkan best practices dari OpenTelemetry Demo

set -e

echo "🔍 Verifying Infrastructure Setup..."
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check functions
check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $1 exists"
        return 0
    else
        echo -e "${RED}✗${NC} $1 missing"
        return 1
    fi
}

check_config() {
    if grep -q "$2" "$1" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} $3 found in $1"
        return 0
    else
        echo -e "${RED}✗${NC} $3 not found in $1"
        return 1
    fi
}

errors=0

echo "📁 Checking Configuration Files..."
echo ""

# Check Tempo config
echo "1. Tempo Configuration:"
check_file "tempo/tempo.yaml" || ((errors++))
check_config "tempo/tempo.yaml" "service-graphs" "service-graphs processor" || ((errors++))
check_config "tempo/tempo.yaml" "remote_write" "remote_write configuration" || ((errors++))
check_config "tempo/tempo.yaml" "prometheus:9090" "Prometheus remote_write" || ((errors++))
echo ""

# Check Prometheus config
echo "2. Prometheus Configuration:"
check_file "mimir/prometheus.yml" || ((errors++))
check_config "mimir/prometheus.yml" "tempo:3200" "Tempo scraping" || ((errors++))
echo "   Note: Remote write receiver is configured in docker-compose.yml (--web.enable-remote-write-receiver)"
echo ""

# Check Grafana datasources
echo "3. Grafana Datasources:"
check_file "grafana/datasources.yml" || ((errors++))
check_config "grafana/datasources.yml" "serviceMap" "serviceMap configuration" || ((errors++))
check_config "grafana/datasources.yml" "nodeGraph" "nodeGraph enabled" || ((errors++))
check_config "grafana/datasources.yml" "tempo" "Tempo datasource" || ((errors++))
check_config "grafana/datasources.yml" "prometheus" "Prometheus datasource" || ((errors++))
check_config "grafana/datasources.yml" "mimir" "Mimir datasource" || ((errors++))
check_config "grafana/datasources.yml" "loki" "Loki datasource" || ((errors++))
echo ""

# Check Grafana dashboards
echo "4. Grafana Dashboards:"
check_file "grafana/dashboards.yml" || ((errors++))
check_file "grafana/dashboards/service-tree.json" || ((errors++))
check_config "grafana/dashboards/service-tree.json" "serviceMap" "serviceMap query" || ((errors++))
check_config "grafana/dashboards/service-tree.json" "nodeGraph" "nodeGraph panel" || ((errors++))
echo ""

# Check Alloy config
echo "5. Grafana Alloy Configuration:"
check_file "alloy/config.alloy" || ((errors++))
check_config "alloy/config.alloy" "tempo:4317" "Tempo exporter" || ((errors++))
check_config "alloy/config.alloy" "prometheus.remote_write" "Prometheus remote_write" || ((errors++))
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $errors -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo ""
    echo "✅ Infrastructure setup is correct"
    echo ""
    echo "Next steps:"
    echo "1. Start services: docker-compose up -d"
    echo "2. Generate traffic: curl http://localhost:8080/api/users-with-profile/1"
    echo "3. Check Grafana: http://localhost:3000"
    echo "4. Open Service Graph Dashboard"
    exit 0
else
    echo -e "${RED}✗ Found $errors error(s)${NC}"
    echo ""
    echo "Please fix the errors above before proceeding."
    exit 1
fi

