package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"job-scheduler/internal/api"
	"job-scheduler/internal/config"
	"job-scheduler/internal/database"
	"job-scheduler/internal/scheduler"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	log.Println("Starting Job Scheduler Daemon...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Check interval: %d seconds", cfg.CheckInterval)

	// Connect to database
	db, err := database.New(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start API server in background
	apiServer := api.New(db, 8080)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping scheduler...")
		cancel()
	}()

	// Create and start the scheduler
	sched := scheduler.New(db, cfg.CheckInterval)
	sched.Run(ctx)

	log.Println("Job Scheduler Daemon stopped")
}
