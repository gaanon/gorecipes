import { writable, derived, get } from 'svelte/store';
import type { Recipe } from '$lib/types'; // Assuming Recipe type is defined
import type { MealPlanEntry } from '$lib/models/mealPlanEntry'; // We'll create this frontend model

// Helper to get the start of a week (Sunday or Monday based on locale, default Monday for now)
export function getStartOfWeek(date: Date, startDay: 'monday' | 'sunday' = 'monday'): Date {
	const d = new Date(date);
	const day = d.getDay(); // Sunday - 0, Monday - 1, ..., Saturday - 6
	const diff = startDay === 'monday' ? (day === 0 ? -6 : 1 - day) : (0 - day); // if Sunday is start, diff is 0-day
	d.setDate(d.getDate() + diff);
	d.setHours(0, 0, 0, 0); // Normalize to start of day
	return d;
}

// Helper to format date to YYYY-MM-DD
export function formatDateToYYYYMMDD(date: Date): string {
	const year = date.getFullYear();
	const month = (date.getMonth() + 1).toString().padStart(2, '0');
	const day = date.getDate().toString().padStart(2, '0');
	return `${year}-${month}-${day}`;
}

// --- Store Definitions ---

// Represents the start date of the week currently being viewed in the main planner panel.
// Defaults to the start of the current week.
export const currentPlannerWeekStartDate = writable<Date>(getStartOfWeek(new Date()));

// Stores the meal plan entries fetched from the backend for the currentPlannerWeekStartDate.
// Keyed by YYYY-MM-DD date string, value is an array of MealPlanEntry objects for that date.
export const mealPlanEntriesMap = writable<Map<string, MealPlanEntry[]>>(new Map());

// Stores the selected date in the planner panel for displaying daily meals.
// Defaults to today.
export const selectedPlannerDate = writable<Date>(new Date());
selectedPlannerDate.subscribe(date => {
    // When selectedPlannerDate changes, ensure currentPlannerWeekStartDate is updated
    // if the selected date falls outside the currently viewed week.
    const currentWeekStart = get(currentPlannerWeekStartDate);
    const endOfCurrentWeek = new Date(currentWeekStart);
    endOfCurrentWeek.setDate(currentWeekStart.getDate() + 6);

    if (date < currentWeekStart || date > endOfCurrentWeek) {
        currentPlannerWeekStartDate.set(getStartOfWeek(date));
    }
});


// --- Derived Store for Convenience ---

// Derived store to get a flat list of all MealPlanEntry objects for the current week
export const plannedMealsForCurrentWeekArray = derived(
	[currentPlannerWeekStartDate, mealPlanEntriesMap],
	([$currentPlannerWeekStartDate, $mealPlanEntriesMap]) => {
		const entries: MealPlanEntry[] = [];
		const weekStart = getStartOfWeek(new Date($currentPlannerWeekStartDate)); // Ensure it's a fresh Date object
		for (let i = 0; i < 7; i++) {
			const day = new Date(weekStart);
			day.setDate(weekStart.getDate() + i);
			const dateStr = formatDateToYYYYMMDD(day);
			const dailyEntries = $mealPlanEntriesMap.get(dateStr);
			if (dailyEntries) {
				entries.push(...dailyEntries);
			}
		}
		return entries;
	}
);


// --- API Interaction Functions ---

const API_BASE_URL = '/api/v1/mealplanner';

/**
 * Fetches meal plan entries for the week starting with the given startDate.
 * Updates the mealPlanEntriesMap store.
 */
export async function fetchMealPlanForWeek(startDate: Date): Promise<void> {
	const weekStartDate = getStartOfWeek(new Date(startDate)); // Ensure it's a fresh Date object
	const weekEndDate = new Date(weekStartDate);
	weekEndDate.setDate(weekStartDate.getDate() + 6); // Get 7 days

	const startDateStr = formatDateToYYYYMMDD(weekStartDate);
	const endDateStr = formatDateToYYYYMMDD(weekEndDate);

	try {
		const response = await fetch(`${API_BASE_URL}/entries?start_date=${startDateStr}&end_date=${endDateStr}`);
		if (!response.ok) {
			throw new Error(`Failed to fetch meal plan: ${response.statusText}`);
		}
		const entries: MealPlanEntry[] = await response.json();
		
		const newMap = new Map<string, MealPlanEntry[]>();
        // Initialize map for all days in the fetched week to ensure reactivity
        for (let i = 0; i < 7; i++) {
            const day = new Date(weekStartDate);
            day.setDate(weekStartDate.getDate() + i);
            newMap.set(formatDateToYYYYMMDD(day), []);
        }

		entries.forEach(entry => {
			// Backend date should be YYYY-MM-DD UTC. Frontend Date object will parse it as local.
            // For consistency, ensure we use the date string as key.
            const entryDate = new Date(entry.date); // entry.date is likely a string like "2023-10-26T00:00:00Z"
			const dateStrKey = formatDateToYYYYMMDD(entryDate);
			const existing = newMap.get(dateStrKey) || [];
			existing.push(entry);
			newMap.set(dateStrKey, existing);
		});
		mealPlanEntriesMap.set(newMap);
	} catch (error) {
		console.error("Error fetching meal plan:", error);
		mealPlanEntriesMap.set(new Map()); // Clear on error or handle appropriately
	}
}

/**
 * Adds a recipe to the meal plan for a specific date.
 * Refetches the meal plan for the relevant week on success.
 */
export async function addRecipeToMealPlan(date: Date, recipeId: string): Promise<MealPlanEntry | null> {
	const dateStr = formatDateToYYYYMMDD(date);
	try {
		const response = await fetch(`${API_BASE_URL}/entries`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ date: dateStr, recipe_id: recipeId }),
		});
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ error: 'Failed to add recipe to meal plan.' }));
			throw new Error(errorData.error || `Failed to add recipe: ${response.statusText}`);
		}
		const newEntry: MealPlanEntry = await response.json();
		// Refetch the week's data to update the UI
		await fetchMealPlanForWeek(get(currentPlannerWeekStartDate)); // Or specifically the week of 'date'
		return newEntry;
	} catch (error) {
		console.error("Error adding recipe to meal plan:", error);
		throw error; // Re-throw to be caught by UI
	}
}

/**
 * Removes a meal plan entry (a specific recipe from a specific day).
 * Refetches the meal plan for the relevant week on success.
 */
export async function removeRecipeFromMealPlan(entryId: string): Promise<void> {
	try {
		const response = await fetch(`${API_BASE_URL}/entries/${entryId}`, {
			method: 'DELETE',
		});
		if (!response.ok && response.status !== 204) { // 204 No Content is a success
			const errorData = await response.json().catch(() => ({ error: 'Failed to remove recipe from meal plan.' }));
			throw new Error(errorData.error || `Failed to remove recipe: ${response.statusText}`);
		}
		// Refetch the week's data to update the UI
		await fetchMealPlanForWeek(get(currentPlannerWeekStartDate));
	} catch (error) {
		console.error("Error removing recipe from meal plan:", error);
		throw error; // Re-throw to be caught by UI
	}
}

/**
 * Adds a custom recipe name directly to the meal plan for a specific date.
 * This doesn't create a recipe in the database, just adds the text as a meal plan entry.
 */
export async function addCustomRecipeToMealPlan(date: Date, recipeName: string): Promise<MealPlanEntry | null> {
	const dateStr = formatDateToYYYYMMDD(date);
	try {
		const response = await fetch(`${API_BASE_URL}/entries`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ date: dateStr, recipe_id: recipeName }), // Use recipe name as ID for custom recipes
		});
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ error: 'Failed to add custom recipe to meal plan.' }));
			throw new Error(errorData.error || `Failed to add custom recipe: ${response.statusText}`);
		}
		const newEntry: MealPlanEntry = await response.json();
		// Refetch the week's data to update the UI
		await fetchMealPlanForWeek(get(currentPlannerWeekStartDate));
		return newEntry;
	} catch (error) {
		console.error("Error adding custom recipe to meal plan:", error);
		throw error; // Re-throw to be caught by UI
	}
}

