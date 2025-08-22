package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil" // For ReadFile, or use os.ReadFile in Go 1.16+
	"log"
	"os" // Required for os.ReadFile if using Go 1.16+ and os.Getenv
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var (
	// DB is the global PostgreSQL database connection pool.
	DB *sql.DB
)

// InitPostgreSQLDB initializes the PostgreSQL database connection.
// It expects a PostgreSQL connection string (e.g., "postgres://user:password@host:port/dbname?sslmode=disable").
func InitPostgreSQLDB(connectionString string) error {
	if connectionString == "" {
		return fmt.Errorf("database connection string cannot be empty")
	}

	log.Println("Initializing PostgreSQL database connection...")
	var err error
	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Pinging PostgreSQL database...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Add timeout for ping
	defer cancel()
	err = DB.PingContext(ctx) // Use PingContext for timeout
	if err != nil {
		DB.Close() // Close the connection if ping fails
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("PostgreSQL database connected successfully.")

	// Run migrations
	migrationsPath := "./internal/database/migrations/"
	if err := executeSQLFile(DB, migrationsPath+"001_initial_schema.sql", "initial schema"); err != nil {
		log.Printf("Could not apply initial schema migration: %v", err)
		// Depending on the desired behavior, you might want to return this error
	}
	if err := executeSQLFile(DB, migrationsPath+"20250613162217_create_comments_table.sql", "comments table migration"); err != nil {
		log.Printf("Could not apply comments table migration: %v", err)
		// Depending on the desired behavior, you might want to return this error
	}

	return nil
}

// ClosePostgreSQLDB closes the PostgreSQL database connection.
func ClosePostgreSQLDB() {
	if DB != nil {
		log.Println("Closing PostgreSQL database connection...")
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("PostgreSQL database connection closed successfully.")
		}
	}
}

// executeSQLFile reads and executes SQL statements from a given file path.
func executeSQLFile(db *sql.DB, filePath string, stepName string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("%s file (%s) not found. Skipping this step.", stepName, filePath)
		return nil // Not an error if file doesn't exist, just skip
	}

	log.Printf("Executing %s from %s...", stepName, filePath)

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Warning: Failed to read %s file %s: %v. Continuing.", stepName, filePath, err)
		return err // Return error to stop further processing if critical, or nil to allow continuation
	}

	// Begin a transaction for this file execution
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Warning: Failed to begin transaction for %s: %v. Continuing.", stepName, err)
		return err
	}

	// Execute the entire file content as a single block of SQL.
	// The pq driver supports multi-statement execution.
	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback() // Rollback transaction on error
		log.Printf("Warning: Failed to execute %s: error executing SQL block: %v.", stepName, err)
		// log.Printf("Failed SQL block during %s: %s", stepName, string(content)) // Optional: log the failing SQL block
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Warning: Failed to commit %s transaction: %v.", stepName, err)
		return err
	}

	log.Printf("%s executed successfully from %s.", stepName, filePath)
	return nil
}
