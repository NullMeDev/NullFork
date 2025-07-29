package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"enhanced-gateway-scraper/internal/config"
	"enhanced-gateway-scraper/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize logger with configuration
	appLogger := logger.InitLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogIncludeCaller)
	appLogger.Info("Starting Web GUI Service...")

	// Create Gin router
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "web-gui",
			"timestamp": time.Now(),
		})
	})

	// GUI dashboard endpoint
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{
			"title": "Enhanced Gateway Scraper GUI",
			"version": "1.0.0",
		})
	})

	// Advanced dashboard endpoint
	r.GET("/advanced", func(c *gin.Context) {
		c.HTML(200, "advanced-dashboard.html", gin.H{
			"title": "Advanced Dashboard",
			"version": "1.0.0",
		})
	})

	// Comprehensive dashboard endpoint
	r.GET("/comprehensive", func(c *gin.Context) {
		c.HTML(200, "comprehensive-dashboard.html", gin.H{
			"title": "Comprehensive Dashboard",
			"version": "1.0.0",
		})
	})

	// NullGen-inspired dashboard
	r.GET("/nullgen", func(c *gin.Context) {
		c.HTML(200, "nullgen-dashboard.html", gin.H{
			"title": "NullScrape - Gateway Scanner",
			"version": "1.0.0",
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.WebPort),
		Handler: r,
	}

	go func() {
		appLogger.WithField("port", cfg.WebPort).Info("Starting GUI server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start GUI server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down GUI server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Fatal("Server forced to shutdown")
	}

	appLogger.Info("GUI service stopped")
}
