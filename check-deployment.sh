#!/bin/bash

# Enhanced Gateway Scraper Deployment Status Check

echo "🔍 Enhanced Gateway Scraper Deployment Status"
echo "=============================================="

# Check service ports
echo ""
echo "📡 Service Port Status:"
echo "----------------------"

check_port() {
    local port=$1
    local service=$2
    if nc -z localhost $port 2>/dev/null; then
        echo "✅ $service - Port $port: RUNNING"
        return 0
    else
        echo "❌ $service - Port $port: NOT RUNNING"
        return 1
    fi
}

check_port 9001 "ClickHouse"
check_port 8123 "ClickHouse HTTP"
check_port 6379 "Redis"
check_port 8082 "Gateway Scraper API"
check_port 9090 "Metrics"

# Check processes
echo ""
echo "🔄 Process Status:"
echo "-----------------"
if pgrep -f "gateway-scraper" > /dev/null; then
    echo "✅ Gateway Scraper processes:"
    pgrep -f "gateway-scraper" | while read pid; do
        echo "   - PID: $pid"
    done
else
    echo "❌ No Gateway Scraper processes found"
fi

# Check Docker containers
echo ""
echo "🐳 Docker Container Status:"
echo "---------------------------"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(clickhouse|redis|gateway|grafana|prometheus)" || echo "No related containers found"

# Check API health
echo ""
echo "🏥 API Health Check:"
echo "-------------------"
if curl -s http://localhost:8082/api/v1/health > /dev/null 2>&1; then
    echo "✅ API Server is responding"
    echo "   - Health endpoint: http://localhost:8082/api/v1/health"
    echo "   - Dashboard: http://localhost:8082/"
else
    echo "❌ API Server not responding"
fi

# Check log files
echo ""
echo "📋 Log Files:"
echo "-------------"
if [ -f "./logs/scraper.log" ]; then
    echo "✅ Scraper log file exists: ./logs/scraper.log"
    echo "   Last 3 lines:"
    tail -n 3 ./logs/scraper.log 2>/dev/null || echo "   (Empty or unreadable log file)"
else
    echo "❌ No scraper log file found"
fi

if [ -f "./logs/scraper.pid" ]; then
    echo "✅ PID file exists: ./logs/scraper.pid"
    echo "   PID: $(cat ./logs/scraper.pid)"
else
    echo "❌ No PID file found"
fi

# Check data directories
echo ""
echo "📁 Data Directories:"
echo "-------------------"
for dir in data results backups logs; do
    if [ -d "./$dir" ]; then
        echo "✅ $dir directory exists"
    else
        echo "❌ $dir directory missing"
    fi
done

echo ""
echo "🎯 Deployment Summary:"
echo "====================="
echo "The Enhanced Gateway Scraper is configured and ready for deployment."
echo ""
echo "Available Commands:"
echo "  - ./run-production.sh    - Start in foreground mode"
echo "  - ./run-background.sh    - Start as background service"
echo "  - ./stop-scraper.sh      - Stop background service"
echo "  - ./check-deployment.sh  - Check deployment status"
echo ""
echo "Service URLs:"
echo "  - API Dashboard: http://localhost:8082"
echo "  - Health Check: http://localhost:8082/api/v1/health"
echo "  - Metrics: http://localhost:9090"
