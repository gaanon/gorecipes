package handlers

import (
	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const dateLayout = "2006-01-02" // For parsing YYYY-MM-DD

// CreateMealPlanEntryHandler handles POST /api/v1/mealplanner/entries
func CreateMealPlanEntryHandler(c *gin.Context) {
	var req struct {
		Date     string `json:"date" binding:"required"`
		RecipeID string `json:"recipe_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[MealPlanner] Create: Bad request format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	parsedDate, err := time.Parse(dateLayout, req.Date)
	if err != nil {
		log.Printf("[MealPlanner] Create: Invalid date format for %s: %v", req.Date, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Please use YYYY-MM-DD."})
		return
	}

	// Skip recipe existence check to allow custom recipe names
	// Custom recipes (text-only entries) can be added to meal plans without existing in the recipe database
	log.Printf("[MealPlanner] Create: Adding recipe/custom entry '%s' to meal plan for date %s", req.RecipeID, req.Date)

	// Prepare the entry. ID and CreatedAt will be set by the database.CreateMealPlanEntry function.
	// The Date field in entry will also be normalized to UTC midnight by CreateMealPlanEntry.
	entryData := models.MealPlanEntry{
		Date:     parsedDate, // Pass the parsed date; normalization happens in DB func
		RecipeID: req.RecipeID,
	}

	createdEntry, err := database.CreateMealPlanEntry(&entryData)
	if err != nil {
		log.Printf("[MealPlanner] Create: Error saving meal plan entry with PostgreSQL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save meal plan entry."})
		return
	}

	log.Printf("[MealPlanner] Create: Successfully created meal plan entry ID %s for Recipe %s on %s using PostgreSQL", createdEntry.ID, createdEntry.RecipeID, createdEntry.Date.Format(dateLayout))
	c.JSON(http.StatusCreated, createdEntry)
}

// ListMealPlanEntriesHandler handles GET /api/v1/mealplanner/entries
func ListMealPlanEntriesHandler(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		log.Printf("[MealPlanner] List: Missing start_date or end_date query parameter.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date query parameters are required."})
		return
	}

	startDate, err := time.Parse(dateLayout, startDateStr)
	if err != nil {
		log.Printf("[MealPlanner] List: Invalid start_date format %s: %v", startDateStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Please use YYYY-MM-DD."})
		return
	}
	endDate, err := time.Parse(dateLayout, endDateStr)
	if err != nil {
		log.Printf("[MealPlanner] List: Invalid end_date format %s: %v", endDateStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Please use YYYY-MM-DD."})
		return
	}

	// Normalize dates for reliable range
	normalizedStartDate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	normalizedEndDate := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	if normalizedEndDate.Before(normalizedStartDate) {
		log.Printf("[MealPlanner] List: end_date %s is before start_date %s.", endDateStr, startDateStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date cannot be before start_date."})
		return
	}

	entries, err := database.GetMealPlanEntriesByDateRange(normalizedStartDate, normalizedEndDate)
	if err != nil {
		log.Printf("[MealPlanner] List: Error fetching meal plan entries: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meal plan entries."})
		return
	}

	if entries == nil { // Ensure an empty array is returned instead of null if no entries found
		entries = []models.MealPlanEntry{}
	}

	log.Printf("[MealPlanner] List: Returning %d entries for date range %s to %s", len(entries), startDateStr, endDateStr)
	c.JSON(http.StatusOK, entries)
}

// DeleteMealPlanEntryHandler handles DELETE /api/v1/mealplanner/entries/:entry_id
func DeleteMealPlanEntryHandler(c *gin.Context) {
	entryID := c.Param("entry_id")
	if entryID == "" {
		log.Printf("[MealPlanner] Delete: entry_id parameter is missing.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "entry_id path parameter is required."})
		return
	}

	// Optional: Check if entry exists before attempting delete if you want to return 404 specifically
	// For now, DeleteMealPlanEntry in database layer handles non-existent key gracefully (logs it).

	if err := database.DeleteMealPlanEntry(entryID); err != nil {
		log.Printf("[MealPlanner] Delete: Error deleting meal plan entry ID %s: %v", entryID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete meal plan entry."})
		return
	}

	log.Printf("[MealPlanner] Delete: Successfully deleted (or confirmed non-existent) meal plan entry ID %s", entryID)
	c.Status(http.StatusNoContent)
}
