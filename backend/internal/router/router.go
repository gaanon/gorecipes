package router

import (
	"gorecipes/backend/internal/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter initializes and returns a new Gin router.
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// CORS Middleware Configuration
	// Allows requests from SvelteKit dev server (typically http://localhost:5173)
	// and common production/preview ports.
	// Adjust origins as needed for your deployment.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:4173", "http://192.168.1.45:5173"}, // Add other origins if needed
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API v1 group
	apiV1 := router.Group("/api/v1")
	{
		// Recipe routes
		recipesBase := apiV1.Group("/recipes")
		{
			recipesBase.POST("", handlers.CreateRecipe) // POST /api/v1/recipes
			recipesBase.GET("", handlers.ListRecipes)   // GET  /api/v1/recipes

			// Routes for a specific recipe, e.g., /api/v1/recipes/:id
			recipeWithID := recipesBase.Group("/:id")
			{
				recipeWithID.GET("", handlers.GetRecipe)       // GET    /api/v1/recipes/:id
				recipeWithID.PUT("", handlers.UpdateRecipe)    // PUT    /api/v1/recipes/:id
				recipeWithID.DELETE("", handlers.DeleteRecipe) // DELETE /api/v1/recipes/:id
				// recipeWithID.POST("/image", handlers.UploadRecipeImage) // Example for specific image upload
			}
			// Comment routes nested under a specific recipe
			recipeWithID.POST("/comments", handlers.CreateCommentHandler)        // POST /api/v1/recipes/:id/comments
			recipeWithID.GET("/comments", handlers.GetCommentsByRecipeIDHandler) // GET /api/v1/recipes/:id/comments
		}

		// Comment routes (for specific comment operations)
		comments := apiV1.Group("/comments")
		{
			comments.PUT("/:id", handlers.UpdateCommentHandler)    // PUT    /api/v1/comments/:id
			comments.DELETE("/:id", handlers.DeleteCommentHandler) // DELETE /api/v1/comments/:id
		}

		// Ingredient routes
		ingredients := apiV1.Group("/ingredients")
		{
			ingredients.GET("", handlers.GetIngredientsAutocomplete) // e.g., /api/v1/ingredients?q=tomato
		}

		// Admin routes (currently no admin-specific routes defined)
		admin := apiV1.Group("/admin")
		{
			admin.POST("/export", handlers.ExportData) // POST /api/v1/admin/export
			admin.POST("/import", handlers.ImportData) // POST /api/v1/admin/import
		}

		// Meal Planner routes
		mealPlanner := apiV1.Group("/mealplanner")
		{
			mealPlanner.POST("/entries", handlers.CreateMealPlanEntryHandler)             // POST /api/v1/mealplanner/entries
			mealPlanner.GET("/entries", handlers.ListMealPlanEntriesHandler)              // GET  /api/v1/mealplanner/entries
			mealPlanner.DELETE("/entries/:entry_id", handlers.DeleteMealPlanEntryHandler) // DELETE /api/v1/mealplanner/entries/:entry_id
		}
	}

	// Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://localhost:8080/swagger/doc.json")))

	// Serve static files (uploaded images)
	// The path "/uploads/images" will correspond to the "uploads/images" directory in the backend.
	// Ensure this directory is relative to where the Go binary is run (usually the 'backend' directory).
	router.Static("/uploads/images", "./uploads/images")

	// Simple health check endpoint (can be outside the API group or within, depending on preference)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	return router
}
