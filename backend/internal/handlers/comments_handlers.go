package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"gorecipes/backend/internal/database"
	"gorecipes/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateCommentHandler handles POST requests to create a new comment.
func CreateCommentHandler(c *gin.Context) {
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe ID is required"})
		return
	}

	var reqBody struct {
		Author  string `json:"author"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body for CreateComment: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if strings.TrimSpace(reqBody.Author) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Author cannot be empty"})
		return
	}
	if strings.TrimSpace(reqBody.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content cannot be empty"})
		return
	}

	comment := models.Comment{
		ID:       uuid.New().String(),
		RecipeID: recipeID,
		Author:   reqBody.Author,
		Content:  reqBody.Content,
	}

	createdComment, err := database.CreateComment(comment)
	if err != nil {
		log.Printf("Error creating comment in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, createdComment)
}

// GetCommentsByRecipeIDHandler handles GET requests to retrieve all comments for a specific recipe.
func GetCommentsByRecipeIDHandler(c *gin.Context) {
	recipeID := c.Param("id")
	if recipeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipe ID is required"})
		return
	}

	comments, err := database.GetCommentsByRecipeID(recipeID)
	if err != nil {
		log.Printf("Error retrieving comments for recipe %s from database: %v", recipeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comments"})
		return
	}

	if comments == nil {
		comments = []models.Comment{} // Ensure we return an empty array, not null
	}

	c.JSON(http.StatusOK, comments)
}

// UpdateCommentHandler handles PUT requests to update an existing comment.
func UpdateCommentHandler(c *gin.Context) {
	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment ID is required"})
		return
	}

	var reqBody struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body for UpdateComment: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if strings.TrimSpace(reqBody.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content cannot be empty"})
		return
	}

	// Fetch existing comment to ensure it exists and get other fields
	existingComment, err := database.GetCommentByID(commentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("Comment with ID %s not found for update: %v", commentID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else {
			log.Printf("Error retrieving comment %s for update: %v", commentID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comment for update"})
		}
		return
	}

	existingComment.Content = reqBody.Content

	updatedComment, err := database.UpdateComment(*existingComment)
	if err != nil {
		log.Printf("Error updating comment %s in database: %v", commentID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, updatedComment)
}

// DeleteCommentHandler handles DELETE requests to delete a comment.
func DeleteCommentHandler(c *gin.Context) {
	commentID := c.Param("id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment ID is required"})
		return
	}

	err := database.DeleteComment(commentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("Comment with ID %s not found (already deleted or never existed): %v", commentID, err)
			c.Status(http.StatusNoContent) // Comment is gone, so operation is effectively successful.
		} else {
			log.Printf("Error deleting comment %s from database: %v", commentID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		}
		return
	}

	log.Printf("Comment deleted successfully: %s", commentID)
	c.Status(http.StatusNoContent)
}
