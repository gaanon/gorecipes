package database

import (
	"encoding/json" // Added for SaveRecipe
	"fmt"
	"gorecipes/backend/internal/models" // Added for SaveRecipe
	"log"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

var (
	// DB is the global BadgerDB instance
	DB *badger.DB
)

// RecipeKeyPrefix is the prefix for all recipe keys in the database.
const RecipeKeyPrefix = "recipe:"

// recipeKey generates a byte slice key for a recipe.
func recipeKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", RecipeKeyPrefix, id))
}

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

// RecipeExists checks if a recipe with the given ID exists in the database.
func RecipeExists(id string) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database not initialized")
	}
	if id == "" {
		return false, fmt.Errorf("recipe ID cannot be empty for existence check")
	}

	key := recipeKey(id)
	var exists bool

	err := DB.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			exists = true // Key found
			return nil
		}
		if err == badger.ErrKeyNotFound {
			exists = false // Key not found
			return nil     // Not an actual error for this check
		}
		return err // Other error
	})

	if err != nil {
		return false, fmt.Errorf("error checking recipe existence for ID %s: %w", id, err)
	}
	return exists, nil
}

// SaveRecipe saves a recipe to the database.
// It will overwrite an existing recipe if the ID matches.
func SaveRecipe(recipe *models.Recipe) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	if recipe.ID == "" {
		return fmt.Errorf("recipe ID cannot be empty")
	}

	key := recipeKey(recipe.ID)
	value, err := json.Marshal(recipe)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe ID %s: %w", recipe.ID, err)
	}

	err = DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
	if err != nil {
		return fmt.Errorf("failed to save recipe ID %s: %w", recipe.ID, err)
	}
	log.Printf("Recipe saved successfully: ID=%s, Name=%s", recipe.ID, recipe.Name)
	return nil
}

// Comment-related functions removed.
