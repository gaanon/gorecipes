package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil" // For ReadFile, or use os.ReadFile in Go 1.16+
	"log"
	"os" // Required for os.ReadFile if using Go 1.16+ and os.Getenv
	"path/filepath"
	"strconv" // For parsing boolean from env var
	"strings" // For parsing boolean from env var (strings.ToLower)
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

// SeedData applies the schema and then seeds data if the respective files exist.
func SeedData(db *sql.DB) error {
	schemaFilePath := filepath.Join("internal", "database", "schema.sql")   // Relative to backend directory
	seedFilePath := filepath.Join("internal", "database", "seed_data.sql") // Relative to backend directory

	// Step 1: Apply schema
	err := executeSQLFile(db, schemaFilePath, "schema")
	if err != nil {
		log.Printf("Error applying schema, seeding will be skipped: %v", err)
		// Decide if this should be a fatal error for the app or just a warning
		return nil // For now, let the app continue even if schema fails, but log it prominently
	}

	// Step 2: Conditionally Seed data
	enableSeedingStr := os.Getenv("GORECIPES_ENABLE_SEED_DATA")
	enableSeeding, _ := strconv.ParseBool(strings.ToLower(enableSeedingStr)) // Defaults to false if parsing fails or var is empty/not "true"/"1"

	if enableSeeding {
		log.Println("GORECIPES_ENABLE_SEED_DATA is true. Proceeding with data seeding.")
		err = executeSQLFile(db, seedFilePath, "seed data")
		if err != nil {
			log.Printf("Error seeding data: %v", err)
			// Decide if this should be a fatal error for the app or just a warning
			return nil // For now, let the app continue even if seeding fails
		}
		log.Println("Database initialization (schema and seeding) complete.")
	} else {
		log.Printf("GORECIPES_ENABLE_SEED_DATA is not 'true' (value: '%s'). Skipping data seeding.", enableSeedingStr)
		log.Println("Database initialization (schema application only) complete.")
	}

	return nil
}





