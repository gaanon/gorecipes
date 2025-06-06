package database

import (
	"context" // Added for QueryContext
	"database/sql"
	// "errors" // For errors.As - No longer needed after ON CONFLICT DO NOTHING
	"fmt"
	"gorecipes/backend/internal/models"
	"log"
	"strings"
	"time"

	"github.com/lib/pq" // For pq.Array
	"github.com/google/uuid"
)

// extractIngredientNameParts is a placeholder for a utility function
// that will parse an ingredient string (e.g., "1 cup flour") into its quantity ("1 cup")
// and normalized name ("flour"). This will be properly implemented later.
func extractIngredientNameParts(fullIngredient string) (quantity string, name string, err error) {
	parts := strings.SplitN(fullIngredient, " ", 2)
	if len(parts) == 1 {
		return "", strings.ToLower(strings.TrimSpace(parts[0])), nil // Assume it's just the name
	}
	return strings.TrimSpace(parts[0]), strings.ToLower(strings.TrimSpace(parts[1])), nil
}

// normalizeIngredientName is a placeholder for a utility function
// to normalize an ingredient name for consistent storage and searching.
func normalizeIngredientName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// RecipeExistsByID checks if a recipe with the given ID exists in the PostgreSQL database.
func RecipeExistsByID(id string) (bool, error) {
	if DB == nil {
		return false, fmt.Errorf("database not initialized")
	}

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM recipes WHERE id = $1)"
	err := DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking recipe existence for ID %s: %w", id, err)
	}
	return exists, nil
}

// GetRecipeByID retrieves a single recipe by its ID from PostgreSQL,
// including its ingredients.
func GetRecipeByID(id string) (*models.Recipe, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var recipe models.Recipe
	recipeQuery := `
		SELECT r.id, r.name, r.method, r.photo_filename, r.created_at, r.updated_at
		FROM recipes r
		WHERE r.id = $1`

	err := DB.QueryRow(recipeQuery, id).Scan(
		&recipe.ID, &recipe.Name, &recipe.Method, &recipe.PhotoFilename, &recipe.CreatedAt, &recipe.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a specific "not found" error
		}
		return nil, fmt.Errorf("error fetching recipe with ID %s: %w", id, err)
	}

	// Fetch ingredients for the recipe
	ingredientsQuery := `
		SELECT ri.quantity_text, i.name
		FROM recipe_ingredients ri
		JOIN ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = $1
		ORDER BY ri.sort_order ASC`

	rows, err := DB.Query(ingredientsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("error fetching ingredients for recipe ID %s: %w", id, err)
	}
	defer rows.Close()

	var ingredients []string
	for rows.Next() {
		var quantityText, ingredientName string
		if err := rows.Scan(&quantityText, &ingredientName); err != nil {
			return nil, fmt.Errorf("error scanning ingredient for recipe ID %s: %w", id, err)
		}
		if quantityText != "" {
			ingredients = append(ingredients, fmt.Sprintf("%s %s", quantityText, ingredientName))
		} else {
			ingredients = append(ingredients, ingredientName) // Handle cases where quantity might be empty
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ingredients for recipe ID %s: %w", id, err)
	}

	recipe.Ingredients = ingredients

	return &recipe, nil
}

// CreateRecipe adds a new recipe to the PostgreSQL database.
// It handles creating the recipe, ingredients, and their associations.
func CreateRecipe(recipe *models.Recipe) (*models.Recipe, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Prepare recipe data
	if recipe.ID == "" {
		recipe.ID = uuid.NewString()
	}
	recipe.CreatedAt = time.Now().UTC()
	recipe.UpdatedAt = recipe.CreatedAt

	// Insert into recipes table
	recipeQuery := `INSERT INTO recipes (id, name, method, photo_filename, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(recipeQuery, recipe.ID, recipe.Name, recipe.Method, recipe.PhotoFilename, recipe.CreatedAt, recipe.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert recipe ID %s: %w", recipe.ID, err)
	}

	// Process and insert ingredients
	for i, fullIngredientStr := range recipe.Ingredients {
		quantityText, ingredientNamePart, err := extractIngredientNameParts(fullIngredientStr)
		if err != nil {
			log.Printf("Error parsing ingredient string '%s': %v. Skipping.", fullIngredientStr, err)
			// Depending on desired behavior, you might want to return an error here
			continue
		}
		normalizedIngredientName := normalizeIngredientName(ingredientNamePart)

		var ingredientID string
		// Check if ingredient exists, otherwise create it
		ingredientQuery := `SELECT id FROM ingredients WHERE name = $1`
		err = tx.QueryRow(ingredientQuery, normalizedIngredientName).Scan(&ingredientID)
		if err == sql.ErrNoRows {
			ingredientID = uuid.NewString()
			insertIngredientQuery := `INSERT INTO ingredients (id, name, created_at, updated_at)
				VALUES ($1, $2, $3, $4)`
			_, err = tx.Exec(insertIngredientQuery, ingredientID, normalizedIngredientName, time.Now().UTC(), time.Now().UTC())
			if err != nil {
				return nil, fmt.Errorf("failed to insert new ingredient '%s': %w", normalizedIngredientName, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to query ingredient '%s': %w", normalizedIngredientName, err)
		}

		// Insert into recipe_ingredients junction table
		recipeIngredientID := uuid.NewString()
		insertRecipeIngredientQuery := `INSERT INTO recipe_ingredients (id, recipe_id, ingredient_id, quantity_text, sort_order)
			VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(insertRecipeIngredientQuery, recipeIngredientID, recipe.ID, ingredientID, quantityText, i)
		if err != nil {
			return nil, fmt.Errorf("failed to insert recipe_ingredient link for recipe ID %s and ingredient ID %s: %w", recipe.ID, ingredientID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return recipe, nil
}

// GetAllRecipes retrieves recipes with optional search, ingredient filtering, and pagination.
func GetAllRecipes(searchTerm string, ingredientFilters []string, page int, pageSize int) ([]models.Recipe, int, error) {
	if DB == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	var args []interface{}
	argCount := 1

	// Ingredient filters are used directly from input (already pre-processed by handler)
	// plainto_tsquery will handle further normalization for tsvector matching.

	// Base query for fetching recipes
	selectSQL := `SELECT r.id, r.name, r.method, r.photo_filename, r.created_at, r.updated_at,
		(
			SELECT COALESCE(array_agg(ri_s.quantity_text || ' ' || i_s.name ORDER BY ri_s.sort_order ASC), '{}'::TEXT[])
			FROM recipe_ingredients ri_s
			JOIN ingredients i_s ON ri_s.ingredient_id = i_s.id
			WHERE ri_s.recipe_id = r.id
		) AS ingredients_list
		FROM recipes r`

	// Base query for counting total matching recipes
	countSQL := `SELECT COUNT(DISTINCT r.id) FROM recipes r`

	// WHERE clauses and JOINs for filtering
	var conditions []string
	var joinClauses string

	if searchTerm != "" {
		conditions = append(conditions, fmt.Sprintf("r.search_vector @@ plainto_tsquery('english', $%d)", argCount))
		args = append(args, searchTerm)
		argCount++
	}

	if len(ingredientFilters) > 0 {
		for i, filterTerm := range ingredientFilters {
			// Each filterTerm must match an ingredient in the recipe.
			// We add a set of JOINs for each filterTerm to ensure AND logic.
			ingredientAlias := fmt.Sprintf("i_f%d", i)
			recipeIngredientAlias := fmt.Sprintf("ri_f%d", i)

			joinSQLPart := fmt.Sprintf(`
				JOIN recipe_ingredients %s ON r.id = %s.recipe_id
				JOIN ingredients %s ON %s.ingredient_id = %s.id AND %s.normalized_name_tsvector @@ plainto_tsquery('english', $%d)`,
				recipeIngredientAlias, recipeIngredientAlias,
				ingredientAlias, recipeIngredientAlias, ingredientAlias,
				ingredientAlias, argCount)
			
			joinClauses += joinSQLPart
			args = append(args, filterTerm)
			argCount++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Construct final count query
	finalCountQuery := countSQL + joinClauses + whereClause
	var totalCount int
    // Re-evaluate arg slicing for count query
    currentArgsForCount := []interface{}{}
    if searchTerm != "" {
        currentArgsForCount = append(currentArgsForCount, searchTerm)
    }
    if len(ingredientFilters) > 0 {
        for _, filterTerm := range ingredientFilters {
            currentArgsForCount = append(currentArgsForCount, filterTerm)
        }
    }
	err := DB.QueryRow(finalCountQuery, currentArgsForCount...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting recipes: %w", err)
	}

	if totalCount == 0 {
		return []models.Recipe{}, 0, nil
	}

	// Construct final select query with ordering and pagination
	orderByClause := " ORDER BY r.updated_at DESC"
	offset := (page - 1) * pageSize
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	finalSelectQuery := selectSQL + joinClauses + whereClause + orderByClause + paginationClause
	args = append(args, pageSize, offset)

	rows, err := DB.Query(finalSelectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching recipes: %w", err)
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		var recipe models.Recipe
		var ingredientsList pq.StringArray
		if err := rows.Scan(
			&recipe.ID, &recipe.Name, &recipe.Method, &recipe.PhotoFilename, 
			&recipe.CreatedAt, &recipe.UpdatedAt, &ingredientsList,
		); err != nil {
			return nil, 0, fmt.Errorf("error scanning recipe row: %w", err)
		}
		recipe.Ingredients = []string(ingredientsList)
		recipes = append(recipes, recipe)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating recipe rows: %w", err)
	}

	return recipes, totalCount, nil
}

// UpdateRecipe updates an existing recipe in the PostgreSQL database.
func UpdateRecipe(recipe *models.Recipe) (*models.Recipe, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if recipe.ID == "" {
		return nil, fmt.Errorf("recipe ID cannot be empty for update")
	}

	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update recipe's main fields
	recipe.UpdatedAt = time.Now().UTC()
	updateRecipeQuery := `UPDATE recipes SET name = $1, method = $2, photo_filename = $3, updated_at = $4
		WHERE id = $5`
	res, err := tx.Exec(updateRecipeQuery, recipe.Name, recipe.Method, recipe.PhotoFilename, recipe.UpdatedAt, recipe.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update recipe ID %s: %w", recipe.ID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected for recipe ID %s: %w", recipe.ID, err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("recipe with ID %s not found for update", recipe.ID) // Or use a specific error type
	}

	// Delete existing ingredients for this recipe
	deleteIngredientsQuery := `DELETE FROM recipe_ingredients WHERE recipe_id = $1`
	_, err = tx.Exec(deleteIngredientsQuery, recipe.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old ingredients for recipe ID %s: %w", recipe.ID, err)
	}

	// Process and insert new ingredients (similar to CreateRecipe)
	for i, fullIngredientStr := range recipe.Ingredients {
		quantityText, ingredientNamePart, err := extractIngredientNameParts(fullIngredientStr)
		if err != nil {
			log.Printf("Error parsing ingredient string '%s' during update: %v. Skipping.", fullIngredientStr, err)
			continue
		}
		normalizedIngredientName := normalizeIngredientName(ingredientNamePart)

		var ingredientID string
		ingredientQuery := `SELECT id FROM ingredients WHERE name = $1`
		err = tx.QueryRow(ingredientQuery, normalizedIngredientName).Scan(&ingredientID)
		if err == sql.ErrNoRows {
			ingredientID = uuid.NewString()
			insertIngredientQuery := `INSERT INTO ingredients (id, name, created_at, updated_at)
				VALUES ($1, $2, $3, $4)`
			_, err = tx.Exec(insertIngredientQuery, ingredientID, normalizedIngredientName, time.Now().UTC(), time.Now().UTC())
			if err != nil {
				return nil, fmt.Errorf("failed to insert new ingredient '%s' during update: %w", normalizedIngredientName, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to query ingredient '%s' during update: %w", normalizedIngredientName, err)
		}

		recipeIngredientID := uuid.NewString()
		insertRecipeIngredientQuery := `INSERT INTO recipe_ingredients (id, recipe_id, ingredient_id, quantity_text, sort_order)
			VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(insertRecipeIngredientQuery, recipeIngredientID, recipe.ID, ingredientID, quantityText, i)
		if err != nil {
			return nil, fmt.Errorf("failed to insert recipe_ingredient link for recipe ID %s and ingredient ID %s during update: %w", recipe.ID, ingredientID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction for recipe update: %w", err)
	}

	return recipe, nil
}

// GetAllRecipesForExport fetches all recipes from the database without pagination or filtering, for export purposes.
func GetAllRecipesForExport() ([]models.Recipe, error) {
	rows, err := DB.QueryContext(context.Background(), `SELECT id, name, method, photo_filename, created_at, updated_at FROM recipes ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("error querying all recipes for export: %w", err)
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		var r models.Recipe
		var photoFilename sql.NullString // Handle potentially NULL photo_filename
		if err := rows.Scan(&r.ID, &r.Name, &r.Method, &photoFilename, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning recipe for export: %w", err)
		}
		if photoFilename.Valid {
			r.PhotoFilename = photoFilename.String
		} else {
			r.PhotoFilename = "" // Or your desired default for NULL photo_filename
		}
		// The Recipe struct's Ingredients field ([]string) is not populated here as it's a denormalized representation.
		// For export, we fetch recipe_ingredients separately.
		recipes = append(recipes, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recipe rows for export: %w", err)
	}
	return recipes, nil
}

// GetAllRecipeIngredients fetches all recipe_ingredients records from the database.
func GetAllRecipeIngredients() ([]models.RecipeIngredient, error) {
	rows, err := DB.QueryContext(context.Background(), `SELECT id, recipe_id, ingredient_id, quantity_text, sort_order FROM recipe_ingredients ORDER BY recipe_id ASC, sort_order ASC`)
	if err != nil {
		return nil, fmt.Errorf("error querying recipe_ingredients: %w", err)
	}
	defer rows.Close()

	var recipeIngredients []models.RecipeIngredient
	for rows.Next() {
		var ri models.RecipeIngredient
		var quantityText sql.NullString // Handle potentially NULL quantity_text
		if err := rows.Scan(&ri.ID, &ri.RecipeID, &ri.IngredientID, &quantityText, &ri.SortOrder); err != nil {
			return nil, fmt.Errorf("error scanning recipe_ingredient: %w", err)
		}
		if quantityText.Valid {
			ri.QuantityText = quantityText.String
		} else {
			ri.QuantityText = ""
		}
		recipeIngredients = append(recipeIngredients, ri)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recipe_ingredient rows: %w", err)
	}
	return recipeIngredients, nil
}

// GetAllIngredients fetches all ingredients from the database.
func GetAllIngredients() ([]models.Ingredient, error) {
	rows, err := DB.QueryContext(context.Background(), `SELECT id, name, normalized_name, created_at, updated_at FROM ingredients ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("error querying ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		var i models.Ingredient
		if err := rows.Scan(&i.ID, &i.Name, &i.NormalizedName, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning ingredient: %w", err)
		}
		ingredients = append(ingredients, i)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ingredient rows: %w", err)
	}
	return ingredients, nil
}

// DeleteRecipe removes a recipe from the PostgreSQL database.
func DeleteRecipe(id string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	if id == "" {
		return fmt.Errorf("recipe ID cannot be empty for deletion")
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, delete from recipe_ingredients (junction table)
	deleteIngredientsQuery := `DELETE FROM recipe_ingredients WHERE recipe_id = $1`
	_, err = tx.Exec(deleteIngredientsQuery, id)
	if err != nil {
		// It's okay if there were no ingredients, but other errors should be reported
		log.Printf("Warning: could not delete recipe_ingredients for recipe ID %s (may not have had any): %v", id, err)
		// Depending on strictness, you might choose to return error here or proceed.
		// For now, we proceed to delete the main recipe entry.
	}

	// Then, delete from recipes table
	deleteRecipeQuery := `DELETE FROM recipes WHERE id = $1`
	res, err := tx.Exec(deleteRecipeQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete recipe ID %s: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// This error is less critical if the delete query itself didn't fail, 
		// but good to log for diagnostics.
		log.Printf("Warning: could not get rows affected after deleting recipe ID %s: %v", id, err)
	}
	if rowsAffected == 0 {
		// If no rows were affected, the recipe didn't exist. 
		// This might not be an error condition depending on desired idempotency.
		// For now, we'll consider it a success if no error occurred during exec.
		log.Printf("Recipe with ID %s not found for deletion, or already deleted.", id)
	}

	// Note: Orphaned ingredients in the 'ingredients' table are not cleaned up here.
	// This could be a separate maintenance task if desired.

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for recipe deletion: %w", err)
	}

	log.Printf("Recipe deleted successfully (or did not exist): ID=%s", id)
	return nil
}

// ImportRecipeDataBundle handles the import of recipes, ingredients, and their links
// within a single database transaction.
// It returns counts of successfully imported items or an error if the process fails.
func ImportRecipeDataBundle(data models.ExportedData) (importedRecipes int, importedIngredients int, importedLinks int, err error) {
	if DB == nil {
		return 0, 0, 0, fmt.Errorf("database not initialized")
	}

	tx, err := DB.Begin()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // if commit fails, err will be set
		}
	}()

	// Maps to store original ID (from JSON) to new/existing DB ID (UUID string)
	ingredientOriginalIDToDbIDMap := make(map[string]string)
	recipeOriginalIDToDbIDMap := make(map[string]string)

	// 1. Import Ingredients
	for _, ingFromFile := range data.Ingredients {
		dbIngredientID, createErr := getOrCreateIngredientTx(tx, ingFromFile)
		if createErr != nil {
			err = fmt.Errorf("error processing ingredient '%s': %w", ingFromFile.Name, createErr)
			return
		}
		ingredientOriginalIDToDbIDMap[ingFromFile.ID] = dbIngredientID
		importedIngredients++
	}
	log.Printf("Processed %d ingredients. Map size: %d", len(data.Ingredients), len(ingredientOriginalIDToDbIDMap))


	// 2. Import Recipes
	for _, recFromFile := range data.Recipes {
		dbRecipeID, createErr := getOrCreateRecipeTx(tx, recFromFile)
		if createErr != nil {
			err = fmt.Errorf("error processing recipe '%s': %w", recFromFile.Name, createErr)
			return
		}
		recipeOriginalIDToDbIDMap[recFromFile.ID] = dbRecipeID
		importedRecipes++
	}
	log.Printf("Processed %d recipes. Map size: %d", len(data.Recipes), len(recipeOriginalIDToDbIDMap))

	// 3. Import Recipe-Ingredient Links
	for _, riFromFile := range data.RecipeIngredients {
		createErr := insertRecipeIngredientLinkTx(tx, riFromFile, recipeOriginalIDToDbIDMap, ingredientOriginalIDToDbIDMap)
		if createErr != nil {
			// Any error from insertRecipeIngredientLinkTx is now considered fatal
			// as ON CONFLICT DO NOTHING should handle duplicates silently.
			err = fmt.Errorf("error processing recipe_ingredient link for recipe '%s' and ingredient '%s': %w", riFromFile.RecipeID, riFromFile.IngredientID, createErr)
			return
		}
		importedLinks++
	}
	log.Printf("Processed %d recipe_ingredient links.", len(data.RecipeIngredients))

	return // err will be nil if commit succeeds, or set by defer if commit fails or rollback occurs
}

// getOrCreateIngredientTx finds an ingredient by its normalized name or creates it if not found.
// Operates within a transaction. Returns the database ID of the ingredient.
// The input ingredient's NormalizedName should be pre-populated if known, otherwise it relies on the DB trigger.
func getOrCreateIngredientTx(tx *sql.Tx, ingredient models.Ingredient) (string, error) {
	var dbIngredientID string
	var existingNormalizedName string // To store what the DB generates/has

	// The schema has a trigger to set normalized_name from name on insert/update.
	// For checking existence, we rely on the normalized_name from the import file if present,
	// or we can try to normalize the name here similarly to the DB function for a better chance of matching.
	// For simplicity, we'll assume `ingredient.NormalizedName` from the JSON is reliable for lookup.
	// If not, we might need to query by name and then compare normalized versions, or just insert and let unique constraints handle it.

	query := `SELECT id, normalized_name FROM ingredients WHERE normalized_name = $1`
	err := tx.QueryRow(query, ingredient.NormalizedName).Scan(&dbIngredientID, &existingNormalizedName)

	if err == sql.ErrNoRows { // Ingredient does not exist, create it
		newID := uuid.NewString()
		insertQuery := `INSERT INTO ingredients (id, name, created_at, updated_at)
						VALUES ($1, $2, $3, $4) RETURNING id, normalized_name`
		// Note: normalized_name is set by a trigger using the 'name' field.
		// We pass ingredient.Name and expect the trigger to work.
		now := time.Now().UTC()
		err = tx.QueryRow(insertQuery, newID, ingredient.Name, now, now).Scan(&dbIngredientID, &existingNormalizedName)
		if err != nil {
			return "", fmt.Errorf("failed to insert new ingredient '%s': %w", ingredient.Name, err)
		}
		log.Printf("Created new ingredient: Name='%s', DB_ID='%s', Normalized='%s'", ingredient.Name, dbIngredientID, existingNormalizedName)
		return dbIngredientID, nil
	} else if err != nil { // Other query error
		return "", fmt.Errorf("failed to query for existing ingredient '%s' (normalized: '%s'): %w", ingredient.Name, ingredient.NormalizedName, err)
	}

	// Ingredient exists
	log.Printf("Found existing ingredient: Name='%s', DB_ID='%s', Normalized_Lookup='%s', Normalized_DB='%s'", ingredient.Name, dbIngredientID, ingredient.NormalizedName, existingNormalizedName)
	return dbIngredientID, nil
}

// getOrCreateRecipeTx finds a recipe by its name or creates it if not found.
// Operates within a transaction. Returns the database ID of the recipe.
func getOrCreateRecipeTx(tx *sql.Tx, recipe models.Recipe) (string, error) {
	var dbRecipeID string
	query := `SELECT id FROM recipes WHERE name = $1`
	err := tx.QueryRow(query, recipe.Name).Scan(&dbRecipeID)

	if err == sql.ErrNoRows { // Recipe does not exist, create it
		newID := uuid.NewString()
		insertQuery := `INSERT INTO recipes (id, name, method, photo_filename, created_at, updated_at)
						VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
		now := time.Now().UTC()
		// Handle empty photo_filename from import gracefully
		var photoFilename sql.NullString
		if recipe.PhotoFilename != "" {
			photoFilename = sql.NullString{String: recipe.PhotoFilename, Valid: true}
		}

		err = tx.QueryRow(insertQuery, newID, recipe.Name, recipe.Method, photoFilename, now, now).Scan(&dbRecipeID)
		if err != nil {
			return "", fmt.Errorf("failed to insert new recipe '%s': %w", recipe.Name, err)
		}
		log.Printf("Created new recipe: Name='%s', DB_ID='%s'", recipe.Name, dbRecipeID)
		return dbRecipeID, nil
	} else if err != nil { // Other query error
		return "", fmt.Errorf("failed to query for existing recipe '%s': %w", recipe.Name, err)
	}
	log.Printf("Found existing recipe: Name='%s', DB_ID='%s'", recipe.Name, dbRecipeID)
	return dbRecipeID, nil
}

// insertRecipeIngredientLinkTx inserts a link between a recipe and an ingredient.
// Operates within a transaction. Uses maps to resolve original JSON IDs to current DB IDs.
func insertRecipeIngredientLinkTx(tx *sql.Tx, ri models.RecipeIngredient, recipeOriginalIDToDbIDMap map[string]string, ingredientOriginalIDToDbIDMap map[string]string) error {
	dbRecipeID, okRecipe := recipeOriginalIDToDbIDMap[ri.RecipeID]
	if !okRecipe {
		return fmt.Errorf("could not find DB ID for original recipe ID '%s'", ri.RecipeID)
	}

	dbIngredientID, okIngredient := ingredientOriginalIDToDbIDMap[ri.IngredientID]
	if !okIngredient {
		return fmt.Errorf("could not find DB ID for original ingredient ID '%s'", ri.IngredientID)
	}

	newLinkID := uuid.NewString()
	// Handle empty quantity_text from import gracefully
	var quantityText sql.NullString
	if ri.QuantityText != "" {
		quantityText = sql.NullString{String: ri.QuantityText, Valid: true}
	}

	insertQuery := `INSERT INTO recipe_ingredients (id, recipe_id, ingredient_id, quantity_text, sort_order)
					VALUES ($1, $2, $3, $4, $5) ON CONFLICT (recipe_id, ingredient_id) DO NOTHING`
	_, err := tx.Exec(insertQuery, newLinkID, dbRecipeID, dbIngredientID, quantityText, ri.SortOrder)
	if err != nil {
		// The caller will check for unique_violation (pq.ErrorCode("23505"))
		return fmt.Errorf("failed to insert recipe_ingredient link (RecipeDB_ID: %s, IngredientDB_ID: %s): %w", dbRecipeID, dbIngredientID, err)
	}
	// log.Printf("Created recipe_ingredient link: RecipeDB_ID='%s', IngredientDB_ID='%s', Qty='%s'", dbRecipeID, dbIngredientID, ri.QuantityText)
	return nil
}

