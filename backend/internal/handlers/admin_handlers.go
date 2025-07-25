package handlers

import (
	"encoding/json"
	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/models"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid" // May not be needed if IDs come from import
)

// ImportRecipesResponse defines the structure for the import API response.
type ImportRecipesResponse struct {
	TotalRecipesInFile        int    `json:"total_recipes_in_file"`
	SuccessfullyImportedCount int    `json:"successfully_imported_count"`
	SkippedDuplicateCount     int    `json:"skipped_duplicate_count"`
	SkippedMalformedCount     int    `json:"skipped_malformed_count"`
	ErrorMessage              string `json:"error_message,omitempty"` // For file-level errors
}

// ImportRecipes handles the POST /api/v1/admin/import endpoint.
func ImportRecipes(c *gin.Context) {
	response := ImportRecipesResponse{}

	file, header, err := c.Request.FormFile("recipes_file")
	if err != nil {
		log.Printf("[ImportRecipes] Error getting form file: %v", err)
		response.ErrorMessage = "Recipes file not provided or error in form data."
		c.JSON(http.StatusBadRequest, response)
		return
	}
	defer file.Close()

	log.Printf("[ImportRecipes] Received file: %s, Size: %d", header.Filename, header.Size)

	// Basic file type check (optional, but good for early exit)
	// if !strings.HasSuffix(strings.ToLower(header.Filename), ".json") {
	// 	response.ErrorMessage = "Invalid file type. Please upload a .json file."
	// 	c.JSON(http.StatusBadRequest, response)
	// 	return
	// }

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[ImportRecipes] Error reading file content: %v", err)
		response.ErrorMessage = "Error reading file content."
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	var recipesFromFile []models.Recipe
	if err := json.Unmarshal(fileBytes, &recipesFromFile); err != nil {
		log.Printf("[ImportRecipes] Error unmarshalling JSON: %v", err)
		response.ErrorMessage = "Invalid JSON file format. Failed to unmarshal recipes."
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response.TotalRecipesInFile = len(recipesFromFile)
	log.Printf("[ImportRecipes] Parsed %d recipes from file.", response.TotalRecipesInFile)

	for _, recipeFromFile := range recipesFromFile {
		// Basic Validation (as per plan)
		if recipeFromFile.ID == "" {
			log.Printf("[ImportRecipes] Skipped: Recipe ID is empty. Name: '%s'", recipeFromFile.Name)
			response.SkippedMalformedCount++
			continue
		}
		if recipeFromFile.Name == "" {
			log.Printf("[ImportRecipes] Skipped: Recipe Name is empty. ID: '%s'", recipeFromFile.ID)
			response.SkippedMalformedCount++
			continue
		}
		if recipeFromFile.Method == "" {
			log.Printf("[ImportRecipes] Skipped: Recipe Method is empty. ID: '%s'", recipeFromFile.ID)
			response.SkippedMalformedCount++
			continue
		}
		if recipeFromFile.CreatedAt.IsZero() || recipeFromFile.UpdatedAt.IsZero() {
			log.Printf("[ImportRecipes] Skipped: Recipe CreatedAt/UpdatedAt is zero. ID: '%s'", recipeFromFile.ID)
			response.SkippedMalformedCount++
			continue
		}
		// Ingredients can be an empty slice, so no check needed unless specific validation is added.

		// Check for Duplicates using PostgreSQL version
		exists, err := database.RecipeExistsByID(recipeFromFile.ID)
		if err != nil {
			log.Printf("[ImportRecipes] Error checking recipe existence for ID %s with PostgreSQL: %v. Skipping.", recipeFromFile.ID, err)
			response.SkippedMalformedCount++ // Treat DB error during check as a reason to skip
			continue
		}
		if exists {
			log.Printf("[ImportRecipes] Skipped duplicate: Recipe ID %s already exists (checked with PostgreSQL).", recipeFromFile.ID)
			response.SkippedDuplicateCount++
			continue
		}

		// Prepare for Save - FilterableIngredientNames is deprecated and handled by CreateRecipe
		recipeToSave := models.Recipe{
			ID:            recipeFromFile.ID,
			Name:          recipeFromFile.Name,
			Ingredients:   recipeFromFile.Ingredients, // CreateRecipe will process these
			Method:        recipeFromFile.Method,
			PhotoFilename: "", // Ignored as per plan, CreateRecipe will handle default if necessary
			CreatedAt:     recipeFromFile.CreatedAt, // Preserve timestamps from import
			UpdatedAt:     recipeFromFile.UpdatedAt, // Preserve timestamps from import
		}
		// If recipeFromFile.Ingredients is nil, ensure it's an empty slice for CreateRecipe
		if recipeToSave.Ingredients == nil {
			recipeToSave.Ingredients = []string{}
		}

		// Save to Database using PostgreSQL CreateRecipe
		// CreateRecipe handles ingredient processing and linking.
		// It also sets CreatedAt/UpdatedAt if they are zero, but here we provide them.
		createdRecipe, err := database.CreateRecipe(&recipeToSave) 
		if err != nil {
			log.Printf("[ImportRecipes] Error saving recipe ID %s with PostgreSQL CreateRecipe: %v. Skipping.", recipeToSave.ID, err)
			response.SkippedMalformedCount++
			continue
		}
		response.SuccessfullyImportedCount++
		log.Printf("[ImportRecipes] Successfully imported recipe ID %s, Name: %s using PostgreSQL", createdRecipe.ID, createdRecipe.Name)
	}

	log.Printf("[ImportRecipes] Import process complete. Results: %+v", response)
	c.JSON(http.StatusOK, response)
}

// Note: The ExportData handler might also be moved here or to a more general admin_handlers.go
// For now, assuming it's in recipes.go as per ADMIN_EXPORT_FEATURE_PLAN.md
