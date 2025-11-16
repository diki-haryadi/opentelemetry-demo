#!/bin/bash
# Script to verify the OpenTelemetry infrastructure is running correctly

echo "🔍 Checking OpenTelemetry Infrastructure Status..."

# Check if docker-compose is running
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed"
    exit 1
fi

echo ""
echo "📦 Container Status:"
echo "===================="
docker-compose ps

echo ""
echo "📊 Service Availability Check:"
echo "=============================="

# Check if services are accessible
SERVICES=(
    "http://localhost:3000"  # Grafana
    "http://localhost:8080/health"  # App
    "http://localhost:3200"  # Tempo
    "http://localhost:3100"  # Loki
    "http://localhost:9009"  # Mimir
    "http://localhost:9090"  # Prometheus
    "http://localhost:12345"  # Alloy UI
)

for service in "${SERVICES[@]}"; do
    if curl -f -s --connect-timeout 5 "$service" > /dev/null; then
        echo "✅ $service - Accessible"
    else
        echo "❌ $service - Not accessible"
    fi
done

echo ""
echo "📈 Generating Sample Traffic to Test Data Flow..."
echo "================================================="

# Generate some traffic to create traces/metrics/logs
curl -s http://localhost:8080/api/users > /dev/null
curl -s http://localhost:8080/api/products > /dev/null
curl -s -X POST -H "Content-Type: application/json" -d '{"user_id": 1,"product_id": 1,"quantity": 1}' http://localhost:8080/api/orders > /dev/null

echo "Generated sample requests to the API..."

echo ""
echo "📋 Verification Complete!"
echo "=========================="
echo ""
echo "To access services:"
echo "📍 Grafana Dashboard: http://localhost:3000 (anonymous login enabled)"
echo "📍 API: http://localhost:8080"
echo "📍 API Health Check: http://localhost:8080/health"
echo ""
echo "The infrastructure is ready for use. Data should be flowing through:"
echo "- Application traces → Tempo"
echo "- Application metrics → Mimir/Prometheus" 
echo "- Application logs → Loki"
echo "- All data visualized in Grafana"