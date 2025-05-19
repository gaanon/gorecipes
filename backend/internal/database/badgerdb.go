package database

import (
	// Added for MealPlanEntry iteration
	"encoding/json"
	"fmt"
	"gorecipes/backend/internal/models"
	"log"
	"os"
	"path/filepath"
	"time" // Added for MealPlanEntry date handling

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

// MealPlanEntryKeyPrefix is the prefix for all meal plan entry keys.
const MealPlanEntryKeyPrefix = "mealplanentry:"

// mealPlanEntryKey generates a byte slice key for a meal plan entry.
func mealPlanEntryKey(entryID string) []byte {
	return []byte(fmt.Sprintf("%s%s", MealPlanEntryKeyPrefix, entryID))
}

// SaveMealPlanEntry saves a meal plan entry to the database.
func SaveMealPlanEntry(entry *models.MealPlanEntry) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	if entry.ID == "" {
		return fmt.Errorf("meal plan entry ID cannot be empty")
	}
	if entry.RecipeID == "" {
		return fmt.Errorf("meal plan entry RecipeID cannot be empty")
	}
	if entry.Date.IsZero() {
		return fmt.Errorf("meal plan entry Date cannot be zero")
	}

	key := mealPlanEntryKey(entry.ID)
	value, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal meal plan entry ID %s: %w", entry.ID, err)
	}

	err = DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
	if err != nil {
		return fmt.Errorf("failed to save meal plan entry ID %s: %w", entry.ID, err)
	}
	log.Printf("Meal plan entry saved successfully: ID=%s, RecipeID=%s, Date=%s", entry.ID, entry.RecipeID, entry.Date.Format("2006-01-02"))
	return nil
}

// GetMealPlanEntriesByDateRange retrieves all meal plan entries within a given date range (inclusive).
// Dates are expected to be YYYY-MM-DD (time part is ignored for comparison).
func GetMealPlanEntriesByDateRange(startDate, endDate time.Time) ([]models.MealPlanEntry, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var entries []models.MealPlanEntry
	prefix := []byte(MealPlanEntryKeyPrefix)

	// Normalize start and end dates to midnight UTC for proper comparison
	startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	err := DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var entry models.MealPlanEntry
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &entry)
			})
			if err != nil {
				log.Printf("Error unmarshalling meal plan entry with key %s: %v. Skipping.", string(item.Key()), err)
				continue // Skip this entry if unmarshalling fails
			}

			// Normalize entry date to midnight UTC for comparison
			entryDateNormalized := time.Date(entry.Date.Year(), entry.Date.Month(), entry.Date.Day(), 0, 0, 0, 0, time.UTC)

			// Check if the entry's date is within the range (inclusive)
			if (entryDateNormalized.Equal(startOfDay) || entryDateNormalized.After(startOfDay)) &&
				(entryDateNormalized.Equal(endOfDay) || entryDateNormalized.Before(endOfDay)) {
				entries = append(entries, entry)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error retrieving meal plan entries: %w", err)
	}
	return entries, nil
}

// DeleteMealPlanEntry deletes a meal plan entry from the database by its ID.
func DeleteMealPlanEntry(entryID string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	if entryID == "" {
		return fmt.Errorf("meal plan entry ID cannot be empty for deletion")
	}

	key := mealPlanEntryKey(entryID)
	err := DB.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err == badger.ErrKeyNotFound {
			log.Printf("Attempted to delete non-existent meal plan entry with ID: %s", entryID)
			return nil // Not an error if key doesn't exist, it's already "deleted"
		}
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to delete meal plan entry ID %s: %w", entryID, err)
	}
	log.Printf("Meal plan entry deleted (or did not exist): ID=%s", entryID)
	return nil
}

// Comment-related functions removed.
