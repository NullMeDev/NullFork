# Enhanced Gateway Scraper & Checker

A high-performance, unified gateway scraper and checker built by integrating the best features from NullVectorBeta and universal-checker. This tool combines concurrent proxy harvesting, validation, payment gateway detection, and account checking into one comprehensive solution.

## 🚀 Features

### 🌐 Proxy Management
- **Concurrent proxy harvesting** from multiple sources (HTTP, HTTPS, SOCKS4, SOCKS5)
- **Real-time proxy validation** with anonymity and latency checking
- **Smart proxy rotation** and health monitoring
- **Auto-scraping** from 50+ reliable proxy sources
- **Proxy quality scoring** and automatic filtering

### 💳 Payment Gateway Detection
- **Advanced payment gateway detection** for Stripe, PayPal, Square, Authorize.Net, and 30+ more
- **Domain scanning** with intelligent website analysis
- **JavaScript rendering** for dynamic payment gateway detection
- **Rate-limited scanning** to respect target servers
- **Comprehensive reporting** with detailed findings

### 🔐 Account Checking
- **Multi-format config support** (.opk, .svb, .loli)
- **High-performance checking** with 500+ CPM capability
- **Smart retry logic** with exponential backoff
- **Account validation** across multiple platforms
- **Live statistics** and progress tracking

### 📊 Data Management
- **ClickHouse integration** for high-performance data storage
- **Real-time analytics** and reporting
- **Data export** in multiple formats (JSON, CSV, TXT)
- **Historical tracking** and trend analysis
- **Backup and restore** functionality

### 🔧 Advanced Features
- **Discord integration** for automated reporting
- **Health check endpoints** for monitoring
- **Structured logging** with configurable levels
- **Docker deployment** with compose setup
- **GUI interface** with real-time updates
- **Command-line interface** for automation

## 🏗️ Architecture

```
enhanced-gateway-scraper/
├── cmd/
│   ├── scraper/          # Main scraper application
│   ├── gui/              # GUI application
│   ├── cli/              # CLI tools
│   └── checker/          # Account checker
├── internal/
│   ├── proxy/            # Proxy management
│   ├── gateway/          # Payment gateway detection
│   ├── checker/          # Account checking engine
│   ├── database/         # Database integrations
│   ├── config/           # Configuration management
│   ├── discord/          # Discord integration
│   ├── logger/           # Structured logging
│   ├── health/           # Health monitoring
│   └── scheduler/        # Task scheduling
├── pkg/
│   ├── types/            # Common types
│   ├── utils/            # Utility functions
│   └── validators/       # Validation functions
├── web/                  # Web interface
├── configs/              # Configuration files
├── docker/               # Docker configuration
└── scripts/              # Deployment scripts
```

## 🚀 Quick Start

### Method 1: Docker (Recommended)
```bash
git clone https://github.com/0CoolDev/enhanced-gateway-scraper
cd enhanced-gateway-scraper
docker-compose up -d --build
```

### Method 2: Go Build
```bash
go mod tidy
go build -o gateway-scraper cmd/scraper/main.go
./gateway-scraper
```

### Method 3: GUI Mode
```bash
go build -o gateway-scraper-gui cmd/gui/main.go
./gateway-scraper-gui
```

## 📋 Configuration

Create a `.env` file:
```env
# Database
CLICKHOUSE_DSN=localhost:9000

# Discord Integration
DISCORD_WEBHOOK=https://discord.com/api/webhooks/your-webhook

# Intervals
PROXY_FETCH_INTERVAL=15m
PROXY_VALIDATE_INTERVAL=30m
GATEWAY_SCAN_INTERVAL=2h
CHECKER_INTERVAL=1h

# Performance
MAX_WORKERS=200
PROXY_TIMEOUT=10s
REQUEST_TIMEOUT=30s
BATCH_SIZE=1000

# Proxy Sources
AUTO_SCRAPE_PROXIES=true
PROXY_QUALITY_THRESHOLD=80
MAX_PROXY_AGE=1h

# Gateway Detection
GATEWAY_SCAN_DEPTH=3
MAX_DOMAINS_PER_SCAN=5000
JAVASCRIPT_RENDERING=true

# Account Checking
CONFIG_FORMATS=opk,svb,loli
MAX_CPM_TARGET=1000
RETRY_ATTEMPTS=3

# Logging
LOG_LEVEL=INFO
LOG_FORMAT=JSON
```

## 🎯 Usage

### Proxy Management
```bash
# Auto-scrape and validate proxies
./gateway-scraper proxy scrape --validate --sources=all

# Check specific proxy list
./gateway-scraper proxy check --file=proxies.txt --output=working.txt

# Monitor proxy health
./gateway-scraper proxy monitor --interval=5m
```

### Gateway Detection
```bash
# Scan top domains for payment gateways
./gateway-scraper gateway scan --domains=1000 --depth=3

# Scan specific websites
./gateway-scraper gateway check --urls=urls.txt --output=gateways.json

# Generate gateway report
./gateway-scraper gateway report --format=json --since=24h
```

### Account Checking
```bash
# Check accounts with config
./gateway-scraper check --config=login.opk --combos=accounts.txt --workers=100

# Multi-config checking
./gateway-scraper check --configs=*.opk --combos=combos.txt --output=results/

# Live checking with GUI
./gateway-scraper check --gui --real-time
```

### Combined Operations
```bash
# Full suite operation
./gateway-scraper run --mode=all --auto-scrape --validate --scan --check

# Scheduled operations
./gateway-scraper schedule --proxy-interval=30m --gateway-interval=2h --check-interval=1h
```

## 📊 API Endpoints

### Health & Metrics
- `GET /health` - Application health status
- `GET /metrics` - Performance metrics
- `GET /stats` - Real-time statistics

### Proxy Management
- `GET /api/v1/proxies` - List all proxies
- `POST /api/v1/proxies/scrape` - Trigger proxy scraping
- `PUT /api/v1/proxies/validate` - Validate proxy list
- `GET /api/v1/proxies/stats` - Proxy statistics

### Gateway Detection
- `GET /api/v1/gateways` - List detected gateways
- `POST /api/v1/gateways/scan` - Start gateway scan
- `GET /api/v1/gateways/report` - Generate report
- `GET /api/v1/gateways/domains/{domain}` - Domain-specific results

### Account Checking
- `POST /api/v1/check/start` - Start checking session
- `GET /api/v1/check/status` - Check session status
- `GET /api/v1/check/results` - Get checking results
- `POST /api/v1/check/stop` - Stop checking session

## 🔧 Advanced Configuration

### Custom Proxy Sources
```yaml
proxy_sources:
  - name: "Custom Source 1"
    url: "https://api.proxyscrape.com/v2/?request=get&protocol=http"
    type: "http"
    format: "ip:port"
    reliability: 0.8
  
  - name: "Premium Proxies"
    url: "https://premium-proxy-api.com/list"
    type: "socks5"
    auth_required: true
    quality_threshold: 90
```

### Gateway Detection Rules
```yaml
gateway_rules:
  stripe:
    patterns:
      - "stripe.com"
      - "js.stripe.com"
      - "stripe-public-key"
    confidence: 0.95
    
  paypal:
    patterns:
      - "paypal.com"
      - "paypal-checkout"
      - "paypal-button"
    confidence: 0.90
```

### Checker Configurations
```yaml
checker_profiles:
  fast:
    workers: 500
    timeout: 10s
    retries: 1
    
  balanced:
    workers: 200
    timeout: 20s
    retries: 3
    
  thorough:
    workers: 50
    timeout: 60s
    retries: 5
```

## 📈 Performance Optimization

### High-Performance Setup
- **Workers**: 200-500 (depending on system resources)
- **Memory**: 4GB+ RAM recommended
- **CPU**: Multi-core processor (8+ cores optimal)
- **Network**: High-bandwidth connection for proxy scraping
- **Storage**: SSD recommended for database operations

### Scaling Guidelines
- Horizontal scaling with multiple instances
- Load balancing across worker pools
- Database sharding for large datasets
- CDN integration for web interface

## 🐳 Docker Deployment

```yaml
version: '3.8'
services:
  gateway-scraper:
    build: .
    ports:
      - "8080:8080"
      - "8123:8123"
    environment:
      - CLICKHOUSE_DSN=clickhouse:9000
      - MAX_WORKERS=200
    depends_on:
      - clickhouse
      - redis
  
  clickhouse:
    image: clickhouse/clickhouse-server:latest
    ports:
      - "9000:9000"
    volumes:
      - ./data/clickhouse:/var/lib/clickhouse
  
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
```

## 🔐 Security

- **Rate limiting** to prevent abuse
- **Input validation** for all user inputs
- **Secure configuration** storage
- **Network isolation** with Docker
- **Access logging** for audit trails
- **API authentication** with tokens

## 📊 Monitoring & Alerts

### Built-in Monitoring
- Real-time performance metrics
- Health check endpoints
- Resource utilization tracking
- Error rate monitoring
- Success rate analytics

### Discord Alerts
- System health notifications
- Performance threshold alerts
- Error notifications
- Daily/weekly reports
- Custom alert rules

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **NullVectorBeta** - Proxy scraping and gateway detection foundations
- **universal-checker** - High-performance checking architecture
- Community contributors and testers

---

**Built with ❤️ by 0CoolDev**
