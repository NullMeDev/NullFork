package api

import (
	"time"
	"github.com/gin-gonic/gin"
	"net/http"
	"enhanced-gateway-scraper/internal/middleware"
	"enhanced-gateway-scraper/internal/config"
)

func StartAPIServer(port string, cfg *config.Config) {
	r := gin.Default()

	// Apply middleware
	r.Use(middleware.RateLimiter(cfg.DefaultRequestDelay))
	r.Use(middleware.UserAgentRotator(cfg.UserAgents))

	// Define routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler)
		v1.GET("/metrics", metricsHandler)
		v1.GET("/stats", statsHandler)
		v1.GET("/proxies", proxiesHandler)
		v1.POST("/proxies/scrape", scrapeProxiesHandler)
		v1.GET("/gateways", gatewaysHandler)
		v1.POST("/gateways/scan", scanGatewaysHandler)
		v1.POST("/check/start", startCheckHandler)
		v1.GET("/check/status", checkStatusHandler)
		v1.POST("/check/stop", stopCheckHandler)
	}

	// Serve static files for GUI
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Enhanced Gateway Scraper",
		})
	})

	// Start the server
	r.Run(port)
}

// API Handlers
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    "12h 30m",
		"version":   "1.0.0",
	})
}

func metricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"proxies_total":    1250,
		"proxies_working":  987,
		"gateways_found":   45,
		"checks_completed": 2340,
		"success_rate":     0.87,
		"timestamp":        time.Now(),
	})
}

func statsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_scans":      152,
		"active_workers":   100,
		"current_cpm":      450.5,
		"errors_24h":       23,
		"last_scan":        time.Now().Add(-15 * time.Minute),
	})
}

func proxiesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"proxies": []gin.H{
			{"host": "192.168.1.1", "port": 8080, "type": "http", "quality": 95, "latency": 120},
			{"host": "192.168.1.2", "port": 3128, "type": "https", "quality": 88, "latency": 200},
			{"host": "192.168.1.3", "port": 1080, "type": "socks5", "quality": 92, "latency": 150},
		},
		"total": 987,
	})
}

func scrapeProxiesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Proxy scraping started",
		"job_id":  "scrape_12345",
		"status":  "initiated",
	})
}

func gatewaysHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"gateways": []gin.H{
			{"domain": "example.com", "gateway": "stripe", "confidence": 0.95, "detected_at": time.Now()},
			{"domain": "shop.test.com", "gateway": "paypal", "confidence": 0.88, "detected_at": time.Now()},
			{"domain": "payment.example.org", "gateway": "square", "confidence": 0.92, "detected_at": time.Now()},
		},
		"total": 45,
	})
}

func scanGatewaysHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Gateway scanning started",
		"job_id":  "scan_67890",
		"status":  "initiated",
	})
}

func startCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":    "Account checking started",
		"session_id": "check_session_abc123",
		"status":     "running",
	})
}

func checkStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"session_id":     "check_session_abc123",
		"status":         "running",
		"progress":       45.2,
		"valid_accounts": 23,
		"total_checked":  150,
		"current_cpm":    320.5,
	})
}

func stopCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Account checking stopped",
		"status":  "stopped",
	})
}
