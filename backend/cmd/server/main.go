package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/router" // Added router import
	// "github.com/gin-gonic/gin" // Gin is now used within the router package primarily
)

func main() {
	fmt.Println("Starting Go Recipes API...")

	// Define database path
	dbPath := "./badger_data" // Relative to the backend directory

	// Initialize Database
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB() // Ensure DB is closed when main function exits

	// Initialize Gin router using the setup function
	appRouter := router.SetupRouter()

	// Start the server
	port := "8080" // We can make this configurable later
	log.Printf("Server listening on port %s", port)
	if err := appRouter.Run(":" + port); err != nil { // Use appRouter here
		log.Fatalf("Failed to run server: %v", err)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}
