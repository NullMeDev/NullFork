package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"enhanced-gateway-scraper/internal/api"
	"enhanced-gateway-scraper/internal/config"
	"enhanced-gateway-scraper/internal/database"
	"enhanced-gateway-scraper/internal/discord"
	"enhanced-gateway-scraper/internal/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.InitLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogIncludeCaller)
	appLogger.Info("Starting Enhanced Gateway Scraper...")

	// Initialize Discord client
	discordClient := discord.NewClient(&cfg.Discord)
	if discordClient != nil {
		appLogger.Info("Discord integration enabled")
		discordClient.SendNotification("🚀 Enhanced Gateway Scraper starting up...")
	}

	// Initialize database connection
	var dbClient *database.ClickHouseClient
	if cfg.Database.DSN != "" {
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
