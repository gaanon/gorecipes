package handlers

import (
	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	// Normalize date to UTC midnight
	normalizedDate := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)

	// Optional: Validate if RecipeID exists
	recipeExists, err := database.RecipeExists(req.RecipeID)
	if err != nil {
		log.Printf("[MealPlanner] Create: Error checking recipe existence for ID %s: %v", req.RecipeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validating recipe ID"})
		return
	}
	if !recipeExists {
		log.Printf("[MealPlanner] Create: Recipe with ID %s not found.", req.RecipeID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found."})
		return
	}

	entry := models.MealPlanEntry{
		ID:        uuid.New().String(),
		Date:      normalizedDate,
		RecipeID:  req.RecipeID,
		CreatedAt: time.Now().UTC(),
	}

	if err := database.SaveMealPlanEntry(&entry); err != nil {
		log.Printf("[MealPlanner] Create: Error saving meal plan entry: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save meal plan entry."})
		return
	}

	log.Printf("[MealPlanner] Create: Successfully created meal plan entry ID %s for Recipe %s on %s", entry.ID, entry.RecipeID, entry.Date.Format(dateLayout))
	c.JSON(http.StatusCreated, entry)
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
