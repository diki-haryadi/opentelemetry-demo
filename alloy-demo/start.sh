#!/bin/bash

# Quick start script for Alloy Demo

echo "🚀 Starting Alloy Demo Observability Stack..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Start the stack
echo "📦 Starting services with Docker Compose..."
docker-compose up -d

# Wait for services to be ready
echo ""
echo "⏳ Waiting for services to start..."
sleep 10

# Check service health
echo ""
echo "🔍 Checking service status..."
docker-compose ps

echo ""
echo "✅ Alloy Demo is starting up!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Access Points:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎨 Grafana Dashboards:"
echo "   http://localhost:3000"
echo "   - Service Graph: http://localhost:3000/d/service-graph"
echo "   - App Metrics:   http://localhost:3000/d/app-metrics"
echo ""
echo "🔧 Observability Stack:"
echo "   - Alloy UI:      http://localhost:12345"
echo "   - Prometheus:    http://localhost:9090"
echo "   - Tempo:         http://localhost:3200"
echo "   - Loki:          http://localhost:3100"
echo ""
echo "🌐 Applications:"
echo "   - Frontend:      http://localhost:8080"
echo "   - Backend:       http://localhost:8081/health"
echo "   - Payment:       http://localhost:8082/health"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 Next Steps:"
echo "   1. Wait 2-3 minutes for all services to fully initialize"
echo "   2. Generate traffic: ./load-generator.sh"
echo "   3. View dashboards in Grafana"
echo ""
echo "📝 View logs: docker-compose logs -f [service-name]"
echo "🛑 Stop stack: docker-compose down"
echo ""
