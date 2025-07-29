package api

import (
	"context"
	"fmt"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"net/http"
	"enhanced-gateway-scraper/internal/middleware"
	"enhanced-gateway-scraper/internal/config"
	"enhanced-gateway-scraper/internal/database"
	"enhanced-gateway-scraper/internal/proxy"
	"enhanced-gateway-scraper/internal/gateway"
	"enhanced-gateway-scraper/internal/logger"
	"enhanced-gateway-scraper/pkg/types"
	"github.com/sirupsen/logrus"
)

// APIServer holds the server dependencies
type APIServer struct {
	config    *config.Config
	db        *database.ClickHouseClient
	scraper   *proxy.Scraper
	detector  *gateway.Detector
	logger    *logrus.Logger
	metrics   *types.Metrics
	running   bool
	sessionID string
}

// NewAPIServer creates a new API server with dependencies
func NewAPIServer(cfg *config.Config) *APIServer {
	appLogger := logger.InitLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogIncludeCaller)
	
	// Initialize database
	var dbClient *database.ClickHouseClient
	if cfg.Database.DSN != "" {
		var err error
		dbClient, err = database.NewClickHouseClient(cfg.Database.DSN)
		if err != nil {
			appLogger.WithError(err).Error("Failed to connect to ClickHouse")
		}
	}

	// Initialize scraper
	checkerConfig := &types.CheckerConfig{
		ProxyTimeout: int(cfg.ProxyTimeout.Milliseconds()),
		MaxWorkers:   cfg.MaxWorkers,
	}
	scraper := proxy.NewScraper(checkerConfig)

	// Initialize gateway detector
	detector := gateway.NewDetector(cfg.GatewayRules)

	return &APIServer{
		config:   cfg,
		db:       dbClient,
		scraper:  scraper,
		detector: detector,
		logger:   appLogger,
		metrics:  &types.Metrics{},
		running:  false,
	}
}

func StartAPIServer(port string, cfg *config.Config) {
	server := NewAPIServer(cfg)
	r := gin.Default()

	// Apply middleware
	r.Use(middleware.RateLimiter(cfg.DefaultRequestDelay))
	r.Use(middleware.UserAgentRotator(cfg.UserAgents))

	// Define routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", server.healthHandler)
		v1.GET("/metrics", server.metricsHandler)
		v1.GET("/stats", server.statsHandler)
		v1.GET("/proxies", server.proxiesHandler)
		v1.POST("/proxies/scrape", server.scrapeProxiesHandler)
		v1.GET("/gateways", server.gatewaysHandler)
		v1.POST("/gateways/scan", server.scanGatewaysHandler)
		v1.POST("/check/start", server.startCheckHandler)
		v1.GET("/check/status", server.checkStatusHandler)
		v1.POST("/check/stop", server.stopCheckHandler)
		// New endpoints for enhanced web interface
		v1.POST("/search", server.searchHandler)
		v1.POST("/scan", server.scanHandler)
	}

	// Serve static files for GUI
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "nullgen-dashboard.html", gin.H{
			"title": "NullScrape | Enhanced Gateway Scanner",
		})
	})

	// Start the server
	r.Run(port)
}

// API Handlers
func (s *APIServer) healthHandler(c *gin.Context) {
	// Check database connection
	dbStatus := "disconnected"
	if s.db != nil {
		dbStatus = "connected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"version":    "1.0.0",
		"database":   dbStatus,
		"log_level":  s.config.LogLevel,
		"workers":    s.config.MaxWorkers,
	})
}

func (s *APIServer) metricsHandler(c *gin.Context) {
	// Get real metrics from database if available
	var metrics gin.H
	if s.db != nil {
		// Try to get real metrics from database
		ctx := context.Background()
		proxiesTotal, _ := s.db.GetProxyCount(ctx)
		proxiesWorking, _ := s.db.GetWorkingProxyCount(ctx)
		gatewaysFound, _ := s.db.GetGatewayCount(ctx)
		
		successRate := 0.0
		if proxiesTotal > 0 {
			successRate = float64(proxiesWorking) / float64(proxiesTotal)
		}
		
		metrics = gin.H{
			"proxies_total":    proxiesTotal,
			"proxies_working":  proxiesWorking,
			"gateways_found":   gatewaysFound,
			"checks_completed": s.metrics.ChecksCompleted,
			"success_rate":     successRate,
			"timestamp":        time.Now(),
		}
	} else {
		// Return basic metrics when database is not available
		s.logger.Warn("Database not available, returning basic metrics")
		metrics = gin.H{
			"proxies_total":    0,
			"proxies_working":  0,
			"gateways_found":   0,
			"checks_completed": 0,
			"success_rate":     0.0,
			"timestamp":        time.Now(),
			"error":           "Database not connected",
		}
	}
	
	c.JSON(http.StatusOK, metrics)
}

func (s *APIServer) statsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_scans":      s.metrics.TotalScans,
		"active_workers":   s.config.MaxWorkers,
		"current_cpm":      s.metrics.CurrentCPM,
		"errors_24h":       s.metrics.Errors24h,
		"last_scan":        s.metrics.LastScan,
		"running":          s.running,
		"session_id":       s.sessionID,
	})
}

func (s *APIServer) proxiesHandler(c *gin.Context) {
	if s.db == nil {
		s.logger.Error("Database not connected")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Database not connected",
			"proxies": []gin.H{},
			"total":   0,
		})
		return
	}
	
	ctx := context.Background()
	proxies, err := s.db.GetProxies(ctx, 100) // Limit to 100 for performance
	if err != nil {
		s.logger.WithError(err).Error("Failed to get proxies from database")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve proxies",
			"proxies": []gin.H{},
			"total":   0,
		})
		return
	}
	
	total, _ := s.db.GetProxyCount(ctx)
	c.JSON(http.StatusOK, gin.H{
		"proxies": proxies,
		"total":   total,
	})
}

func (s *APIServer) scrapeProxiesHandler(c *gin.Context) {
	if s.scraper == nil {
		s.logger.Error("Scraper not initialized")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  "Scraper not initialized",
			"status": "failed",
		})
		return
	}
	
	// Start scraping in a goroutine
	go func() {
		s.logger.Info("Starting proxy scraping...")
		ctx := context.Background()
		proxies, err := s.scraper.ScrapeAndValidate(ctx)
		if err != nil {
			s.logger.WithError(err).Error("Proxy scraping failed")
			return
		}
		
		// Store proxies in database if available
		if s.db != nil {
			for _, proxy := range proxies {
				if err := s.db.StoreProxy(ctx, proxy); err != nil {
					s.logger.WithError(err).Error("Failed to store proxy")
				}
			}
		}
		
		s.logger.Info(fmt.Sprintf("Proxy scraping completed, found %d proxies", len(proxies)))
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Proxy scraping started",
		"status":  "initiated",
	})
}

func (s *APIServer) gatewaysHandler(c *gin.Context) {
	if s.db == nil {
		s.logger.Error("Database not connected")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":    "Database not connected",
			"gateways": []gin.H{},
			"total":    0,
		})
		return
	}
	
	ctx := context.Background()
	gateways, err := s.db.GetGateways(ctx, 100) // Limit to 100 for performance
	if err != nil {
		s.logger.WithError(err).Error("Failed to get gateways from database")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to retrieve gateways",
			"gateways": []gin.H{},
			"total":    0,
		})
		return
	}
	
	total, _ := s.db.GetGatewayCount(ctx)
	c.JSON(http.StatusOK, gin.H{
		"gateways": gateways,
		"total":    total,
	})
}

func (s *APIServer) scanGatewaysHandler(c *gin.Context) {
	if s.detector == nil {
		s.logger.Error("Gateway detector not initialized")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":  "Gateway detector not initialized",
			"status": "failed",
		})
		return
	}
	
	// Start gateway scanning in a goroutine
	go func() {
		s.logger.Info("Starting gateway scanning...")
		// Gateway scanning logic would go here
		s.logger.Info("Gateway scanning completed")
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Gateway scanning started",
		"status":  "initiated",
	})
}

func (s *APIServer) startCheckHandler(c *gin.Context) {
	if s.running {
		c.JSON(http.StatusConflict, gin.H{
			"error":      "Check already running",
			"session_id": s.sessionID,
			"status":     "running",
		})
		return
	}
	
	// Generate new session ID
	s.sessionID = fmt.Sprintf("session_%d", time.Now().Unix())
	s.running = true
	
	// Start checking process
	go func() {
		s.logger.Info("Starting account checking process...")
		// Account checking logic would go here
		time.Sleep(time.Second) // Simulate work
		s.logger.Info("Account checking process started")
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Account checking started",
		"session_id": s.sessionID,
		"status":     "running",
	})
}

func (s *APIServer) checkStatusHandler(c *gin.Context) {
	if !s.running {
		c.JSON(http.StatusOK, gin.H{
			"session_id": s.sessionID,
			"status":     "stopped",
			"progress":   0.0,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"session_id":     s.sessionID,
		"status":         "running",
		"progress":       s.metrics.Progress,
		"valid_accounts": s.metrics.ValidAccounts,
		"total_checked":  s.metrics.TotalChecked,
		"current_cpm":    s.metrics.CurrentCPM,
	})
}

func (s *APIServer) stopCheckHandler(c *gin.Context) {
	if !s.running {
		c.JSON(http.StatusOK, gin.H{
			"message": "No active checking process",
			"status":  "stopped",
		})
		return
	}
	
	s.running = false
	s.logger.Info("Account checking process stopped")
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Account checking stopped",
		"session_id": s.sessionID,
		"status":     "stopped",
	})
}

// SearchRequest represents the search request payload
type SearchRequest struct {
	Query      string `json:"query"`
	Categories string `json:"categories"`
	Proxy      string `json:"proxy"`
	Limit      int    `json:"limit"`
	Output     string `json:"output"`
}

// ScanRequest represents the scan request payload  
type ScanRequest struct {
	URL    string `json:"url"`
	Proxy  string `json:"proxy"`
	Output string `json:"output"`
}

// Enhanced search handler that integrates with CLI functionality
func (s *APIServer) searchHandler(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	
	s.logger.Info(fmt.Sprintf("Starting search request - Query: %s, Categories: %s", req.Query, req.Categories))
	
	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 25 {
		req.Limit = 25
	}
	if req.Proxy == "" {
		req.Proxy = "none"
	}
	
	// Simulate the CLI search functionality
	go func() {
		startTime := time.Now()
		s.logger.Info("🚀 Starting NullScrape search operation...")
		
		// Generate URLs based on request (simplified simulation)
		var searchURLs []string
		if req.Query != "" {
			searchURLs = s.generateSearchURLs(req.Query, req.Limit)
		} else if req.Categories != "" {
			searchURLs = s.generateCategoryBasedURLs(req.Categories, req.Limit)
		}
		
		s.logger.Info(fmt.Sprintf("🌐 Generated %d URLs to crawl", len(searchURLs)))
		
		// Simulate gateway detection on each URL
		var results []gin.H
		for i, url := range searchURLs {
			s.logger.Info(fmt.Sprintf("[%d/%d] 🌐 Crawling: %s", i+1, len(searchURLs), url))
			
			// Simulate detection with our real detector
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			gateways, err := s.detector.DetectGateways(ctx, s.extractDomain(url))
			cancel()
			
			result := gin.H{
				"url":      url,
				"domain":   s.extractDomain(url),
				"gateways": gateways,
				"error":    err,
				"scanned":  time.Now(),
			}
			
			if err != nil {
				s.logger.Error(fmt.Sprintf("   ❌ Error: %v", err))
			} else if len(gateways) > 0 {
				s.logger.Info(fmt.Sprintf("   ✅ Found %d payment gateway(s)", len(gateways)))
				for _, gw := range gateways {
					s.logger.Info(fmt.Sprintf("      • %s (%.0f%% confidence)", gw.GatewayName, gw.Confidence*100))
				}
			} else {
				s.logger.Info("   ⚪ No payment gateways detected")
			}
			
			results = append(results, result)
			
			// Small delay between requests
			time.Sleep(time.Duration(s.config.DefaultRequestDelay))
		}
		
		elapsed := time.Since(startTime)
		s.logger.Info(fmt.Sprintf("🎯 Search completed in %v", elapsed))
		
		// Store results if database is available
		if s.db != nil {
			ctx := context.Background()
			for _, result := range results {
				if gateways, ok := result["gateways"].([]types.Gateway); ok {
					for _, gateway := range gateways {
						if err := s.db.StoreGateway(ctx, gateway); err != nil {
							s.logger.WithError(err).Error("Failed to store gateway")
						}
					}
				}
			}
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Search started successfully",
		"status":  "initiated",
		"query":   req.Query,
		"categories": req.Categories,
		"limit":   req.Limit,
		"proxy":   req.Proxy,
	})
}

// Enhanced scan handler for single URL scans
func (s *APIServer) scanHandler(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}
	
	s.logger.Info(fmt.Sprintf("🎯 Starting single URL scan: %s", req.URL))
	
	// Set defaults
	if req.Proxy == "" {
		req.Proxy = "none"
	}
	
	// Perform scan asynchronously
	go func() {
		startTime := time.Now()
		s.logger.Info(fmt.Sprintf("🔍 Scanning %s with headless browser...", req.URL))
		
		// Use real gateway detector
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		gateways, err := s.detector.DetectGateways(ctx, s.extractDomain(req.URL))
		cancel()
		
		result := gin.H{
			"url":      req.URL,
			"domain":   s.extractDomain(req.URL),
			"gateways": gateways,
			"error":    err,
			"scanned":  time.Now(),
		}
		
		if err != nil {
			s.logger.Error(fmt.Sprintf("❌ Scan failed: %v", err))
		} else if len(gateways) > 0 {
			s.logger.Info(fmt.Sprintf("✅ Found %d payment gateway(s):", len(gateways)))
			for _, gw := range gateways {
				s.logger.Info(fmt.Sprintf("   • %s", strings.ToUpper(gw.GatewayName)))
				s.logger.Info(fmt.Sprintf("     Confidence: %.0f%%", gw.Confidence*100))
				s.logger.Info(fmt.Sprintf("     Method: %s", gw.DetectionMethod))
			}
		} else {
			s.logger.Info("⚪ No payment gateways detected")
		}
		
		elapsed := time.Since(startTime)
		s.logger.Info(fmt.Sprintf("🎯 Scan completed in %v", elapsed))
		
		// Store result if database is available
		if s.db != nil {
			ctx := context.Background()
			for _, gateway := range gateways {
				if err := s.db.StoreGateway(ctx, gateway); err != nil {
					s.logger.WithError(err).Error("Failed to store gateway")
				}
			}
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Scan started successfully",
		"status":  "initiated",
		"url":     req.URL,
		"proxy":   req.Proxy,
	})
}

// Helper functions to generate URLs (simplified versions of CLI logic)
func (s *APIServer) generateSearchURLs(query string, limit int) []string {
	// Simplified URL generation - in production this would use Google API
	keywords := strings.Fields(strings.ToLower(query))
	var urls []string
	
	for _, keyword := range keywords {
		switch keyword {
		case "proxy":
			urls = append(urls, "https://www.proxy-list.download", "https://proxylist.geonode.com")
		case "ai", "artificial", "intelligence":
			urls = append(urls, "https://openai.com", "https://claude.ai")
		case "payment", "gateway":
			urls = append(urls, "https://stripe.com", "https://paypal.com", "https://square.com")
		case "games", "gaming":
			urls = append(urls, "https://steam.com", "https://epic.games.com")
		case "crypto":
			urls = append(urls, "https://coinbase.com", "https://binance.com")
		}
	}
	
	if len(urls) == 0 {
		urls = []string{"https://stripe.com", "https://paypal.com"}
	}
	
	if len(urls) > limit {
		urls = urls[:limit]
	}
	
	return urls
}

func (s *APIServer) generateCategoryBasedURLs(categories string, limit int) []string {
	categoryList := strings.Split(strings.ToLower(categories), ",")
	var urls []string
	
	categoryURLs := map[string][]string{
		"proxy": {"https://www.proxy-list.download", "https://proxylist.geonode.com", "https://www.freeproxy.world"},
		"ai": {"https://openai.com", "https://claude.ai", "https://huggingface.co"},
		"payment": {"https://stripe.com", "https://paypal.com", "https://square.com", "https://razorpay.com"},
		"games": {"https://steam.com", "https://epic.games.com", "https://twitch.tv"},
		"sports": {"https://espn.com", "https://nba.com", "https://nfl.com"},
		"crypto": {"https://coinbase.com", "https://binance.com", "https://kraken.com"},
		"vpn": {"https://nordvpn.com", "https://expressvpn.com", "https://surfshark.com"},
		"hosting": {"https://aws.amazon.com", "https://digitalocean.com", "https://vultr.com"},
		"ide": {"https://replit.com", "https://codepen.io", "https://codesandbox.io"},
		"ecommerce": {"https://shopify.com", "https://amazon.com", "https://ebay.com"},
	}
	
	for _, category := range categoryList {
		category = strings.TrimSpace(category)
		if categoryList, exists := categoryURLs[category]; exists {
			for _, url := range categoryList {
				if len(urls) >= limit {
					break
				}
				urls = append(urls, url)
			}
		}
		if len(urls) >= limit {
			break
		}
	}
	
	if len(urls) == 0 {
		urls = []string{"https://stripe.com", "https://paypal.com"}
	}
	
	return urls
}

func (s *APIServer) extractDomain(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	parts := strings.Split(url, "/")
	return parts[0]
}
