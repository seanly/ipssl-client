package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ipssl-client/internal/config"
	"ipssl-client/internal/ipssl"
	"ipssl-client/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	logger := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	// Create IPSSL client
	client, err := ipssl.NewClient(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create IPSSL client", "error", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, stopping...")
		cancel()
	}()

	// Start the IPSSL client
	logger.Info("Starting IPSSL client", "client_ip", cfg.ClientIP)
	if err := client.Start(ctx); err != nil {
		logger.Fatal("IPSSL client failed", "error", err)
	}
}
