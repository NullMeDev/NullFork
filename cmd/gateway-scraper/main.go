package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"enhanced-gateway-scraper/internal/api"
	"enhanced-gateway-scraper/internal/config"
	"enhanced-gateway-scraper/internal/database"
	"enhanced-gateway-scraper/internal/discord"
	"enhanced-gateway-scraper/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Version information - set at build time
var (
	version   = "v1.2.0" // Default version, overridden by build
	buildDate = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Command line flags
	var (
		guiMode = flag.Bool("gui", false, "Run in GUI mode (web interface)")
		version = flag.Bool("version", false, "Show version information")
		help    = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.InitLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogIncludeCaller)
	
	if *guiMode {
		appLogger.Info("Starting Enhanced Gateway Scraper (GUI Mode)...")
		runGUIMode(cfg, appLogger)
	} else {
		appLogger.Info("Starting Enhanced Gateway Scraper (CLI Mode)...")
		runCLIMode(cfg, appLogger)
	}
}

func runCLIMode(cfg *config.Config, appLogger *logrus.Logger) {
	// Initialize Discord client
	discordClient := discord.NewClient(&cfg.Discord)
	if discordClient != nil {
		appLogger.Info("Discord integration enabled")
		discordClient.SendNotification("🚀 Enhanced Gateway Scraper starting up...")
	}

	// Initialize database connection
	var dbClient *database.ClickHouseClient
	if cfg.Database.DSN != "" {
		var err error
		dbClient, err = database.NewClickHouseClient(cfg.Database.DSN)
		if err != nil {
			appLogger.WithError(err).Warn("Failed to connect to ClickHouse, continuing without database")
		} else {
			appLogger.Info("ClickHouse database connected successfully")
		}
	}

	// Create context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start API server in a goroutine
	go func() {
		appLogger.Info(fmt.Sprintf("Starting API server on port %d", cfg.APIPort))
		api.StartAPIServer(fmt.Sprintf(":%d", cfg.APIPort), cfg)
	}()

	// Send startup notification
	if discordClient != nil {
		discordClient.SendNotification("✅ Enhanced Gateway Scraper is now online!")
	}

	// Wait for shutdown signal
	<-ctx.Done()
	appLogger.Info("Shutting down Enhanced Gateway Scraper...")

	// Cleanup
	if dbClient != nil {
		dbClient.Close()
	}

	if discordClient != nil {
		discordClient.SendNotification("🛑 Enhanced Gateway Scraper shutting down...")
	}

	appLogger.Info("Shutdown complete")
}

func runGUIMode(cfg *config.Config, appLogger *logrus.Logger) {
	// Create Gin router
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "gateway-scraper-gui",
			"timestamp": time.Now(),
		})
	})

	// GUI dashboard endpoint
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{
			"title":   "Enhanced Gateway Scraper GUI",
			"version": version,
		})
	})

	// Advanced dashboard endpoint
	r.GET("/advanced", func(c *gin.Context) {
		c.HTML(200, "advanced-dashboard.html", gin.H{
			"title":   "Advanced Dashboard",
			"version": version,
		})
	})

	// Comprehensive dashboard endpoint
	r.GET("/comprehensive", func(c *gin.Context) {
		c.HTML(200, "comprehensive-dashboard.html", gin.H{
			"title":   "Comprehensive Dashboard",
			"version": version,
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		appLogger.Info("Starting GUI server on port 8081")
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

func showHelp() {
	fmt.Println("Enhanced Gateway Scraper")
	fmt.Println("Usage:")
	fmt.Println("  gateway-scraper [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -gui      Run in GUI mode (web interface on port 8081)")
	fmt.Println("  -version  Show version information")
	fmt.Println("  -help     Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gateway-scraper           # Run in CLI mode")
	fmt.Println("  gateway-scraper -gui      # Run in GUI mode")
}

func showVersion() {
	fmt.Printf("NullScrape - Enhanced Gateway Scraper %s\n", version)
	fmt.Printf("Build Date: %s\n", buildDate)
	fmt.Printf("Git Commit: %s\n", gitCommit)
	fmt.Println("Build: Unified CLI/GUI version")
	fmt.Println("")
	fmt.Println("A high-performance web scraping and proxy management tool")
	fmt.Println("for security research and competitive analysis.")
}
