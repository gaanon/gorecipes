package models

import "time"

// MealPlanEntry represents a single recipe planned for a specific date.
// Each assignment of a recipe to a day is a unique entry.
type MealPlanEntry struct {
	ID        string    `json:"id"`         // Unique ID for this meal plan entry (e.g., UUID)
	Date      time.Time `json:"date"`       // The specific date (YYYY-MM-DD), time part normalized to UTC midnight
	RecipeID  string    `json:"recipe_id"`  // ID of the planned recipe
	CreatedAt time.Time `json:"created_at"` // Timestamp of when the entry was created
	// UserID    string    `json:"user_id"`    // Future: For multi-user support
}
