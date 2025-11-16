#!/bin/bash

# Script untuk check service graph setup dan data

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔍 Checking Service Graph Setup...${NC}"
echo ""

# 1. Check Tempo metrics
echo -e "${BLUE}1. Checking Tempo Metrics Generator...${NC}"
tempo_metrics=$(curl -s http://localhost:3200/metrics 2>/dev/null | grep -c "traces_service_graph" || echo "0")
if [ "$tempo_metrics" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Tempo metrics generator is active (found $tempo_metrics service graph metrics)"
    curl -s http://localhost:3200/metrics | grep "traces_service_graph" | head -5
else
    echo -e "${RED}✗${NC} No service graph metrics found in Tempo"
    echo "   Check: docker logs tempo | grep metrics_generator"
fi
echo ""

# 2. Check Prometheus metrics
echo -e "${BLUE}2. Checking Prometheus Metrics...${NC}"
prom_result=$(curl -s "http://localhost:9090/api/v1/query?query=traces_service_graph_request_total" 2>/dev/null || echo "")
if echo "$prom_result" | grep -q "result.*client"; then
    echo -e "${GREEN}✓${NC} Service graph metrics found in Prometheus"
    echo "$prom_result" | jq -r '.data.result[] | "   \(.metric.client) -> \(.metric.server): \(.value[1])"' 2>/dev/null || echo "   (Data found but jq not available)"
else
    echo -e "${RED}✗${NC} No service graph metrics in Prometheus"
    echo "   Check: docker logs tempo | grep remote_write"
    echo "   Check: docker logs prometheus | grep remote"
fi
echo ""

# 3. Check Mimir metrics
echo -e "${BLUE}3. Checking Mimir Metrics...${NC}"
mimir_result=$(curl -s "http://localhost:9009/prometheus/api/v1/query?query=traces_service_graph_request_total" 2>/dev/null || echo "")
if echo "$mimir_result" | grep -q "result.*client"; then
    echo -e "${GREEN}✓${NC} Service graph metrics found in Mimir"
    echo "$mimir_result" | jq -r '.data.result[] | "   \(.metric.client) -> \(.metric.server): \(.value[1])"' 2>/dev/null || echo "   (Data found but jq not available)"
else
    echo -e "${YELLOW}⚠${NC} No service graph metrics in Mimir (might be using Prometheus)"
fi
echo ""

# 4. Check traces
echo -e "${BLUE}4. Checking Traces in Tempo...${NC}"
traces=$(curl -s "http://localhost:3200/api/search?limit=5" 2>/dev/null || echo "")
if echo "$traces" | grep -q "traces"; then
    trace_count=$(echo "$traces" | jq '.traces | length' 2>/dev/null || echo "?")
    echo -e "${GREEN}✓${NC} Found traces in Tempo (showing $trace_count)"
    echo "$traces" | jq -r '.traces[] | "   Service: \(.rootServiceName), Spans: \(.spanSet.spans | length)"' 2>/dev/null | head -5 || echo "   (Traces found but jq not available)"
else
    echo -e "${RED}✗${NC} No traces found in Tempo"
    echo "   Generate traffic: curl http://localhost:8080/api/users-with-profile/1"
fi
echo ""

# 5. Check service names
echo -e "${BLUE}5. Checking Service Names...${NC}"
services=$(curl -s "http://localhost:3200/api/search?limit=100" 2>/dev/null | jq -r '.traces[].rootServiceName' 2>/dev/null | sort -u || echo "")
if [ -n "$services" ]; then
    echo -e "${GREEN}✓${NC} Services found:"
    echo "$services" | while read -r svc; do
        echo "   - $svc"
    done
else
    echo -e "${YELLOW}⚠${NC} No services found (might need traffic)"
fi
echo ""

# 6. Recommendations
echo -e "${BLUE}📋 Recommendations:${NC}"
echo ""

if [ "$tempo_metrics" -eq 0 ]; then
    echo -e "${YELLOW}1.${NC} Restart Tempo to ensure metrics generator is active:"
    echo "   docker-compose restart tempo"
    echo ""
fi

if [ -z "$prom_result" ] || ! echo "$prom_result" | grep -q "result.*client"; then
    echo -e "${YELLOW}2.${NC} Generate cross-service traffic:"
    echo "   curl http://localhost:8080/api/users-with-profile/1"
    echo "   (Wait 1-2 minutes for metrics to be generated)"
    echo ""
fi

if [ -z "$traces" ] || ! echo "$traces" | grep -q "traces"; then
    echo -e "${YELLOW}3.${NC} Check if services are running:"
    echo "   docker-compose ps"
    echo ""
fi

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "For more details, see: infrastructure/TROUBLESHOOTING.md"

