package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Configure logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	
	logrus.Info("Starting Web GUI Service...")

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

	// Start server
	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		logrus.Info("Starting GUI server on port 8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start GUI server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down GUI server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("GUI service stopped")
}
