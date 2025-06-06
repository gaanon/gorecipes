package models

import "time"

// Ingredient represents a unique ingredient item.
type Ingredient struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	NormalizedName string    `json:"normalized_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RecipeIngredient represents the link between a recipe and an ingredient,
// including quantity and order.
type RecipeIngredient struct {
	ID           string `json:"id"`
	RecipeID     string `json:"recipe_id"`
	IngredientID string `json:"ingredient_id"`
	QuantityText string `json:"quantity_text,omitempty"`
	SortOrder    int    `json:"sort_order"`
}
