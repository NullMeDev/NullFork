#!/bin/bash
# NullScrape Diagnostics Mode - Real-time logging and monitoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Diagnostics configuration
LOG_DIR="/tmp/nullscrape-logs"
GUI_LOG="$LOG_DIR/gui.log"
AGENT_LOG="$LOG_DIR/agent.log"
DEBUG_LOG="$LOG_DIR/debug.log"
WEB_PORT=8082
API_PORT=8080

echo -e "${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    NullScrape Diagnostics                    ║"
echo "║                     Real-time Monitoring                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Create log directory
mkdir -p "$LOG_DIR"

# Clean up any existing processes
echo -e "${YELLOW}🧹 Cleaning up existing processes...${NC}"
pkill -f "enhanced-gateway-scraper" 2>/dev/null || true
pkill -f "nullscrape" 2>/dev/null || true
sleep 2

# Function to start GUI service
start_gui() {
    echo -e "${GREEN}🖥️  Starting GUI service (Port: $WEB_PORT)...${NC}"
    cd /home/null/enhanced-gateway-scraper
    WEB_PORT=$WEB_PORT LOG_LEVEL=DEBUG ./dist/gui > "$GUI_LOG" 2>&1 &
    GUI_PID=$!
    echo "GUI PID: $GUI_PID" > "$LOG_DIR/gui.pid"
    
    # Wait for GUI to start
    echo -e "${YELLOW}⏳ Waiting for GUI to initialize...${NC}"
    for i in {1..15}; do
        if curl -s http://localhost:$WEB_PORT/health > /dev/null 2>&1; then
            echo -e "${GREEN}✅ GUI service online at http://localhost:$WEB_PORT${NC}"
            echo -e "${GREEN}🌐 NullGen Dashboard: http://localhost:$WEB_PORT/nullgen${NC}"
            break
        fi
        sleep 1
        if [ $i -eq 15 ]; then
            echo -e "${RED}❌ GUI failed to start within 15 seconds${NC}"
            cat "$GUI_LOG"
            exit 1
        fi
    done
}

# Function to monitor logs in real-time
start_monitoring() {
    echo -e "${CYAN}📊 Starting real-time log monitoring...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop diagnostics mode${NC}"
    echo ""
    
    # Create named pipes for log streaming
    mkfifo "$LOG_DIR/gui_pipe" 2>/dev/null || true
    mkfifo "$LOG_DIR/agent_pipe" 2>/dev/null || true
    
    # Start log monitoring in background
    {
        tail -f "$GUI_LOG" 2>/dev/null | while read line; do
            echo -e "${GREEN}[GUI]${NC} $line"
        done
    } &
    
    {
        tail -f "$AGENT_LOG" 2>/dev/null | while read line; do
            echo -e "${BLUE}[AGENT]${NC} $line"
        done
    } &
    
    {
        tail -f "$DEBUG_LOG" 2>/dev/null | while read line; do
            echo -e "${PURPLE}[DEBUG]${NC} $line"
        done
    } &
}

# Function to display system status
show_status() {
    echo -e "${CYAN}📈 System Status:${NC}"
    echo -e "  GUI Service: ${GREEN}http://localhost:$WEB_PORT${NC}"
    echo -e "  NullGen Dashboard: ${GREEN}http://localhost:$WEB_PORT/nullgen${NC}"
    echo -e "  Health Check: ${GREEN}http://localhost:$WEB_PORT/health${NC}"
    echo ""
    echo -e "${CYAN}📋 Available Commands:${NC}"
    echo -e "  ${YELLOW}./dist/agent search --query=\"stripe.com\" --output=table${NC}"
    echo -e "  ${YELLOW}./dist/agent scan --url=https://paypal.com --proxy=socks5${NC}"
    echo -e "  ${YELLOW}curl http://localhost:$WEB_PORT/health${NC}"
    echo ""
    echo -e "${CYAN}📁 Log Files:${NC}"
    echo -e "  GUI Logs: ${YELLOW}$GUI_LOG${NC}"
    echo -e "  Agent Logs: ${YELLOW}$AGENT_LOG${NC}"
    echo -e "  Debug Logs: ${YELLOW}$DEBUG_LOG${NC}"
    echo ""
}

# Function to run agent with logging
run_agent_with_logging() {
    local cmd="$1"
    echo -e "${BLUE}🤖 Running: $cmd${NC}" | tee -a "$DEBUG_LOG"
    echo "$(date): Agent command: $cmd" >> "$AGENT_LOG"
    
    cd /home/null/enhanced-gateway-scraper
    eval "$cmd" 2>&1 | tee -a "$AGENT_LOG"
    
    echo "$(date): Agent command completed" >> "$AGENT_LOG"
}

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down diagnostics mode...${NC}"
    
    # Kill GUI process
    if [ -f "$LOG_DIR/gui.pid" ]; then
        GUI_PID=$(cat "$LOG_DIR/gui.pid")
        kill $GUI_PID 2>/dev/null || true
        rm -f "$LOG_DIR/gui.pid"
    fi
    
    # Kill any remaining processes
    pkill -f "enhanced-gateway-scraper" 2>/dev/null || true
    pkill -f "tail -f" 2>/dev/null || true
    
    # Clean up pipes
    rm -f "$LOG_DIR/gui_pipe" "$LOG_DIR/agent_pipe" 2>/dev/null || true
    
    echo -e "${GREEN}✅ Cleanup complete${NC}"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Start services
start_gui
sleep 2
show_status

# Initialize log files
echo "$(date): NullScrape Diagnostics Started" > "$DEBUG_LOG"
echo "$(date): GUI service initialized" >> "$GUI_LOG"
echo "$(date): Agent logging initialized" >> "$AGENT_LOG"

# Start log monitoring
start_monitoring

echo -e "${GREEN}🚀 NullScrape Diagnostics Mode Active!${NC}"
echo -e "${CYAN}💡 Open a new terminal and run agent commands to see real-time logs${NC}"
echo ""

# Interactive mode
while true; do
    echo -e "${YELLOW}Choose an option:${NC}"
    echo "1) Test agent scan (stripe.com)"
    echo "2) Test agent search (payment gateways)"
    echo "3) Show system status"
    echo "4) View recent GUI logs"
    echo "5) View recent agent logs"
    echo "6) Test GUI health"
    echo "7) Custom agent command"
    echo "q) Quit diagnostics mode"
    echo ""
    read -p "Enter choice: " choice

    case $choice in
        1)
            run_agent_with_logging "./dist/agent scan --url=https://stripe.com --output=table"
            ;;
        2)
            run_agent_with_logging "./dist/agent search --query=\"payment gateways\" --limit=2 --output=table"
            ;;
        3)
            show_status
            ;;
        4)
            echo -e "${GREEN}📄 Recent GUI logs:${NC}"
            tail -20 "$GUI_LOG" | while read line; do
                echo -e "${GREEN}[GUI]${NC} $line"
            done
            ;;
        5)
            echo -e "${BLUE}📄 Recent agent logs:${NC}"
            tail -20 "$AGENT_LOG" | while read line; do
                echo -e "${BLUE}[AGENT]${NC} $line"
            done
            ;;
        6)
            echo -e "${CYAN}🏥 Testing GUI health...${NC}"
            curl -s http://localhost:$WEB_PORT/health | jq . || echo "Health check failed"
            ;;
        7)
            read -p "Enter agent command: " custom_cmd
            run_agent_with_logging "$custom_cmd"
            ;;
        q|Q)
            cleanup
            ;;
        *)
            echo -e "${RED}Invalid choice. Please try again.${NC}"
            ;;
    esac
    echo ""
done
