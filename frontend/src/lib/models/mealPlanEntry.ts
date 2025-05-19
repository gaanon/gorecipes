/**
 * Represents a single recipe planned for a specific date on the frontend.
 * This should mirror the backend's `models.MealPlanEntry` struct.
 */
export interface MealPlanEntry {
	id: string;         // Unique ID for this meal plan entry
	date: string;       // The specific date as an ISO string (e.g., "2023-10-26T00:00:00Z")
	recipe_id: string;  // ID of the planned recipe
	created_at: string; // Timestamp of when the entry was created (ISO string)
	// recipe?: Recipe; // Optional: To hold hydrated recipe details if fetched
}

// You might also want to include the Recipe type here or import it if it's used often with MealPlanEntry
// For now, keeping it separate as per current store logic.
// import type { Recipe } from '$lib/types';