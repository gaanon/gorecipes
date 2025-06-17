package handlers

import (
	"encoding/json"
	"fmt" // Added for Pexels integration
	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/models"
	"io"
	"log"
	"math" // Added for pagination (Ceil)
	"mime/multipart"
	"net/http"
	"net/url" // Added for Pexels integration (URL encoding)
	"os"
	"path/filepath"
	"regexp"  // Added for ingredient parsing
	"strconv" // Added for pagination
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const uploadsDir = "uploads/images/" // Relative to backend directory
const defaultPageLimit = 25
const pexelsAPIURL = "https://api.pexels.com/v1/search"
const placeholderImage = "placeholder.jpg"

// Pexels API Response Structures
type PexelsPhotoSource struct {
	Original  string `json:"original"`
	Large     string `json:"large"`
	Large2x   string `json:"large2x"`
	Medium    string `json:"medium"`
	Small     string `json:"small"`
	Portrait  string `json:"portrait"`
	Landscape string `json:"landscape"`
	Tiny      string `json:"tiny"`
}

type PexelsPhoto struct {
	ID           int               `json:"id"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	URL          string            `json:"url"` // Pexels page URL for the photo
	Photographer string            `json:"photographer"`
	Src          PexelsPhotoSource `json:"src"`
	Alt          string            `json:"alt"`
}

type PexelsSearchResponse struct {
	Page         int           `json:"page"`
	PerPage      int           `json:"per_page"`
	Photos       []PexelsPhoto `json:"photos"`
	TotalResults int           `json:"total_results"`
	NextPage     string        `json:"next_page"`
}

// fetchAndSaveImageFromPexels tries to fetch an image from Pexels based on the query,
// download it, and save it. It returns the saved filename or an error.
func fetchAndSaveImageFromPexels(query string, recipeID string, apiKey string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("Pexels API key is not configured")
	}

	// 1. Construct Pexels API Request
	reqURL, err := url.Parse(pexelsAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse Pexels API URL: %w", err)
	}
	qParams := reqURL.Query()
	qParams.Set("query", query)
	qParams.Set("per_page", "1")
	reqURL.RawQuery = qParams.Encode()

	// 2. Execute HTTP GET Request
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Pexels API request: %w", err)
	}
	req.Header.Set("Authorization", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute Pexels API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Pexels API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 3. Parse JSON Response
	var pexelsResp PexelsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&pexelsResp); err != nil {
		return "", fmt.Errorf("failed to decode Pexels API response: %w", err)
	}

	// 4. Extract Image URL
	if len(pexelsResp.Photos) == 0 || pexelsResp.Photos[0].Src.Large == "" {
		return "", fmt.Errorf("no suitable image found on Pexels for query: %s", query)
	}
	imageURL := pexelsResp.Photos[0].Src.Large // Using 'large' size

	// 5. Download Image
	imgResp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image from Pexels URL %s: %w", imageURL, err)
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image from Pexels, status: %d", imgResp.StatusCode)
	}

	// 6. Determine File Extension
	contentType := imgResp.Header.Get("Content-Type")
	var extension string
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	default:
		// Try to infer from URL, or default
		ext := filepath.Ext(imageURL)
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			extension = ext
		} else {
			extension = ".jpg" // Default to .jpg if unsure
			log.Printf("Warning: Unknown content type '%s' for Pexels image, defaulting to .jpg", contentType)
		}
	}

	// 7. Generate Unique Filename
	savedFilename := recipeID + "_pexels" + extension
	dstPath := filepath.Join(uploadsDir, savedFilename)

	// 8. Save Image
	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for Pexels image: %w", err)
	}
	outFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file for Pexels image: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, imgResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save Pexels image to file: %w", err)
	}

	log.Printf("Successfully fetched and saved image from Pexels for recipe %s as %s", recipeID, savedFilename)
	return savedFilename, nil
}

// Helper function to save uploaded file
func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

var (
	// commonUnits is a list of common units of measurement to remove.
	// This list can be expanded.
	commonUnits = []string{
		"g", "kg", "mg", "oz", "lb", "lbs",
		"ml", "l", "cl", "dl",
		"tsp", "tbsp", "fl oz", "cup", "cups", "pt", "qt", "gal",
		"pinch", "dash", "clove", "cloves", "head", "heads",
		"slice", "slices", "piece", "pieces",
		// Fractional words
		"half", "quarter",
	}

	// commonDescriptors is a list of common adjectives or preparation states to remove.
	// This list can be expanded.
	commonDescriptors = []string{
		"fresh", "dried", "frozen", "canned", "cooked", "uncooked", "raw",
		"chopped", "diced", "sliced", "minced", "grated", "crushed", "peeled", "seeded",
		"large", "medium", "small",
		"ripe", "unripe",
		"optional", "to taste", "for garnish",
		"plain", "all-purpose", "self-raising", "whole", "ground", "granulated", "powdered",
		"boneless", "skinless",
		"finely", "coarsely", "roughly",
		"hot", "cold", "warm", "chilled",
		"sweet", "unsweetened", "salted", "unsalted",
	}

	// commonStopWords are words that are generally not useful for filtering.
	// This list can be expanded.
	commonStopWords = []string{
		"a", "an", "the", "of", "and", "or", "with", "without", "in", "on", "at", "for", "to", "from",
		"some", "any", "about", "into", "over", "under",
	}
)

// normalizeAndCleanToken prepares a token by lowercasing and trimming.
func normalizeAndCleanToken(token string) string {
	return strings.ToLower(strings.TrimSpace(token))
}

// removeSubstrings removes all occurrences of given substrings from a string.
func removeSubstrings(s string, substrings []string) string {
	for _, sub := range substrings {
		// Use word boundaries for units and descriptors to avoid partial matches in words
		// e.g., removing "g" from "garlic"
		// Regex for word boundary: \b
		// However, simple Replace might be okay for a first pass if lists are curated.
		// For more precision, regex would be better: regexp.MustCompile(`\b` + regexp.QuoteMeta(sub) + `\b`)
		// For now, let's do a simpler replace, but be mindful of this.
		// To be safer, add spaces around the substrings to be removed if they are standalone words.
		s = strings.ReplaceAll(s, " "+sub+" ", " ") // for words in the middle
		s = strings.ReplaceAll(s, sub+" ", " ")     // for words at the beginning (after number removal)
		s = strings.ReplaceAll(s, " "+sub, " ")     // for words at the end
	}
	return s
}

// extractFilterableNames attempts to extract core ingredient names from a full ingredient string.
// Example: "180g plain flour, finely chopped" -> ["plain flour", "flour"]
func extractFilterableNames(fullIngredient string) []string {
	if strings.TrimSpace(fullIngredient) == "" {
		return nil
	}

	// 1. Lowercase
	processed := strings.ToLower(fullIngredient)

	// 2. Remove quantities (numbers and common fractions like 1/2, 1 1/2)
	// Regex to remove numbers, fractions, and mixed numbers.
	// This regex handles: "1", "1.5", "1/2", "1 1/2", "1-2"
	reNum := regexp.MustCompile(`\d+(\s*[-/]\s*\d+)?(\.\d+)?(\s+\d+/\d+)?`)
	processed = reNum.ReplaceAllString(processed, "")

	// 3. Remove common units
	processed = removeSubstrings(processed, commonUnits)

	// 4. Remove common descriptors
	processed = removeSubstrings(processed, commonDescriptors)

	// 5. Split into words, remove stop words, and collect remaining.
	// Also, consider multi-word ingredient names that might remain.
	potentialNames := make(map[string]bool)

	// Add the processed string as a whole, if it's meaningful
	trimmedProcessed := strings.TrimSpace(processed)
	// Replace multiple spaces with a single space
	trimmedProcessed = regexp.MustCompile(`\s+`).ReplaceAllString(trimmedProcessed, " ")
	if len(trimmedProcessed) > 2 { // Arbitrary length to avoid very short/meaningless "names"
		potentialNames[trimmedProcessed] = true
	}

	// Split into individual words and add them if not stop words
	words := strings.Fields(trimmedProcessed)
	for _, word := range words {
		cleanedWord := normalizeAndCleanToken(word)
		isStopWord := false
		for _, stop := range commonStopWords {
			if cleanedWord == stop {
				isStopWord = true
				break
			}
		}
		if !isStopWord && len(cleanedWord) > 2 { // Arbitrary length
			potentialNames[cleanedWord] = true
		}
	}

	if len(potentialNames) == 0 && len(words) > 0 {
		// Fallback: if after all filtering nothing is left, but there were words,
		// maybe the original string (just lowercased and trimmed) is the best we can do.
		// This handles cases where an ingredient is just "salt" or "pepper".
		originalTrimmed := normalizeAndCleanToken(fullIngredient)
		// Remove numbers from this fallback too
		originalTrimmed = reNum.ReplaceAllString(originalTrimmed, "")
		originalTrimmed = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(originalTrimmed), " ")
		if len(originalTrimmed) > 1 {
			potentialNames[originalTrimmed] = true
		}
	}

	// Convert map to slice
	var result []string
	for name := range potentialNames {
		result = append(result, name)
	}
	return result
}

// @Summary Create a new recipe
// @Description Create a new recipe with name, method, ingredients, and an optional photo.
// @Tags recipes
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Name of the recipe"
// @Param method formData string true "Cooking method"
// @Param ingredients formData string false "Newline-separated list of ingredients"
// @Param photo formData file false "Recipe photo"
// @Success 201 {object} models.Recipe "Recipe created successfully"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /recipes [post]
func CreateRecipe(c *gin.Context) {
	var recipe models.Recipe
	// Generate ID in handler for use in photo filename generation before DB call.
	// database.CreateRecipe will use this ID if provided.
	recipe.ID = uuid.New().String()

	recipe.Name = c.PostForm("name")
	recipe.Method = c.PostForm("method")
	ingredientsStr := c.PostForm("ingredients")

	if strings.TrimSpace(recipe.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe name cannot be empty"})
		return
	}
	if strings.TrimSpace(recipe.Method) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe method cannot be empty"})
		return
	}

	// Process ingredients from comma-separated string to []string
	recipe.Ingredients = []string{}
	if ingredientsStr != "" {
		rawIngredients := strings.Split(ingredientsStr, "\n")
		uniqueIngredients := make(map[string]bool)
		for _, ing := range rawIngredients {
			trimmedIng := strings.TrimSpace(ing)
			if trimmedIng != "" && !uniqueIngredients[trimmedIng] {
				recipe.Ingredients = append(recipe.Ingredients, trimmedIng)
				uniqueIngredients[trimmedIng] = true
			}
		}
	}

	// Handle photo upload / Pexels integration
	file, errFile := c.FormFile("photo")
	if errFile == nil {
		// User uploaded a photo
		photoFilename := recipe.ID + filepath.Ext(file.Filename)
		// Ensure uploadsDir exists
		if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
			if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
				log.Printf("Error creating uploads directory %s: %v", uploadsDir, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
				return
			}
		}
		dst := filepath.Join(uploadsDir, photoFilename)
		if err := saveUploadedFile(file, dst); err != nil {
			log.Printf("Error saving uploaded file for new recipe %s: %v", recipe.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
			return
		}
		recipe.PhotoFilename = photoFilename
		log.Printf("Photo saved for new recipe %s: %s", recipe.ID, photoFilename)
	} else if errFile != http.ErrMissingFile {
		// Some other error with file upload
		log.Printf("Error retrieving photo from form for new recipe %s: %v", recipe.ID, errFile)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing photo upload"})
		return
	} else { // http.ErrMissingFile: No file uploaded by user, try Pexels or placeholder
		pexelsAPIKey := os.Getenv("PEXELS_API_KEY")
		if pexelsAPIKey != "" && recipe.Name != "" {
			log.Printf("No photo uploaded for recipe %s. Attempting to fetch from Pexels...", recipe.ID)
			fetchedFilename, errPexels := fetchAndSaveImageFromPexels(recipe.Name, recipe.ID, pexelsAPIKey)
			if errPexels == nil && fetchedFilename != "" {
				recipe.PhotoFilename = fetchedFilename
				log.Printf("Successfully used Pexels image %s for recipe %s", fetchedFilename, recipe.ID)
			} else {
				log.Printf("Failed to fetch image from Pexels for recipe %s (query: %s): %v. Using placeholder.", recipe.ID, recipe.Name, errPexels)
				recipe.PhotoFilename = placeholderImage // Ensure placeholder is set if Pexels fails
			}
		} else {
			if pexelsAPIKey == "" {
				log.Printf("Pexels API key not configured. Using placeholder image for recipe %s.", recipe.ID)
			} else { // recipe.Name is empty
				log.Printf("Recipe name is empty, cannot fetch from Pexels. Using placeholder for recipe %s.", recipe.ID)
			}
			recipe.PhotoFilename = placeholderImage
		}
	}

	// If PhotoFilename is still empty after all attempts (e.g. Pexels disabled and no upload), set placeholder as a safeguard.
	if recipe.PhotoFilename == "" {
		recipe.PhotoFilename = placeholderImage
	}

	// Timestamps (CreatedAt, UpdatedAt) will be set by the database.CreateRecipe function.

	// Save recipe to PostgreSQL database
	createdRecipe, errDb := database.CreateRecipe(&recipe)
	if errDb != nil {
		log.Printf("Error saving recipe to database (ID attempted: %s): %v", recipe.ID, errDb)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save recipe"})
		return
	}

	log.Printf("Recipe created successfully: ID=%s, Name=%s", createdRecipe.ID, createdRecipe.Name)
	c.JSON(http.StatusCreated, createdRecipe)
}

// PaginatedRecipesResponse defines the structure for paginated recipe results.
type PaginatedRecipesResponse struct {
	Recipes      []models.Recipe `json:"recipes"`
	TotalRecipes int             `json:"total_recipes"`
	Page         int             `json:"page"`
	Limit        int             `json:"limit"`
	TotalPages   int             `json:"total_pages"`
}

// @Summary List all recipes
// @Description Get a paginated list of all recipes, with optional search and ingredient filtering.
// @Tags recipes
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(25)
// @Param search query string false "Search term for recipe name or method"
// @Param tags query string false "Comma-separated list of ingredient tags to filter by"
// @Success 200 {object} PaginatedRecipesResponse "Successfully retrieved recipes"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /recipes [get]
func ListRecipes(c *gin.Context) {
	// Parse query parameters for pagination
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultPageLimit))

	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil || page < 1 {
		page = 1
	}

	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil || limit <= 0 {
		limit = defaultPageLimit
	}
	// Optional: Add a max limit if desired, e.g., if limit > 100 { limit = 100 }

	// Parse query parameters for filtering and searching
	searchTerm := strings.TrimSpace(c.Query("search"))
	tagsQuery := c.Query("tags")
	var ingredientFilters []string
	if tagsQuery != "" {
		rawTags := strings.Split(tagsQuery, ",")
		for _, t := range rawTags {
			trimmedTag := strings.ToLower(strings.TrimSpace(t))
			if trimmedTag != "" {
				ingredientFilters = append(ingredientFilters, trimmedTag)
			}
		}
	}

	log.Printf("[ListRecipes] Query Params: page=%d, limit=%d, search='%s', tags=%v", page, limit, searchTerm, ingredientFilters)

	// Fetch recipes from PostgreSQL database
	recipes, totalCount, err := database.GetAllRecipes(searchTerm, ingredientFilters, page, limit)
	if err != nil {
		log.Printf("Error retrieving recipes from database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipes"})
		return
	}

	if recipes == nil {
		recipes = []models.Recipe{} // Ensure we return an empty array, not null
	}

	totalPages := 0
	if totalCount > 0 && limit > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	response := PaginatedRecipesResponse{
		Recipes:      recipes,
		TotalRecipes: totalCount,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// containsAnyTag checks if the recipeIngredients list contains at least one tag from the filterTags list.
// It now accepts recipeID for logging purposes.
func containsAnyTag(recipeID string, recipeIngredients []string, filterTags []string) bool {
	if len(filterTags) == 0 {
		return true // Should not happen if called from ListRecipes where len(filterTags) > 0
	}
	normalizedRecipeIngredients := make(map[string]bool)
	// recipeIngredients now refers to recipe.FilterableIngredientNames, which are already somewhat processed.
	// Normalizing them again here is fine and ensures consistency.
	for _, name := range recipeIngredients { // Changed 'ing' to 'name' for clarity
		normalizedRecipeIngredients[strings.ToLower(strings.TrimSpace(name))] = true
	}
	// Updated log to reflect that we are dealing with FilterableIngredientNames
	log.Printf("[containsAnyTag] Recipe ID: %s, Normalized Filterable Names: %v || Filter Tags: %v", recipeID, normalizedRecipeIngredients, filterTags)

	for _, filterTag := range filterTags {
		// filterTags are already normalized in ListRecipes.
		// No need to re-normalize filterTag here if ListRecipes guarantees it.
		// However, for safety, let's keep it:
		normalizedFilterTag := strings.ToLower(strings.TrimSpace(filterTag))
		if _, ok := normalizedRecipeIngredients[normalizedFilterTag]; ok {
			log.Printf("[containsAnyTag] Match FOUND for Recipe ID: %s. Filterable name '%s' matches filter tag '%s'", recipeID, normalizedFilterTag, filterTag)
			return true
		}
	}
	log.Printf("[containsAnyTag] No match found for Recipe ID: %s after checking all filter tags against recipe's filterable names.", recipeID)
	return false
}

// @Summary Get a recipe by ID
// @Description Get a single recipe by its unique ID.
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200 {object} models.Recipe "Successfully retrieved recipe"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Recipe not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /recipes/{id} [get]
func GetRecipe(c *gin.Context) {
	recipeID := c.Param("id")

	if recipeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe ID cannot be empty"})
		return
	}

	recipe, err := database.GetRecipeByID(recipeID)
	if err != nil {
		// Check if the error is due to the recipe not being found.
		// database.GetRecipeByID is expected to return an error that can be identified as 'not found'.
		// For example, if it wraps sql.ErrNoRows, we could check for that.
		// Assuming database.GetRecipeByID returns a specific error type or message for not found.
		// For now, let's assume a generic error check and log it.
		// A more robust way would be to define a custom error in the database package, e.g., database.ErrNotFound.
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("Recipe with ID %s not found: %v", recipeID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		} else {
			log.Printf("Error retrieving recipe %s from database: %v", recipeID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipe"})
		}
		return
	}

	c.JSON(http.StatusOK, recipe)
}

// @Summary Update an existing recipe
// @Description Update an existing recipe by its ID with new name, method, ingredients, and optional photo.
// @Tags recipes
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Recipe ID"
// @Param name formData string true "Name of the recipe"
// @Param method formData string true "Cooking method"
// @Param ingredients formData string false "Newline-separated list of ingredients"
// @Param photo formData file false "New recipe photo"
// @Success 200 {object} models.Recipe "Recipe updated successfully"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Recipe not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /recipes/{id} [put]
func UpdateRecipe(c *gin.Context) {
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe ID cannot be empty"})
		return
	}

	// Fetch existing recipe to get current photo filename and other details
	existingRecipe, err := database.GetRecipeByID(recipeID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("Recipe with ID %s not found for update: %v", recipeID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		} else {
			log.Printf("Error retrieving recipe %s for update: %v", recipeID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipe for update"})
		}
		return
	}

	// Create a recipe model to hold updated values
	recipeToUpdate := *existingRecipe // Start with existing values

	// Update fields from form data
	recipeToUpdate.Name = c.PostForm("name")
	recipeToUpdate.Method = c.PostForm("method")
	ingredientsStr := c.PostForm("ingredients")

	if strings.TrimSpace(recipeToUpdate.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe name cannot be empty"})
		return
	}
	if strings.TrimSpace(recipeToUpdate.Method) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe method cannot be empty"})
		return
	}

	// Process ingredients
	var updatedIngredients []string
	if ingredientsStr != "" {
		rawIngredients := strings.Split(ingredientsStr, "\n")
		for _, ing := range rawIngredients {
			trimmedIng := strings.TrimSpace(ing)
			if trimmedIng != "" {
				updatedIngredients = append(updatedIngredients, trimmedIng)
			}
		}
	}
	recipeToUpdate.Ingredients = updatedIngredients // Can be empty if ingredientsStr was empty or all spaces

	// Handle photo update
	oldPhotoFilename := existingRecipe.PhotoFilename
	newPhotoUploaded := false

	file, errUpload := c.FormFile("photo")
	if errUpload == nil {
		// New photo uploaded
		newPhotoFilename := recipeID + "_updated_" + uuid.New().String() + filepath.Ext(file.Filename)
		// Ensure uploadsDir exists
		if _, errStat := os.Stat(uploadsDir); os.IsNotExist(errStat) {
			if errMkdir := os.MkdirAll(uploadsDir, os.ModePerm); errMkdir != nil {
				log.Printf("Error creating uploads directory %s during update: %v", uploadsDir, errMkdir)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
				return
			}
		}
		dst := filepath.Join(uploadsDir, newPhotoFilename)
		if errSave := saveUploadedFile(file, dst); errSave != nil {
			log.Printf("Error saving updated photo file for recipe %s: %v", recipeID, errSave)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save updated photo"})
			return // Or decide to proceed without photo update
		} else {
			recipeToUpdate.PhotoFilename = newPhotoFilename
			newPhotoUploaded = true
			log.Printf("New photo saved for recipe %s: %s", recipeID, newPhotoFilename)
		}
	} else if errUpload != http.ErrMissingFile {
		// Error other than 'no file'
		log.Printf("Error retrieving photo from form during update for recipe %s: %v", recipeID, errUpload)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing photo upload for update"})
		return
	}
	// If no new photo was uploaded (errUpload == http.ErrMissingFile), recipeToUpdate.PhotoFilename remains existingRecipe.PhotoFilename

	// If a new photo was successfully uploaded and there was an old one (not placeholder, not the same as new),
	// delete the old photo from the filesystem.
	if newPhotoUploaded && oldPhotoFilename != "" && oldPhotoFilename != placeholderImage && oldPhotoFilename != recipeToUpdate.PhotoFilename {
		oldPhotoPath := filepath.Join(uploadsDir, oldPhotoFilename)
		if errRemove := os.Remove(oldPhotoPath); errRemove != nil {
			log.Printf("Error deleting old photo %s for recipe %s: %v", oldPhotoPath, recipeID, errRemove)
			// Non-fatal, just log it.
		} else {
			log.Printf("Old photo deleted for recipe %s: %s", recipeID, oldPhotoPath)
		}
	}

	// If after all, PhotoFilename is empty (e.g. was placeholder and no new upload), ensure it's set to placeholder.
	if recipeToUpdate.PhotoFilename == "" {
		recipeToUpdate.PhotoFilename = placeholderImage
	}

	// Timestamps (UpdatedAt) will be handled by database.UpdateRecipe

	updatedRecipe, errDb := database.UpdateRecipe(&recipeToUpdate)
	if errDb != nil {
		// database.UpdateRecipe might also return a 'not found' error if the ID doesn't exist at the time of update.
		if strings.Contains(strings.ToLower(errDb.Error()), "not found") || strings.Contains(errDb.Error(), "no rows in result set") {
			log.Printf("Recipe with ID %s not found during database update: %v", recipeID, errDb)
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found for update"})
		} else {
			log.Printf("Error updating recipe %s in database: %v", recipeID, errDb)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
		}
		return
	}

	log.Printf("Recipe updated successfully: ID=%s, Name=%s", updatedRecipe.ID, updatedRecipe.Name)
	c.JSON(http.StatusOK, updatedRecipe)
}

// @Summary Delete a recipe
// @Description Delete a recipe by its unique ID.
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /recipes/{id} [delete]
func DeleteRecipe(c *gin.Context) {
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe ID cannot be empty"})
		return
	}

	// Step 1: Fetch the recipe to get its photo filename before deleting from DB.
	recipeToDelete, err := database.GetRecipeByID(recipeID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("Recipe with ID %s not found (already deleted or never existed): %v", recipeID, err)
			c.Status(http.StatusNoContent) // Recipe is gone, so operation is effectively successful.
		} else {
			log.Printf("Error retrieving recipe %s for deletion: %v", recipeID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipe before deletion"})
		}
		return
	}

	// Step 2: Delete the recipe from the database.
	errDbDelete := database.DeleteRecipe(recipeID)
	if errDbDelete != nil {
		// If GetRecipeByID succeeded, a "not found" here would be unusual but handle defensively.
		if strings.Contains(strings.ToLower(errDbDelete.Error()), "not found") || strings.Contains(errDbDelete.Error(), "no rows in result set") {
			log.Printf("Recipe with ID %s was not found during DB deletion (possibly deleted concurrently): %v", recipeID, errDbDelete)
			// Proceed to photo deletion if recipeToDelete has photo info, then return 204.
		} else {
			log.Printf("Error deleting recipe %s from database: %v", recipeID, errDbDelete)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe from database"})
			return
		}
	}

	// Step 3: If recipe was fetched and had a photo (and it's not the placeholder), delete the photo file.
	if recipeToDelete != nil && recipeToDelete.PhotoFilename != "" && recipeToDelete.PhotoFilename != placeholderImage {
		photoPath := filepath.Join(uploadsDir, recipeToDelete.PhotoFilename)
		// Ensure uploadsDir exists before trying to remove a file from it (though unlikely to be an issue here)
		if _, errStat := os.Stat(uploadsDir); os.IsNotExist(errStat) {
			log.Printf("Uploads directory %s does not exist, cannot delete photo %s", uploadsDir, photoPath)
		} else {
			if errRemove := os.Remove(photoPath); errRemove != nil {
				// Log error but don't fail the overall operation if DB deletion was successful.
				log.Printf("Error deleting photo file %s for deleted recipe %s: %v", photoPath, recipeID, errRemove)
			} else {
				log.Printf("Photo file deleted for recipe %s: %s", recipeID, photoPath)
			}
		}
	}

	log.Printf("Recipe deleted successfully: %s", recipeID)
	c.Status(http.StatusNoContent)
}

// GetIngredientsAutocomplete handles fetching ingredient suggestions.
// GET /api/v1/ingredients?q=<query>
func GetIngredientsAutocomplete(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	var matchingIngredients []string

	if query == "" {
		c.JSON(http.StatusOK, matchingIngredients)
		return
	}

	// TODO: Implement PostgreSQL specific logic to query the 'ingredients' table
	// For now, returning empty to avoid build errors.
	log.Println("[GetIngredientsAutocomplete] BadgerDB logic removed. Needs PostgreSQL implementation.")
	var err error // Keep err declared for the check below
	// err = fmt.Errorf("PostgreSQL implementation pending") // Example of setting an error

	if err != nil {
		log.Printf("Error searching ingredients: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search ingredients"})
		return
	}

	if matchingIngredients == nil {
		matchingIngredients = []string{}
	}
	c.JSON(http.StatusOK, matchingIngredients)
}

// ExportData handles exporting all recipe and related data.
// POST /api/v1/admin/export
func ExportData(c *gin.Context) {
	var exportedData models.ExportedData
	var err error

	exportedData.Recipes, err = database.GetAllRecipesForExport()
	if err != nil {
		log.Printf("Error fetching recipes for export: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipes for export"})
		return
	}

	exportedData.Ingredients, err = database.GetAllIngredients()
	if err != nil {
		log.Printf("Error fetching ingredients for export: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ingredients for export"})
		return
	}

	exportedData.RecipeIngredients, err = database.GetAllRecipeIngredients()
	if err != nil {
		log.Printf("Error fetching recipe ingredients for export: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe ingredients for export"})
		return
	}

	log.Printf("Successfully fetched data for export. Recipes: %d, Ingredients: %d, RecipeIngredients: %d",
		len(exportedData.Recipes), len(exportedData.Ingredients), len(exportedData.RecipeIngredients))

	c.Header("Content-Disposition", "attachment; filename=gorecipes_export.json")
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, exportedData)
}

// ImportData handles importing data from a JSON file.
// POST /api/v1/admin/import
func ImportData(c *gin.Context) {
	file, err := c.FormFile("importFile")
	if err != nil {
		log.Printf("Error getting import file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Import file is required"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		log.Printf("Error opening import file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open import file"})
		return
	}
	defer openedFile.Close()

	byteValue, err := io.ReadAll(openedFile)
	if err != nil {
		log.Printf("Error reading import file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read import file"})
		return
	}

	var dataToImport models.ExportedData
	if err := json.Unmarshal(byteValue, &dataToImport); err != nil {
		log.Printf("Error unmarshalling import file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in import file"})
		return
	}

	log.Printf("Successfully parsed import file. Recipes: %d, Ingredients: %d, RecipeIngredients: %d",
		len(dataToImport.Recipes), len(dataToImport.Ingredients), len(dataToImport.RecipeIngredients))

	importedRecipes, importedIngredients, importedLinks, err := database.ImportRecipeDataBundle(dataToImport)
	if err != nil {
		log.Printf("Error importing data to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to import data: %v", err)})
		return
	}

	log.Printf("Successfully imported data. Recipes: %d, Ingredients: %d, RecipeIngredients Links: %d",
		importedRecipes, importedIngredients, importedLinks)

	c.JSON(http.StatusOK, gin.H{
		"message":               "Data imported successfully.",
		"imported_recipes":      importedRecipes,
		"imported_ingredients":  importedIngredients,
		"imported_recipe_links": importedLinks,
	})
}
