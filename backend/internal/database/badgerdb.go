package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

var (
	// DB is the global BadgerDB instance
	DB *badger.DB
)

// InitDB initializes the BadgerDB database.
// It creates the database directory if it doesn't exist.
func InitDB(dbPath string) error {
	// Ensure the database path directory exists
	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		log.Printf("Database directory %s does not exist, creating it...", dbDir)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err // Some other error stating the directory
	}

	log.Printf("Initializing database at: %s", dbPath)
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable Badger's own logger to reduce noise, or configure it as needed

	var err error
	DB, err = badger.Open(opts)
	if err != nil {
		return err
	}
	log.Println("Database initialized successfully.")
	return nil
}

// CloseDB closes the BadgerDB database.
func CloseDB() {
	if DB != nil {
		log.Println("Closing database...")
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database closed successfully.")
		}
	}
}
