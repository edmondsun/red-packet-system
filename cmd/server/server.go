package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"syscall"
	"time"

	"red-packet-system/config"
	"red-packet-system/db"
	"red-packet-system/pkg/logger"
	"red-packet-system/redisclient"
	"red-packet-system/routes"
)

func main() {
	log := logger.GetLogger()

	// Load configuration from environment variables
	cfg := config.LoadConfig()

	// Initialize MySQL connection
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	fmt.Println("Database connected successfully")

	// Ensure database connection is closed before shutting down
	defer func() {
		if db.GetDB() != nil {
			log.Println("Closing MySQL connection...")
			if err := db.CloseDB(); err != nil {
				log.Printf("Failed to close database: %v", err)
			} else {
				log.Println("MySQL connection closed")
			}
		}
	}()

	// Initialize Redis connection
	if err := redisclient.InitRedis(cfg); err != nil {
		log.Printf("[WARN] Redis connection failed (initial attempt): %v", err)
	}

	// If `--seed` flag is provided, run database seeding and exit
	if len(os.Args) > 1 && os.Args[1] == "--seed" {
		db.RunSeeds()
		return
	}

	// Set up Gin router
	router := routes.SetupRouter()

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: http.TimeoutHandler(router, 10*time.Second, "Request Timeout"), // Request timeout: 10 seconds
	}

	// Start API server (non-blocking)
	go func() {
		log.Println("API Server is running on port: ", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Capture system signals (CTRL+C, etc.)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal: %v, shutting down server...", sig)

	// Gracefully shutdown with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}
