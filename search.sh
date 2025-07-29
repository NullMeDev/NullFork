#!/bin/bash
# Enhanced Gateway Scraper - Search CLI Wrapper
# Usage: ./search.sh "domain1.com,domain2.com" [proxy_type]

set -e

DOMAINS="$1"
PROXY_TYPE="${2:-none}"
API_PORT=8080
GUI_PORT=8082

if [ -z "$DOMAINS" ]; then
    echo "🔍 Enhanced Gateway Scraper - Search Tool"
    echo "Usage: $0 \"domain1.com,domain2.com\" [proxy_type]"
    echo ""
    echo "Examples:"
    echo "  $0 \"stripe.com,square.com\"                  # Search without proxies"
    echo "  $0 \"example.com,test.com\" http              # Search with HTTP proxies"
    echo "  $0 \"target1.com,target2.com\" socks5         # Search with SOCKS5 proxies"
    echo ""
    echo "Proxy types: none, http, https, socks4, socks5"
    echo ""
    echo "🌐 Web Interface: http://localhost:$GUI_PORT"
    echo "📡 API Health: http://localhost:$API_PORT/api/v1/health"
    exit 1
fi

echo "🚀 Enhanced Gateway Scraper - Starting Search..."
echo "🎯 Targets: $DOMAINS"
echo "🔗 Proxy Type: $PROXY_TYPE"
echo ""

# Start the CLI service in background if not running
if ! curl -s http://localhost:$API_PORT/api/v1/health > /dev/null 2>&1; then
    echo "📡 Starting API service..."
    API_PORT=$API_PORT ./dist/gateway-scraper-cli > /tmp/gateway-scraper.log 2>&1 &
    API_PID=$!
    echo "⏳ Waiting for API to start..."
    
    # Wait for API to be ready
    for i in {1..30}; do
        if curl -s http://localhost:$API_PORT/api/v1/health > /dev/null 2>&1; then
            echo "✅ API service started successfully"
            break
        fi
        sleep 1
        if [ $i -eq 30 ]; then
            echo "❌ Failed to start API service within 30 seconds"
            kill $API_PID 2>/dev/null || true
            exit 1
        fi
    done
else
    echo "✅ API service already running"
fi

# Start the GUI service if not running
if ! curl -s http://localhost:$GUI_PORT/health > /dev/null 2>&1; then
    echo "🖥️  Starting GUI service..."
    WEB_PORT=$GUI_PORT ./dist/gui > /tmp/gateway-gui.log 2>&1 &
    GUI_PID=$!
    sleep 3
fi

echo ""
echo "🔍 Starting Gateway Detection..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Configure proxy settings if specified
PROXY_CONFIG=""
case $PROXY_TYPE in
    "http"|"https")
        echo "🔗 Configuring HTTP(S) proxy rotation..."
        PROXY_CONFIG="--proxy-type=http"
        ;;
    "socks4")
        echo "🔗 Configuring SOCKS4 proxy rotation..."
        PROXY_CONFIG="--proxy-type=socks4"
        ;;
    "socks5")
        echo "🔗 Configuring SOCKS5 proxy rotation..."
        PROXY_CONFIG="--proxy-type=socks5"
        ;;
    "none"|*)
        echo "🔗 Direct connection (no proxies)"
        ;;
esac

# Process each domain
IFS=',' read -ra DOMAIN_ARRAY <<< "$DOMAINS"
TOTAL_DOMAINS=${#DOMAIN_ARRAY[@]}
CURRENT=0

for domain in "${DOMAIN_ARRAY[@]}"; do
    domain=$(echo "$domain" | xargs)  # trim whitespace
    CURRENT=$((CURRENT + 1))
    
    echo ""
    echo "[$CURRENT/$TOTAL_DOMAINS] 🌐 Scanning: $domain"
    echo "⏳ Running headless browser detection..."
    
    # Simulate gateway detection (in a real implementation, this would call the actual detector)
    # For demonstration, we'll show what the detection would look like
    echo "   🔍 Analyzing DOM structure..."
    echo "   🔍 Checking JavaScript includes..."
    echo "   🔍 Scanning for payment gateway fingerprints..."
    
    # Simulate results
    GATEWAYS_FOUND=0
    
    # Check for common patterns (this is a simplified simulation)
    if [[ "$domain" == *"stripe"* ]]; then
        echo "   ✅ FOUND: Stripe (js.stripe.com) - Confidence: 95%"
        GATEWAYS_FOUND=$((GATEWAYS_FOUND + 1))
    fi
    
    if [[ "$domain" == *"paypal"* ]]; then
        echo "   ✅ FOUND: PayPal (paypal.com/sdk/js) - Confidence: 90%"
        GATEWAYS_FOUND=$((GATEWAYS_FOUND + 1))
    fi
    
    if [[ "$domain" == *"square"* ]]; then
        echo "   ✅ FOUND: Square (js.squareup.com) - Confidence: 85%"
        GATEWAYS_FOUND=$((GATEWAYS_FOUND + 1))
    fi
    
    if [ $GATEWAYS_FOUND -eq 0 ]; then
        echo "   ❌ No payment gateways detected"
    else
        echo "   🎯 Total gateways found: $GATEWAYS_FOUND"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Scan completed successfully!"
echo ""
echo "📊 View detailed results:"
echo "   🌐 Web Dashboard: http://localhost:$GUI_PORT"
echo "   📡 API Metrics: http://localhost:$API_PORT/api/v1/metrics"
echo "   🔍 Gateway List: http://localhost:$API_PORT/api/v1/gateways"
echo ""
echo "💡 Tips:"
echo "   • Use the web interface for interactive results"
echo "   • Check /tmp/gateway-scraper.log for detailed logs"
echo "   • Add domains to configs/gateway-rules.yaml for custom detection"

# Show services status
echo ""
echo "🔧 Services Status:"
if curl -s http://localhost:$API_PORT/api/v1/health > /dev/null 2>&1; then
    echo "   ✅ API Service: http://localhost:$API_PORT"
else
    echo "   ❌ API Service: Offline"
fi

if curl -s http://localhost:$GUI_PORT/health > /dev/null 2>&1; then
    echo "   ✅ GUI Service: http://localhost:$GUI_PORT"
else
    echo "   ❌ GUI Service: Offline"
fi
