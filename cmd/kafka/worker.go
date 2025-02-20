package main

import (
	"os"
	"os/signal"
	"syscall"

	"red-packet-system/config"
	"red-packet-system/db"
	"red-packet-system/kafka"
	"red-packet-system/pkg/logger"
)

func main() {
	log := logger.GetLogger()
	log.Println("Starting Kafka Consumer...")

	// Load environment configuration
	cfg := config.LoadConfig()

	// Initialize MySQL connection
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Ensure DB connection is properly initialized
	if db.GetDB() == nil {
		log.Fatal("Database connection is not initialized")
	}

	// Ensure database connection is closed when the worker stops
	defer func() {
		log.Println("Closing MySQL connection...")
		if err := db.CloseDB(); err != nil {
			log.Printf("Failed to close database: %v", err)
		} else {
			log.Println("MySQL connection closed")
		}
	}()

	// Start Kafka consumer in a separate goroutine
	go kafka.StartConsumer()

	// Capture shutdown signals (CTRL+C, Docker Stop, etc.)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal: %v, shutting down Kafka Consumer...", sig)
}
