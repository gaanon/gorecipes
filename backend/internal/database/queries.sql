-- Common queries for GoRecipes PostgreSQL database
-- This file contains frequently used queries that will be implemented in the Go code

-- Get all recipes with ingredient count
SELECT 
    r.id,
    r.name,
    r.method,
    r.photo_filename,
    r.created_at,
    r.updated_at,
    COUNT(ri.ingredient_id) as ingredient_count
FROM recipes r
LEFT JOIN recipe_ingredients ri ON r.id = ri.recipe_id
GROUP BY r.id, r.name, r.method, r.photo_filename, r.created_at, r.updated_at
ORDER BY r.created_at DESC;

-- Get a single recipe with all its ingredients
SELECT 
    r.id,
    r.name,
    r.method,
    r.photo_filename,
    r.created_at,
    r.updated_at,
    ARRAY_AGG(
        json_build_object(
            'ingredient_name', i.name,
            'quantity_text', ri.quantity_text,
            'sort_order', ri.sort_order
        ) ORDER BY ri.sort_order
    ) FILTER (WHERE i.id IS NOT NULL) as ingredients
FROM recipes r
LEFT JOIN recipe_ingredients ri ON r.id = ri.recipe_id
LEFT JOIN ingredients i ON ri.ingredient_id = i.id
WHERE r.id = $1
GROUP BY r.id, r.name, r.method, r.photo_filename, r.created_at, r.updated_at;

-- Search recipes by ingredient names (flexible matching)
SELECT DISTINCT r.id, r.name, r.photo_filename, r.created_at
FROM recipes r
JOIN recipe_ingredients ri ON r.id = ri.recipe_id
JOIN ingredients i ON ri.ingredient_id = i.id
WHERE i.normalized_name ILIKE ANY(ARRAY['%tomato%', '%onion%']) -- Example search terms
ORDER BY r.name;

-- Search recipes that contain ALL specified ingredients
SELECT r.id, r.name, r.photo_filename
FROM recipes r
WHERE r.id IN (
    SELECT ri.recipe_id
    FROM recipe_ingredients ri
    JOIN ingredients i ON ri.ingredient_id = i.id
    WHERE i.normalized_name ILIKE ANY(ARRAY['%flour%', '%eggs%']) -- All required ingredients
    GROUP BY ri.recipe_id
    HAVING COUNT(DISTINCT i.id) = 2 -- Number of required ingredients
)
ORDER BY r.name;

-- Get ingredient autocomplete suggestions
SELECT DISTINCT name
FROM ingredients
WHERE name ILIKE $1 || '%'
ORDER BY 
    CASE WHEN name ILIKE $1 || '%' THEN 1 ELSE 2 END, -- Exact prefix matches first
    LENGTH(name), -- Shorter names first
    name
LIMIT 10;

-- Full-text search on ingredient names
SELECT DISTINCT name,
    ts_rank(to_tsvector('english', name), plainto_tsquery('english', $1)) as rank
FROM ingredients
WHERE to_tsvector('english', name) @@ plainto_tsquery('english', $1)
ORDER BY rank DESC, name
LIMIT 10;

-- Get recipes by date range for meal planning
SELECT 
    mpe.date,
    r.id,
    r.name,
    r.photo_filename
FROM meal_plan_entries mpe
JOIN recipes r ON mpe.recipe_id = r.id
WHERE mpe.date BETWEEN $1 AND $2
ORDER BY mpe.date, r.name;

-- Get most popular ingredients (by usage count)
SELECT 
    i.name,
    COUNT(ri.recipe_id) as usage_count
FROM ingredients i
JOIN recipe_ingredients ri ON i.id = ri.ingredient_id
GROUP BY i.id, i.name
ORDER BY usage_count DESC, i.name
LIMIT 20;

-- Search recipes by name with ranking
SELECT 
    id,
    name,
    photo_filename,
    ts_rank(to_tsvector('english', name), plainto_tsquery('english', $1)) as rank
FROM recipes
WHERE to_tsvector('english', name) @@ plainto_tsquery('english', $1)
ORDER BY rank DESC, name
LIMIT 20;

-- Get recipes that can be made with available ingredients
-- (recipes that have all their ingredients in the provided list)
WITH available_ingredients AS (
    SELECT id FROM ingredients WHERE normalized_name = ANY($1) -- Array of available ingredient names
),
recipe_ingredient_counts AS (
    SELECT 
        ri.recipe_id,
        COUNT(*) as total_ingredients,
        COUNT(ai.id) as available_ingredients_count
    FROM recipe_ingredients ri
    LEFT JOIN available_ingredients ai ON ri.ingredient_id = ai.id
    GROUP BY ri.recipe_id
)
SELECT 
    r.id,
    r.name,
    r.photo_filename,
    ric.total_ingredients,
    ric.available_ingredients_count,
    ROUND((ric.available_ingredients_count::decimal / ric.total_ingredients) * 100, 2) as match_percentage
FROM recipes r
JOIN recipe_ingredient_counts ric ON r.id = ric.recipe_id
WHERE ric.available_ingredients_count = ric.total_ingredients -- Can make with available ingredients
ORDER BY r.name;

-- Partial match: recipes where you have most ingredients
WITH available_ingredients AS (
    SELECT id FROM ingredients WHERE normalized_name = ANY($1)
),
recipe_ingredient_counts AS (
    SELECT 
        ri.recipe_id,
        COUNT(*) as total_ingredients,
        COUNT(ai.id) as available_ingredients_count
    FROM recipe_ingredients ri
    LEFT JOIN available_ingredients ai ON ri.ingredient_id = ai.id
    GROUP BY ri.recipe_id
)
SELECT 
    r.id,
    r.name,
    r.photo_filename,
    ric.total_ingredients,
    ric.available_ingredients_count,
    ROUND((ric.available_ingredients_count::decimal / ric.total_ingredients) * 100, 2) as match_percentage
FROM recipes r
JOIN recipe_ingredient_counts ric ON r.id = ric.recipe_id
WHERE ric.available_ingredients_count > 0
ORDER BY match_percentage DESC, r.name
LIMIT 20;

-- create-comment: Inserts a new comment into the comments table
INSERT INTO comments (id, recipe_id, author, content)
VALUES ($1, $2, $3, $4)
RETURNING id, recipe_id, author, content, created_at, updated_at;

-- get-comments-by-recipe-id: Retrieves all comments for a given recipe_id, ordered chronologically
SELECT id, recipe_id, author, content, created_at, updated_at
FROM comments
WHERE recipe_id = $1
ORDER BY created_at ASC;

-- update-comment: Updates the content of an existing comment
UPDATE comments
SET content = $1, updated_at = NOW()
WHERE id = $2
RETURNING id, recipe_id, author, content, created_at, updated_at;

-- delete-comment: Deletes a comment by its ID
DELETE FROM comments
WHERE id = $1;

