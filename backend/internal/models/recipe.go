package models

import "time"

// Recipe represents a cooking recipe
type Recipe struct {
	ID                        string    `json:"id"`
	Name                      string    `json:"name"`
	Ingredients               []string  `json:"ingredients"`
	FilterableIngredientNames []string  `json:"filterable_ingredient_names,omitempty"`
	Method                    string    `json:"method"`
	PhotoFilename             string    `json:"photo_filename,omitempty"` // omitempty if no photo
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}
