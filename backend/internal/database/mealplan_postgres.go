package database

import (
	"fmt"
	"gorecipes/backend/internal/models"
	"log"
	"time"
	"context"      // Added for QueryContext
	"database/sql" // Added for sql.NullString

	"github.com/google/uuid"
)

// CreateMealPlanEntry adds a new meal plan entry to the PostgreSQL database.
func CreateMealPlanEntry(entry *models.MealPlanEntry) (*models.MealPlanEntry, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Prepare entry data
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	entry.CreatedAt = time.Now().UTC()
	// Ensure the Date field is just the date part, without time, for DATE column compatibility
	entry.Date = time.Date(entry.Date.Year(), entry.Date.Month(), entry.Date.Day(), 0, 0, 0, 0, time.UTC)

	query := `INSERT INTO meal_plan_entries (id, recipe_id, date, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := DB.Exec(query, entry.ID, entry.RecipeID, entry.Date, entry.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert meal plan entry ID %s: %w", entry.ID, err)
	}

	log.Printf("Meal plan entry created successfully: ID=%s, RecipeID=%s, Date=%s", entry.ID, entry.RecipeID, entry.Date.Format("2006-01-02"))
	return entry, nil
}

// GetMealPlanEntriesByDateRange retrieves all meal plan entries within a given date range (inclusive).
func GetMealPlanEntriesByDateRange(startDate, endDate time.Time) ([]models.MealPlanEntry, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Normalize dates to ensure the entire day is covered for comparison with DATE type in SQL.
	// For DATE type, '2023-01-01' is equivalent to '2023-01-01 00:00:00'.
	// So, we can use the date part directly.
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	query := `SELECT id, recipe_id, date, created_at
		FROM meal_plan_entries
		WHERE date >= $1 AND date <= $2
		ORDER BY date ASC, created_at ASC`

	rows, err := DB.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("error querying meal plan entries by date range: %w", err)
	}
	defer rows.Close()

	var entries []models.MealPlanEntry
	for rows.Next() {
		var entry models.MealPlanEntry
		if err := rows.Scan(&entry.ID, &entry.RecipeID, &entry.Date, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning meal plan entry: %w", err)
		}
		// Ensure the Date from DB (which is DATE type) is correctly parsed into time.Time (usually midnight UTC)
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating meal plan entries: %w", err)
	}

	return entries, nil
}

// DeleteMealPlanEntry removes a meal plan entry from the PostgreSQL database by its ID.
func DeleteMealPlanEntry(entryID string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	if entryID == "" {
		return fmt.Errorf("meal plan entry ID cannot be empty for deletion")
	}

	query := `DELETE FROM meal_plan_entries WHERE id = $1`

	res, err := DB.Exec(query, entryID)
	if err != nil {
		return fmt.Errorf("failed to delete meal plan entry ID %s: %w", entryID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// This error is less critical if the delete query itself didn't fail,
		// but good to log for diagnostics.
		log.Printf("Warning: could not get rows affected after deleting meal plan entry ID %s: %v", entryID, err)
	}

	if rowsAffected == 0 {
		// If no rows were affected, the entry didn't exist.
		// This might not be an error condition depending on desired idempotency.
		log.Printf("Meal plan entry with ID %s not found for deletion, or already deleted.", entryID)
		// Optionally, return a specific 'not found' error here if strictness is required.
		// return fmt.Errorf("meal plan entry with ID %s not found", entryID) 
	}

	log.Printf("Meal plan entry deleted successfully (or did not exist): ID=%s", entryID)
	return nil
}

// GetAllMealPlanEntries fetches all meal_plan_entries from the database.
func GetAllMealPlanEntries() ([]models.MealPlanEntry, error) {
	rows, err := DB.QueryContext(context.Background(), `SELECT id, recipe_id, date, notes, created_at FROM meal_plan_entries ORDER BY date ASC, created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("error querying meal_plan_entries: %w", err)
	}
	defer rows.Close()

	var mealPlanEntries []models.MealPlanEntry
	for rows.Next() {
		var mpe models.MealPlanEntry
		var notes sql.NullString // Use sql.NullString for nullable text fields
		if err := rows.Scan(&mpe.ID, &mpe.RecipeID, &mpe.Date, &notes, &mpe.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning meal_plan_entry: %w", err)
		}
		if notes.Valid {
			mpe.Notes = notes.String
		} else {
			mpe.Notes = "" // Represent NULL notes as an empty string
		}
		mealPlanEntries = append(mealPlanEntries, mpe)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating meal_plan_entry rows: %w", err)
	}
	return mealPlanEntries, nil
}
