#!/bin/bash

# Enhanced Gateway Scraper Production Deployment Script
# This script ensures all services are running and starts the scraper with proper configuration

set -e

echo "🚀 Starting Enhanced Gateway Scraper in Production Mode..."

# Check if required services are running
echo "📋 Checking service dependencies..."

# Check ClickHouse
if ! nc -z localhost 9001 2>/dev/null; then
    echo "❌ ClickHouse is not running on port 9001"
    echo "🔄 Attempting to start ClickHouse with Docker..."
    sudo docker run -d --name clickhouse-production \
        -p 8123:8123 -p 9001:9000 \
        -e CLICKHOUSE_DB=gateway_scraper \
        -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 \
        --restart unless-stopped \
        clickhouse/clickhouse-server:latest
    
    # Wait for ClickHouse to start
    echo "⏳ Waiting for ClickHouse to start..."
    sleep 10
    
    if ! nc -z localhost 9001 2>/dev/null; then
        echo "❌ Failed to start ClickHouse"
        exit 1
    fi
fi
echo "✅ ClickHouse is running on port 9001"

# Check Redis
if ! nc -z localhost 6379 2>/dev/null && ! nc -z localhost 6380 2>/dev/null; then
    echo "❌ Redis is not running"
    echo "🔄 Attempting to start Redis with Docker..."
    sudo docker run -d --name redis-production \
        -p 6379:6379 \
        --restart unless-stopped \
        redis:7-alpine redis-server --appendonly yes
    
    # Wait for Redis to start
    echo "⏳ Waiting for Redis to start..."
    sleep 5
    
    if ! nc -z localhost 6379 2>/dev/null; then
        echo "❌ Failed to start Redis"
        exit 1
    fi
fi
echo "✅ Redis is running"

# Set production environment variables
echo "🔧 Setting production environment variables..."
export CLICKHOUSE_DSN="localhost:9001"
export API_PORT="8082"
export PRODUCTION_MODE="true"
export GIN_MODE="release"
export LOG_LEVEL="INFO"
export MAX_WORKERS="100"
export ENABLE_METRICS="true"
export METRICS_PORT="9090"

# Ensure required directories exist
echo "📁 Creating required directories..."
mkdir -p ./data
mkdir -p ./results
mkdir -p ./backups
mkdir -p ./logs

# Set proper permissions
echo "🔐 Setting proper permissions..."
chmod +x ./gateway-scraper
chmod 755 ./data ./results ./backups ./logs

# Start the scraper
echo "🏃 Starting Enhanced Gateway Scraper..."
echo "   - API Server: http://localhost:8082"
echo "   - Metrics: http://localhost:9090"
echo "   - ClickHouse: localhost:9001"
echo "   - Redis: localhost:6379"
echo ""
echo "Press Ctrl+C to stop the scraper"
echo ""

# Run the scraper with nohup for background execution if desired
# Uncomment the line below and comment the direct execution for background mode
# nohup ./gateway-scraper > ./logs/scraper.log 2>&1 &

# Direct execution (foreground)
./gateway-scraper
