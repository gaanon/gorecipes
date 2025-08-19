package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"gorecipes/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ProcessRecipePhotoRequest represents the request body for processing a recipe photo
type ProcessRecipePhotoRequest struct {
	ImageData string `json:"image_data"` // Base64 encoded image data
}

// ProcessRecipePhotoResponse represents the response from the photo processing API
type ProcessRecipePhotoResponse struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Method      string   `json:"method"`
}

// ProcessRecipePhoto handles the processing of a recipe photo to extract recipe information
// @Summary Process a recipe photo
// @Description Upload a photo of a recipe to extract recipe details using AI
// @Tags recipes
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Recipe photo"
// @Success 200 {object} ProcessRecipePhotoResponse "Successfully processed photo"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /recipes/process-photo [post]
func ProcessRecipePhoto(c *gin.Context) {
	log.Println("\n=== ProcessRecipePhoto: Starting to process photo with Gemini ===")

	file, fileHeader, err := c.Request.FormFile("photo")
	if err != nil {
		errMsg := fmt.Sprintf("No file uploaded: %v", err)
		log.Println("ProcessRecipePhoto:", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}
	defer file.Close()

	log.Printf("ProcessRecipePhoto: Received file: %s (Size: %d bytes)", fileHeader.Filename, fileHeader.Size)

	geminiService, err := services.NewGeminiService(c.Request.Context())
	if err != nil {
		log.Printf("Error creating Gemini service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize AI service"})
		return
	}

	content, err := geminiService.ProcessRecipeImage(c.Request.Context(), fileHeader)
	if err != nil {
		log.Printf("Error processing image with Gemini: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image with AI"})
		return
	}

	// Clean up the response (sometimes the AI includes markdown code blocks)
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json\n")
	content = strings.TrimSuffix(content, "\n```")

	// Parse the JSON response
	var result ProcessRecipePhotoResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		log.Printf("Error parsing AI response JSON: %v\nResponse was: %s", err, content)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse AI response"})
		return
	}

	c.JSON(http.StatusOK, result)
}
