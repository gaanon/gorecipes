package main

import (
	"context" // Import context
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/router"
)

func main() {
	log.Println("Starting Go Recipes API...") // Changed from fmt.Println for consistency

	// Database Configuration
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("WARNING: DATABASE_URL environment variable not set. Using default development URL.")
		// This is an example default for local development.
		// Ensure your PostgreSQL server is running and accessible with these credentials.
		dbURL = "postgres://postgres:password@localhost:5432/gorecipes_dev?sslmode=disable"
		log.Printf("Using default DATABASE_URL: %s (Ensure this is correctly configured for your environment)", dbURL)
	}

	// Initialize Database
	if err := database.InitPostgreSQLDB(dbURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Seed the database with sample data

	// defer database.CloseDB() // Will call this explicitly on shutdown

	// Initialize Gin router using the setup function
	appRouter := router.SetupRouter()

	// Start the server
	port := os.Getenv("PORT") // Use environment variable for port
	if port == "" {
		port = "8080" // Default port
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: appRouter,
	}

	// Start server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the requests it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Release resources if main completes before timeout

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
	database.ClosePostgreSQLDB() // Close DB after server has shut down
}
