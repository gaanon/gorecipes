package main

import (
	"context" // Import context
	"log"
	"net/http" // Import net/http
	"os"
	"os/signal"
	"syscall"
	"time" // Import time

	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/router"
)

func main() {
	log.Println("Starting Go Recipes API...") // Changed from fmt.Println for consistency

	// Define database path
	// When running in Docker, this path should match the volume mount point.
	// The Dockerfile sets WORKDIR /app
	// The docker-compose.yml mounts the volume to /app/data/badgerdb
	dbPath := "/app/data/badgerdb"

	// Initialize Database
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	// defer database.CloseDB() // Will call this explicitly on shutdown

	// Initialize Gin router using the setup function
	appRouter := router.SetupRouter()

	// Start the server
	port := os.Getenv("PORT") // Use environment variable for port
	if port == "" {
		port = "8080" // Default port
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0" // Default to all interfaces
	}

	srv := &http.Server{
		Addr:    host + ":" + port,
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
	database.CloseDB() // Close DB after server has shut down
}
