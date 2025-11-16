#!/bin/bash

# Load Generator for Alloy Demo
# Generates traffic to test the observability stack

echo "🚀 Starting load generator for Alloy Demo..."
echo "Press Ctrl+C to stop"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter
count=0

while true; do
    count=$((count + 1))
    
    # Random delay between requests (100-500ms)
    sleep $(awk -v min=0.1 -v max=0.5 'BEGIN{srand(); print min+rand()*(max-min)}')
    
    # Randomly choose an endpoint
    rand=$((RANDOM % 10))
    
    if [ $rand -lt 5 ]; then
        # 50% - Get products (frontend -> backend -> payment)
        echo -e "${GREEN}[$count]${NC} GET /products ${BLUE}(frontend -> backend -> payment)${NC}"
        curl -s http://localhost:8080/products > /dev/null
    elif [ $rand -lt 8 ]; then
        # 30% - Checkout (frontend -> payment)
        echo -e "${GREEN}[$count]${NC} POST /checkout ${BLUE}(frontend -> payment)${NC}"
        curl -s -X POST http://localhost:8080/checkout > /dev/null
    else
        # 20% - Home page (frontend only)
        echo -e "${GREEN}[$count]${NC} GET / ${BLUE}(frontend)${NC}"
        curl -s http://localhost:8080/ > /dev/null
    fi
    
    # Show stats every 50 requests
    if [ $((count % 50)) -eq 0 ]; then
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${YELLOW}📊 Sent $count requests${NC}"
        echo -e "${YELLOW}🔗 View dashboards at:${NC}"
        echo -e "${YELLOW}   - Service Graph: http://localhost:3000/d/service-graph${NC}"
        echo -e "${YELLOW}   - App Metrics: http://localhost:3000/d/app-metrics${NC}"
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
    fi
done
