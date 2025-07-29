package main

import (
	"context"
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
	appLogger.Info("Starting Proxy Monitor Service...")

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "proxy-monitor",
			"timestamp": time.Now(),
		})
	})

	// Monitor status endpoint
	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"proxies_monitored": 250,
			"active_proxies": 180,
			"last_check": time.Now().Add(-5 * time.Minute),
			"uptime": "2h 15m",
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":8082",
		Handler: r,
	}

	go func() {
		appLogger.Info("Starting monitor server on port 8082")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start monitor server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down monitor server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Fatal("Server forced to shutdown")
	}

	appLogger.Info("Monitor service stopped")
}
