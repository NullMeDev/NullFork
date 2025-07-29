#!/bin/bash

# Enhanced Gateway Scraper Background Service Script
# This script runs the scraper as a background service with proper logging

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🚀 Starting Enhanced Gateway Scraper Background Service..."

# Set production environment variables
export CLICKHOUSE_DSN="localhost:9001"
export API_PORT="8082"
export PRODUCTION_MODE="true"
export GIN_MODE="release"
export LOG_LEVEL="INFO"
export MAX_WORKERS="100"
export ENABLE_METRICS="true"
export METRICS_PORT="9090"

# Ensure required directories exist
mkdir -p ./data ./results ./backups ./logs

# Set proper permissions
chmod +x ./gateway-scraper
chmod 755 ./data ./results ./backups ./logs

# Check if already running
if pgrep -f "gateway-scraper" > /dev/null; then
    echo "❌ Gateway Scraper is already running!"
    echo "   To stop it, run: ./stop-scraper.sh"
    exit 1
fi

# Start the scraper in background
echo "🏃 Starting Enhanced Gateway Scraper in background..."
echo "   - API Server: http://localhost:8082"
echo "   - Metrics: http://localhost:9090"
echo "   - ClickHouse: localhost:9001"
echo "   - Redis: localhost:6379"
echo "   - Logs: ./logs/scraper.log"

nohup ./gateway-scraper > ./logs/scraper.log 2>&1 &
SCRAPER_PID=$!

# Save PID for stopping later
echo $SCRAPER_PID > ./logs/scraper.pid

echo "✅ Gateway Scraper started successfully!"
echo "   - PID: $SCRAPER_PID"
echo "   - Log file: ./logs/scraper.log"
echo "   - PID file: ./logs/scraper.pid"
echo ""
echo "To stop the scraper, run: ./stop-scraper.sh"
echo "To view logs, run: tail -f ./logs/scraper.log"
