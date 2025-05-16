package handlers

import (
	"archive/zip" // Added for image ZIP export
	"bytes"
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
	"sort"    // Added for sorting recipes
	"strconv" // Added for pagination
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
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

// CreateRecipe handles the creation of a new recipe.
// POST /api/v1/recipes - expects multipart/form-data
func CreateRecipe(c *gin.Context) {
	var recipe models.Recipe
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

	// Initialize Ingredients and FilterableIngredientNames
	recipe.Ingredients = []string{}
	recipe.FilterableIngredientNames = []string{}

	uniqueFilterableNames := make(map[string]bool)

	if ingredientsStr != "" {
		rawIngredients := strings.Split(ingredientsStr, ",")
		for _, ing := range rawIngredients {
			trimmedIng := strings.TrimSpace(ing)
			if trimmedIng != "" {
				recipe.Ingredients = append(recipe.Ingredients, trimmedIng)
				// Extract and add filterable names
				filterable := extractFilterableNames(trimmedIng)
				for _, fname := range filterable {
					if fname != "" && !uniqueFilterableNames[fname] {
						recipe.FilterableIngredientNames = append(recipe.FilterableIngredientNames, fname)
						uniqueFilterableNames[fname] = true
					}
				}
			}
		}
	}
	// No need to check for nil recipe.Ingredients as it's initialized

	file, errFile := c.FormFile("photo")
	if errFile == nil {
		photoFilename := recipe.ID + filepath.Ext(file.Filename)
		dst := filepath.Join(uploadsDir, photoFilename)
		if err := saveUploadedFile(file, dst); err != nil {
			log.Printf("Error saving uploaded file for new recipe %s: %v", recipe.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
			return
		}
		recipe.PhotoFilename = photoFilename
		log.Printf("Photo saved for new recipe %s: %s", recipe.ID, photoFilename)
	} else if errFile != http.ErrMissingFile {
		log.Printf("Error retrieving photo from form for new recipe %s: %v", recipe.ID, errFile)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing photo upload"})
		return
	} else if errFile == http.ErrMissingFile {
		// No file uploaded by user, try to fetch from Pexels
		pexelsAPIKey := os.Getenv("PEXELS_API_KEY")
		if pexelsAPIKey != "" && recipe.Name != "" {
			log.Printf("No photo uploaded for recipe %s. Attempting to fetch from Pexels...", recipe.ID)
			fetchedFilename, errPexels := fetchAndSaveImageFromPexels(recipe.Name, recipe.ID, pexelsAPIKey)
			if errPexels == nil && fetchedFilename != "" {
				recipe.PhotoFilename = fetchedFilename
				log.Printf("Successfully used Pexels image %s for recipe %s", fetchedFilename, recipe.ID)
			} else {
				log.Printf("Failed to fetch image from Pexels for recipe %s (query: %s): %v. Using placeholder.", recipe.ID, recipe.Name, errPexels)
				recipe.PhotoFilename = placeholderImage
			}
		} else {
			if pexelsAPIKey == "" {
				log.Printf("Pexels API key not configured. Using placeholder image for recipe %s.", recipe.ID)
			} else {
				log.Printf("Recipe name is empty, cannot fetch from Pexels. Using placeholder for recipe %s.", recipe.ID)
			}
			recipe.PhotoFilename = placeholderImage
		}
	}

	now := time.Now().UTC()
	recipe.CreatedAt = now
	recipe.UpdatedAt = now

	recipeJSON, errMarshal := json.Marshal(recipe)
	if errMarshal != nil {
		log.Printf("Error marshalling recipe %s to JSON: %v", recipe.ID, errMarshal)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process recipe data"})
		return
	}

	errDb := database.DB.Update(func(txn *badger.Txn) error {
		recipeKey := []byte("recipe:" + recipe.ID)
		if err := txn.Set(recipeKey, recipeJSON); err != nil {
			return err
		}
		for _, ingredientName := range recipe.Ingredients {
			if strings.TrimSpace(ingredientName) == "" {
				continue
			}
			ingredientKey := []byte("ingredient:" + strings.ToLower(strings.TrimSpace(ingredientName)))
			if err := txn.Set(ingredientKey, []byte("1")); err != nil {
				log.Printf("Error saving ingredient key %s for recipe %s: %v", ingredientKey, recipe.ID, err)
			}
		}
		return nil
	})

	if errDb != nil {
		log.Printf("Error saving recipe %s to database: %v", recipe.ID, errDb)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save recipe"})
		return
	}

	log.Printf("Recipe created successfully: %s", recipe.ID)
	c.JSON(http.StatusCreated, recipe)
}

// PaginatedRecipesResponse defines the structure for paginated recipe results.
type PaginatedRecipesResponse struct {
	Recipes      []models.Recipe `json:"recipes"`
	TotalRecipes int             `json:"total_recipes"`
	Page         int             `json:"page"`
	Limit        int             `json:"limit"`
	TotalPages   int             `json:"total_pages"`
}

// ListRecipes handles listing recipes with pagination and filtering.
// GET /api/v1/recipes
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
	// Optional: Add a max limit if desired
	// if limit > 100 { limit = 100 }

	var filteredRecipes []models.Recipe // Stores recipes after tag filtering, before pagination
	tagsQuery := c.Query("tags")
	var filterTags []string
	if tagsQuery != "" {
		rawTags := strings.Split(tagsQuery, ",")
		for _, t := range rawTags {
			trimmedTag := strings.ToLower(strings.TrimSpace(t))
			if trimmedTag != "" {
				filterTags = append(filterTags, trimmedTag)
			}
		}
	}
	log.Printf("[ListRecipes] Received tagsQuery: '%s', Parsed filterTags: %v", tagsQuery, filterTags) // DEBUG LOGGING

	err := database.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10 // PrefetchSize can be tuned
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("recipe:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var recipe models.Recipe
			errValue := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &recipe)
			})
			if errValue != nil {
				log.Printf("Error unmarshalling recipe %s: %v", item.Key(), errValue)
				// Potentially skip this recipe or return error
				return errValue // This will stop the iteration
			}

			if len(filterTags) > 0 {
				// Log details of the recipe being checked
				log.Printf("[ListRecipes] Checking Recipe - ID: %s, Name: '%s', FilterableIngredientNames: %v. Against filterTags: %v", recipe.ID, recipe.Name, recipe.FilterableIngredientNames, filterTags)
				if containsAnyTag(recipe.ID, recipe.FilterableIngredientNames, filterTags) {
					log.Printf("[ListRecipes] Match FOUND for Recipe ID: %s, Name: '%s'. Adding to results.", recipe.ID, recipe.Name)
					filteredRecipes = append(filteredRecipes, recipe)
				} else {
					log.Printf("[ListRecipes] No match for Recipe ID: %s, Name: '%s'.", recipe.ID, recipe.Name)
				}
			} else {
				filteredRecipes = append(filteredRecipes, recipe) // Add if no filters are active
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error retrieving recipes from database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipes"})
		return
	}

	// Sort recipes alphabetically by name before pagination
	sort.Slice(filteredRecipes, func(i, j int) bool {
		return strings.ToLower(filteredRecipes[i].Name) < strings.ToLower(filteredRecipes[j].Name)
	})

	// Perform in-memory pagination on the filteredRecipes
	totalFilteredRecipes := len(filteredRecipes)
	if totalFilteredRecipes == 0 {
		c.JSON(http.StatusOK, PaginatedRecipesResponse{
			Recipes:      []models.Recipe{},
			TotalRecipes: 0,
			Page:         page,
			Limit:        limit,
			TotalPages:   0,
		})
		return
	}

	totalPages := int(math.Ceil(float64(totalFilteredRecipes) / float64(limit)))
	if page > totalPages {
		// If requested page is out of bounds after filtering,
		// it's debatable whether to return an error or an empty list for that page.
		// For "load more" style, returning empty might be fine.
		// For strict pagination, an error or last page might be better.
		// Let's return empty for now.
		page = totalPages // Or handle as an error, e.g. c.JSON(http.StatusNotFound, ...)
	}

	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex >= totalFilteredRecipes {
		// This case means the requested page is beyond the available data
		c.JSON(http.StatusOK, PaginatedRecipesResponse{
			Recipes:      []models.Recipe{}, // Empty slice for this page
			TotalRecipes: totalFilteredRecipes,
			Page:         page,
			Limit:        limit,
			TotalPages:   totalPages,
		})
		return
	}

	if endIndex > totalFilteredRecipes {
		endIndex = totalFilteredRecipes
	}

	paginatedSlice := filteredRecipes[startIndex:endIndex]
	if paginatedSlice == nil { // Ensure it's an empty slice not nil
		paginatedSlice = []models.Recipe{}
	}

	response := PaginatedRecipesResponse{
		Recipes:      paginatedSlice,
		TotalRecipes: totalFilteredRecipes,
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

// GetRecipe handles fetching a single recipe by ID.
// GET /api/v1/recipes/:id
func GetRecipe(c *gin.Context) {
	recipeID := c.Param("id")
	var recipe models.Recipe

	err := database.DB.View(func(txn *badger.Txn) error {
		key := []byte("recipe:" + recipeID)
		item, errGet := txn.Get(key)
		if errGet != nil {
			return errGet
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &recipe)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			log.Printf("Recipe with ID %s not found: %v", recipeID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
			return
		}
		log.Printf("Error retrieving recipe %s from database: %v", recipeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipe"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// UpdateRecipe handles updating an existing recipe.
// PUT /api/v1/recipes/:id - expects multipart/form-data
func UpdateRecipe(c *gin.Context) {
	recipeID := c.Param("id")

	name := c.PostForm("name")
	method := c.PostForm("method")
	ingredientsStr := c.PostForm("ingredients")

	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe name cannot be empty"})
		return
	}
	if strings.TrimSpace(method) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe method cannot be empty"})
		return
	}

	var updatedIngredients []string
	if ingredientsStr != "" {
		rawIngredients := strings.Split(ingredientsStr, ",")
		for _, ing := range rawIngredients {
			trimmedIng := strings.TrimSpace(ing)
			if trimmedIng != "" {
				updatedIngredients = append(updatedIngredients, trimmedIng)
			}
		}
	}
	if updatedIngredients == nil {
		updatedIngredients = []string{}
	}

	var existingRecipe models.Recipe
	var newPhotoFilename string // Stores name if a new photo is uploaded

	err := database.DB.Update(func(txn *badger.Txn) error {
		key := []byte("recipe:" + recipeID)
		item, errGet := txn.Get(key)
		if errGet != nil {
			return errGet
		}
		errGet = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &existingRecipe)
		})
		if errGet != nil {
			return errGet
		}

		oldPhotoFilename := existingRecipe.PhotoFilename

		existingRecipe.Name = name
		// Process ingredients and filterable names
		existingRecipe.Ingredients = updatedIngredients // This is already populated from lines 655-667
		existingRecipe.FilterableIngredientNames = []string{}
		uniqueFilterableNames := make(map[string]bool)
		for _, ing := range existingRecipe.Ingredients { // Iterate over the already processed full ingredient strings
			filterable := extractFilterableNames(ing)
			for _, fname := range filterable {
				if fname != "" && !uniqueFilterableNames[fname] {
					existingRecipe.FilterableIngredientNames = append(existingRecipe.FilterableIngredientNames, fname)
					uniqueFilterableNames[fname] = true
				}
			}
		}

		existingRecipe.Method = method
		existingRecipe.UpdatedAt = time.Now().UTC()

		file, errUpload := c.FormFile("photo")
		if errUpload == nil {
			newPhotoFilename = recipeID + "_updated_" + uuid.New().String() + filepath.Ext(file.Filename)
			dst := filepath.Join(uploadsDir, newPhotoFilename)
			if errSave := saveUploadedFile(file, dst); errSave != nil {
				log.Printf("Error saving updated photo file for recipe %s: %v", recipeID, errSave)
				// Decide if this should be a fatal error for the update
			} else {
				existingRecipe.PhotoFilename = newPhotoFilename
				log.Printf("New photo saved for recipe %s: %s", recipeID, newPhotoFilename)
				if oldPhotoFilename != "" && oldPhotoFilename != newPhotoFilename {
					oldPhotoPath := filepath.Join(uploadsDir, oldPhotoFilename)
					if errRemove := os.Remove(oldPhotoPath); errRemove != nil {
						log.Printf("Error deleting old photo %s for recipe %s: %v", oldPhotoPath, recipeID, errRemove)
					} else {
						log.Printf("Old photo deleted for recipe %s: %s", recipeID, oldPhotoPath)
					}
				}
			}
		} else if errUpload != http.ErrMissingFile {
			log.Printf("Error retrieving photo from form during update for recipe %s: %v", recipeID, errUpload)
			// Decide if this should be a fatal error
		}
		// If errUpload == http.ErrMissingFile and no new photo was uploaded,
		// and if existingRecipe.PhotoFilename was empty, set it to placeholder.
		// If it already had a photo, it will keep it.
		if errUpload == http.ErrMissingFile && existingRecipe.PhotoFilename == "" {
			existingRecipe.PhotoFilename = "placeholder.jpg"
		}

		updatedRecipeJSON, errMarshal := json.Marshal(existingRecipe)
		if errMarshal != nil {
			return errMarshal
		}
		if errSet := txn.Set(key, updatedRecipeJSON); errSet != nil {
			return errSet
		}

		// Update ingredient keys
		// For simplicity, this adds new ones but doesn't remove old ones if ingredients are removed.
		// A more robust solution would involve checking which ingredients were removed and potentially decrementing a counter or similar.
		for _, ingredientName := range existingRecipe.Ingredients {
			if strings.TrimSpace(ingredientName) == "" {
				continue
			}
			ingredientKey := []byte("ingredient:" + strings.ToLower(strings.TrimSpace(ingredientName)))
			if errSetIng := txn.Set(ingredientKey, []byte("1")); errSetIng != nil {
				log.Printf("Error saving ingredient key %s during update for recipe %s: %v", ingredientKey, recipeID, errSetIng)
			}
		}
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			log.Printf("Recipe with ID %s not found for update: %v", recipeID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
			return
		}
		log.Printf("Error updating recipe %s in database: %v", recipeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
		return
	}

	log.Printf("Recipe updated successfully: %s", recipeID)
	c.JSON(http.StatusOK, existingRecipe)
}

// DeleteRecipe handles deleting a recipe by ID.
// DELETE /api/v1/recipes/:id
func DeleteRecipe(c *gin.Context) {
	recipeID := c.Param("id")
	var recipeToDelete models.Recipe // To get photo filename for deletion

	err := database.DB.Update(func(txn *badger.Txn) error {
		key := []byte("recipe:" + recipeID)
		item, errGet := txn.Get(key)
		if errGet != nil {
			return errGet
		}
		// Get recipe data to find photo filename before deleting
		errGet = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &recipeToDelete)
		})
		if errGet != nil {
			// Log but proceed with deletion of recipe key if unmarshal fails
			log.Printf("Error unmarshalling recipe %s before deletion (photo might not be deleted): %v", recipeID, errGet)
		}

		return txn.Delete(key)
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			log.Printf("Recipe with ID %s not found for deletion: %v", recipeID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
			return
		}
		log.Printf("Error deleting recipe %s from database: %v", recipeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe"})
		return
	}

	// If recipe was deleted and had a photo, delete the photo file
	if recipeToDelete.PhotoFilename != "" {
		photoPath := filepath.Join(uploadsDir, recipeToDelete.PhotoFilename)
		if errRemove := os.Remove(photoPath); errRemove != nil {
			log.Printf("Error deleting photo file %s for deleted recipe %s: %v", photoPath, recipeID, errRemove)
		} else {
			log.Printf("Photo file deleted for recipe %s: %s", recipeID, photoPath)
		}
	}
	// TODO: Consider logic for removing ingredient keys if they are no longer used by any recipe.

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

	err := database.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("ingredient:")
		// searchPrefix := append(prefix, []byte(query)...) // This was too specific

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			ingredientNameBytes := bytes.TrimPrefix(key, prefix)
			ingredientName := string(ingredientNameBytes)

			if strings.HasPrefix(strings.ToLower(ingredientName), query) {
				matchingIngredients = append(matchingIngredients, ingredientName)
			}
			if len(matchingIngredients) >= 10 { // Limit suggestions
				break
			}
		}
		return nil
	})

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

// MigrateRecipeIngredients handles a one-time migration to populate
// the FilterableIngredientNames field for all existing recipes.
// POST /api/v1/admin/migrate-ingredients (or similar)
func MigrateRecipeIngredients(c *gin.Context) {
	log.Println("Starting ingredient migration for existing recipes...")
	var recipesToUpdate []models.Recipe
	var keysToUpdate [][]byte // Store keys to update in a separate transaction if needed, or update one by one

	// Phase 1: Read all recipes
	errView := database.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true // We need the values
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("recipe:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var recipe models.Recipe
			errValue := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &recipe)
			})
			if errValue != nil {
				log.Printf("Error unmarshalling recipe %s during migration scan: %v", item.Key(), errValue)
				return errValue // Or continue to next item
			}
			recipesToUpdate = append(recipesToUpdate, recipe)
			keysToUpdate = append(keysToUpdate, item.KeyCopy(nil))
		}
		return nil
	})

	if errView != nil {
		log.Printf("Error reading recipes during migration: %v", errView)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read recipes for migration", "details": errView.Error()})
		return
	}

	log.Printf("Found %d recipes to process for migration.", len(recipesToUpdate))
	updatedCount := 0
	errorCount := 0

	// Phase 2: Process and update each recipe
	for i, recipe := range recipesToUpdate {
		key := keysToUpdate[i]

		// Re-initialize FilterableIngredientNames for a clean slate during migration
		recipe.FilterableIngredientNames = []string{}
		uniqueFilterableNames := make(map[string]bool)

		for _, ingStr := range recipe.Ingredients {
			filterable := extractFilterableNames(ingStr)
			for _, fname := range filterable {
				if fname != "" && !uniqueFilterableNames[fname] {
					recipe.FilterableIngredientNames = append(recipe.FilterableIngredientNames, fname)
					uniqueFilterableNames[fname] = true
				}
			}
		}
		recipe.UpdatedAt = time.Now().UTC() // Update timestamp

		recipeJSON, errMarshal := json.Marshal(recipe)
		if errMarshal != nil {
			log.Printf("Error marshalling recipe %s during migration update: %v", recipe.ID, errMarshal)
			errorCount++
			continue // Skip this recipe
		}

		errUpdate := database.DB.Update(func(txn *badger.Txn) error {
			return txn.Set(key, recipeJSON)
		})

		if errUpdate != nil {
			log.Printf("Error updating recipe %s during migration: %v", recipe.ID, errUpdate)
			errorCount++
		} else {
			log.Printf("Successfully migrated ingredients for recipe ID: %s, Name: %s", recipe.ID, recipe.Name)
			updatedCount++
		}
	}

	log.Printf("Ingredient migration finished. Recipes updated: %d, Errors: %d", updatedCount, errorCount)
	c.JSON(http.StatusOK, gin.H{
		"message":         "Ingredient migration process completed.",
		"recipes_found":   len(recipesToUpdate),
		"recipes_updated": updatedCount,
		"errors":          errorCount,
	})
}

// ExportData handles requests to export recipes and/or images.
// POST /api/v1/admin/export
func ExportData(c *gin.Context) {
	log.Println("Starting data export process...")

	var options struct {
		ExportRecipes bool `json:"export_recipes"`
		// RecipeFormat  string `json:"recipe_format"` // Removed as we stick to JSON for now
		ExportImages bool `json:"export_images"`
	}

	if err := c.ShouldBindJSON(&options); err != nil {
		log.Printf("Error binding JSON for export options: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid export options provided", "details": err.Error()})
		return
	}

	log.Printf("Export options received: ExportRecipes=%t, ExportImages=%t", options.ExportRecipes, options.ExportImages)

	if !options.ExportRecipes && !options.ExportImages {
		log.Println("No export type selected (recipes or images).")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please select at least one item to export (recipes or images)."})
		return
	}

	// Fetch all recipes once if either recipes or images are needed.
	var allRecipes []models.Recipe
	if options.ExportRecipes || options.ExportImages {
		errView := database.DB.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = true
			it := txn.NewIterator(opts)
			defer it.Close()
			prefix := []byte("recipe:")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				var recipe models.Recipe
				if err := item.Value(func(val []byte) error { return json.Unmarshal(val, &recipe) }); err != nil {
					log.Printf("Error unmarshalling recipe %s during export scan: %v", item.Key(), err)
					return err
				}
				allRecipes = append(allRecipes, recipe)
			}
			return nil
		})
		if errView != nil {
			log.Printf("Error reading recipes for export: %v", errView)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read recipes for export", "details": errView.Error()})
			return
		}
		if len(allRecipes) == 0 {
			log.Println("No recipes found in the database.")
			// Let specific handlers decide what to do with empty allRecipes list.
		}
	}

	// Case 1: Export both Recipes (JSON) and Images (Combined ZIP)
	if options.ExportRecipes && options.ExportImages {
		log.Println("Combined export: Recipes (JSON) and Images.")
		if len(allRecipes) == 0 {
			log.Println("No recipes (and thus no recipe data or images) to export for combined ZIP.")
			c.JSON(http.StatusOK, gin.H{"message": "No recipes found to include in the export."})
			return
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		// Add recipes.json to ZIP
		jsonData, errMarshal := json.MarshalIndent(allRecipes, "", "  ")
		if errMarshal != nil {
			log.Printf("Error marshalling recipes to JSON for combined ZIP: %v", errMarshal)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal recipes for ZIP"})
			return
		}
		jsonFileInZip, errZipCreate := zipWriter.Create("recipes.json")
		if errZipCreate != nil {
			log.Printf("Error creating recipes.json in ZIP: %v", errZipCreate)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipe file in ZIP"})
			return
		}
		_, errWriteJSON := jsonFileInZip.Write(jsonData)
		if errWriteJSON != nil {
			log.Printf("Error writing recipes.json to ZIP: %v", errWriteJSON)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write recipe data to ZIP"})
			return
		}

		// Add images to ZIP (in an 'images/' folder within the zip)
		uniqueImageFiles := make(map[string]bool)
		for _, recipe := range allRecipes {
			if recipe.PhotoFilename != "" && recipe.PhotoFilename != placeholderImage {
				uniqueImageFiles[recipe.PhotoFilename] = true
			}
		}
		log.Printf("Found %d unique images for combined ZIP.", len(uniqueImageFiles))

		for filename := range uniqueImageFiles {
			filePath := filepath.Join(uploadsDir, filename)
			fileInfo, errStat := os.Stat(filePath)
			if os.IsNotExist(errStat) {
				log.Printf("Image file not found for ZIP, skipping: %s", filePath)
				continue
			}
			if errStat != nil {
				log.Printf("Error stating image file %s for ZIP, skipping: %v", filePath, errStat)
				continue
			}
			if fileInfo.IsDir() {
				log.Printf("Path is a directory, skipping image for ZIP: %s", filePath)
				continue
			}

			imgData, errRead := os.ReadFile(filePath)
			if errRead != nil {
				log.Printf("Error reading image file %s for ZIP: %v", filePath, errRead)
				continue
			}

			// Store images in an "images" folder within the zip
			imageFileInZip, errZipImgCreate := zipWriter.Create("images/" + filename)
			if errZipImgCreate != nil {
				log.Printf("Error creating image entry %s in ZIP: %v", filename, errZipImgCreate)
				continue
			}
			_, errWriteImg := imageFileInZip.Write(imgData)
			if errWriteImg != nil {
				log.Printf("Error writing image %s to ZIP: %v", filename, errWriteImg)
				continue
			}
		}

		if err := zipWriter.Close(); err != nil {
			log.Printf("Error closing zip writer for combined export: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize combined export archive"})
			return
		}
		c.Header("Content-Disposition", "attachment; filename=gorecipes_export.zip")
		c.Header("Content-Type", "application/zip")
		c.Data(http.StatusOK, "application/zip", buf.Bytes())
		log.Println("Combined recipes and images ZIP export completed successfully.")
		return

		// Case 2: Export Recipes only (JSON)
	} else if options.ExportRecipes {
		log.Println("Exporting Recipes only (JSON).")
		if len(allRecipes) == 0 {
			log.Println("No recipes found to export as JSON.")
			// Return empty JSON array for consistency with file download
		}
		jsonData, errMarshal := json.MarshalIndent(allRecipes, "", "  ")
		if errMarshal != nil {
			log.Printf("Error marshalling recipes to JSON for export: %v", errMarshal)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal recipes to JSON"})
			return
		}
		c.Header("Content-Disposition", "attachment; filename=recipes.json")
		c.Data(http.StatusOK, "application/json; charset=utf-8", jsonData)
		log.Println("Recipe JSON export completed successfully.")
		return

		// Case 3: Export Images only (ZIP)
	} else if options.ExportImages {
		log.Println("Exporting Images only (ZIP).")
		uniqueImageFiles := make(map[string]bool)
		for _, recipe := range allRecipes { // allRecipes is already fetched
			if recipe.PhotoFilename != "" && recipe.PhotoFilename != placeholderImage {
				uniqueImageFiles[recipe.PhotoFilename] = true
			}
		}
		if len(uniqueImageFiles) == 0 {
			log.Println("No images found to export.")
			c.JSON(http.StatusOK, gin.H{"message": "No images found to export."})
			return
		}
		log.Printf("Found %d unique images for image-only ZIP.", len(uniqueImageFiles))
		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)
		for filename := range uniqueImageFiles {
			filePath := filepath.Join(uploadsDir, filename)
			fileInfo, errStat := os.Stat(filePath)
			if os.IsNotExist(errStat) {
				log.Printf("Image file not found for ZIP, skipping: %s", filePath)
				continue
			}
			if errStat != nil {
				log.Printf("Error stating image file %s for ZIP, skipping: %v", filePath, errStat)
				continue
			}
			if fileInfo.IsDir() {
				log.Printf("Path is a directory, skipping image for ZIP: %s", filePath)
				continue
			}

			imgData, errRead := os.ReadFile(filePath)
			if errRead != nil {
				log.Printf("Error reading image file %s for zipping: %v", filePath, errRead)
				continue
			}

			f, errZip := zipWriter.Create(filename)
			if errZip != nil {
				log.Printf("Error creating zip entry for %s: %v", filename, errZip)
				continue
			}
			_, errWrite := f.Write(imgData)
			if errWrite != nil {
				log.Printf("Error writing image data to zip for %s: %v", filename, errWrite)
				continue
			}
		}
		if err := zipWriter.Close(); err != nil {
			log.Printf("Error closing zip writer for image-only export: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize image archive"})
			return
		}
		c.Header("Content-Disposition", "attachment; filename=recipe_images.zip")
		c.Header("Content-Type", "application/zip")
		c.Data(http.StatusOK, "application/zip", buf.Bytes())
		log.Println("Image ZIP export completed successfully.")
		return
	}
}
