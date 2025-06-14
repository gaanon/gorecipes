package models

import (
	"time"
)

// Comment represents a comment on a recipe.
type Comment struct {
	ID        string    `json:"id"`
	RecipeID  string    `json:"recipe_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
