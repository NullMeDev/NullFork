#!/bin/bash

# Enhanced Gateway Scraper Stop Script
# This script stops the background scraper service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🛑 Stopping Enhanced Gateway Scraper..."

# Check if PID file exists
if [ -f "./logs/scraper.pid" ]; then
    PID=$(cat ./logs/scraper.pid)
    
    if ps -p $PID > /dev/null 2>&1; then
        echo "   - Stopping process with PID: $PID"
        kill $PID
        
        # Wait for process to stop
        sleep 2
        
        if ps -p $PID > /dev/null 2>&1; then
            echo "   - Force killing process..."
            kill -9 $PID
        fi
        
        # Remove PID file
        rm -f ./logs/scraper.pid
        echo "✅ Gateway Scraper stopped successfully!"
    else
        echo "❌ Process with PID $PID not found"
        rm -f ./logs/scraper.pid
    fi
else
    # Try to find and kill any running gateway-scraper processes
    if pgrep -f "gateway-scraper" > /dev/null; then
        echo "   - Found running gateway-scraper processes"
        pkill -f "gateway-scraper"
        echo "✅ Gateway Scraper stopped successfully!"
    else
        echo "❌ No running Gateway Scraper processes found"
    fi
fi
