package models

// ExportedData is a container for all data to be exported or imported.
type ExportedData struct {
	Recipes           []Recipe           `json:"recipes"`
	Ingredients       []Ingredient       `json:"ingredients"`
	RecipeIngredients []RecipeIngredient `json:"recipe_ingredients"`
}
