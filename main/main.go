package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"magebase/apis/analytics/app"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create new application
	app := app.NewApp(port)
	app.SetupRoutes()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	// Start the application
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Failed to start analytics service: %v", err)
	}
}
