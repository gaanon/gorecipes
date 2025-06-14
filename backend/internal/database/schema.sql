-- PostgreSQL Schema for GoRecipes Application
-- This file contains the complete database schema including tables, indexes, and constraints

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create recipes table
CREATE TABLE IF NOT EXISTS recipes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    method TEXT NOT NULL,
    photo_filename VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create ingredients table
CREATE TABLE IF NOT EXISTS ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    normalized_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Ensure 'updated_at' column exists in 'ingredients' table for existing tables
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'ingredients'
          AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE ingredients ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
        RAISE NOTICE 'Column updated_at added to ingredients table by conditional DDL.';
    ELSE
        RAISE NOTICE 'Column updated_at already exists in ingredients table or was added by CREATE TABLE.';
    END IF;
END $$;

-- Create recipe_ingredients junction table
CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity_text VARCHAR(500), -- Store original quantity string like "2 cups", "1 large"
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(recipe_id, ingredient_id) -- Prevent duplicate ingredient assignments
);

-- Create meal_plan_entries table
CREATE TABLE IF NOT EXISTS meal_plan_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipe_id TEXT NOT NULL, -- Changed from UUID to TEXT to allow custom recipe names
    date DATE NOT NULL,
    notes TEXT NULL, -- Added notes column
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(recipe_id, date) -- Prevent duplicate recipe assignments for the same date
);

-- Create comments table
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    author TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Remove the foreign key constraint if it exists to allow custom recipe names
-- This allows meal_plan_entries.recipe_id to be either a UUID (for real recipes) or a custom string
DO $$
BEGIN
    -- Drop foreign key constraint if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'meal_plan_entries_recipe_id_fkey' 
        AND table_name = 'meal_plan_entries'
    ) THEN
        ALTER TABLE meal_plan_entries DROP CONSTRAINT meal_plan_entries_recipe_id_fkey;
        RAISE NOTICE 'Foreign key constraint meal_plan_entries_recipe_id_fkey dropped to allow custom recipe names.';
    END IF;
    
    -- Change column type from UUID to TEXT if it's currently UUID
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'meal_plan_entries' 
        AND column_name = 'recipe_id' 
        AND data_type = 'uuid'
    ) THEN
        ALTER TABLE meal_plan_entries ALTER COLUMN recipe_id TYPE TEXT;
        RAISE NOTICE 'Column recipe_id in meal_plan_entries changed from UUID to TEXT to support custom recipe names.';
    END IF;
END $$;

-- Create indexes for performance

-- Recipes indexes
CREATE INDEX IF NOT EXISTS idx_recipes_name ON recipes USING GIN (to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_recipes_created_at ON recipes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recipes_updated_at ON recipes(updated_at DESC);

-- Ingredients indexes
CREATE INDEX IF NOT EXISTS idx_ingredients_name ON ingredients(name);
CREATE INDEX IF NOT EXISTS idx_ingredients_normalized_name ON ingredients(normalized_name);
CREATE INDEX IF NOT EXISTS idx_ingredients_name_gin ON ingredients USING GIN (to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_ingredients_normalized_name_gin ON ingredients USING GIN (to_tsvector('english', normalized_name));

-- Recipe ingredients indexes
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_id ON recipe_ingredients(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_ingredient_id ON recipe_ingredients(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_sort_order ON recipe_ingredients(recipe_id, sort_order);

-- Meal plan entries indexes
CREATE INDEX IF NOT EXISTS idx_meal_plan_entries_date ON meal_plan_entries(date DESC);
CREATE INDEX IF NOT EXISTS idx_meal_plan_entries_recipe_id ON meal_plan_entries(recipe_id);
CREATE INDEX IF NOT EXISTS idx_meal_plan_entries_date_range ON meal_plan_entries(date, recipe_id);

-- Create a function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at for recipes
DROP TRIGGER IF EXISTS update_recipes_updated_at ON recipes;
CREATE TRIGGER update_recipes_updated_at BEFORE UPDATE ON recipes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create trigger to automatically update updated_at for ingredients
DROP TRIGGER IF EXISTS update_ingredients_updated_at ON ingredients;
CREATE TRIGGER update_ingredients_updated_at BEFORE UPDATE ON ingredients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create trigger to automatically update updated_at for comments
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create a view for recipes with ingredient count (useful for API responses)
CREATE OR REPLACE VIEW recipes_with_stats AS
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
GROUP BY r.id, r.name, r.method, r.photo_filename, r.created_at, r.updated_at;

-- Create a view for ingredient usage statistics
CREATE OR REPLACE VIEW ingredient_usage_stats AS
SELECT 
    i.id,
    i.name,
    i.normalized_name,
    i.created_at,
    COUNT(ri.recipe_id) as usage_count
FROM ingredients i
LEFT JOIN recipe_ingredients ri ON i.id = ri.ingredient_id
GROUP BY i.id, i.name, i.normalized_name, i.created_at
ORDER BY usage_count DESC, i.name;

-- Function to normalize ingredient names for consistent searching
CREATE OR REPLACE FUNCTION normalize_ingredient_name(input_name TEXT)
RETURNS TEXT AS $$
BEGIN
    -- Convert to lowercase, trim whitespace, and remove common descriptors
    RETURN TRIM(LOWER(
        REGEXP_REPLACE(
            REGEXP_REPLACE(
                REGEXP_REPLACE(input_name, '\b(fresh|dried|frozen|canned|cooked|raw|chopped|diced|sliced|minced|grated|large|medium|small|whole)\b', '', 'gi'),
                '\d+\s*(g|kg|mg|oz|lb|lbs|ml|l|cl|dl|tsp|tbsp|cup|cups|pt|qt|gal)\b', '', 'gi'
            ),
            '\s+', ' ', 'g'
        )
    ));
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create a trigger to automatically set normalized_name when inserting ingredients
CREATE OR REPLACE FUNCTION set_normalized_ingredient_name()
RETURNS TRIGGER AS $$
BEGIN
    NEW.normalized_name = normalize_ingredient_name(NEW.name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_set_normalized_ingredient_name ON ingredients;
CREATE TRIGGER trigger_set_normalized_ingredient_name
    BEFORE INSERT OR UPDATE ON ingredients
    FOR EACH ROW
    EXECUTE FUNCTION set_normalized_ingredient_name();

-- Add tsvector column for full-text search on normalized ingredient names
-- Ensure this is idempotent for repeated script execution
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'ingredients'
          AND column_name = 'normalized_name_tsvector'
    ) THEN
        ALTER TABLE ingredients ADD COLUMN normalized_name_tsvector TSVECTOR;
        RAISE NOTICE 'Column normalized_name_tsvector added to ingredients table.';
    ELSE
        RAISE NOTICE 'Column normalized_name_tsvector already exists in ingredients table.';
    END IF;
END $$;

-- Create GIN index on the new tsvector column
CREATE INDEX IF NOT EXISTS idx_ingredients_normalized_name_tsvector ON ingredients USING GIN(normalized_name_tsvector);

-- Function to update the normalized_name_tsvector column
CREATE OR REPLACE FUNCTION update_normalized_name_tsvector()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure this runs after normalized_name is set/updated
    NEW.normalized_name_tsvector = to_tsvector('english', COALESCE(NEW.normalized_name, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update normalized_name_tsvector on insert or update
DROP TRIGGER IF EXISTS trigger_update_normalized_name_tsvector ON ingredients;
CREATE TRIGGER trigger_update_normalized_name_tsvector
    BEFORE INSERT OR UPDATE ON ingredients
    FOR EACH ROW
    -- WHEN (pg_trigger_depth() = 0) -- Consider if needed based on other triggers
    EXECUTE FUNCTION update_normalized_name_tsvector();

-- Comments for documentation
COMMENT ON TABLE recipes IS 'Core recipe information including name, cooking method, and photo';
COMMENT ON TABLE ingredients IS 'Master list of all ingredients used across recipes';
COMMENT ON TABLE recipe_ingredients IS 'Junction table linking recipes to their ingredients with quantities';
COMMENT ON TABLE meal_plan_entries IS 'Meal planning entries associating recipes with specific dates';
COMMENT ON TABLE comments IS 'User comments on recipes';

COMMENT ON COLUMN ingredients.normalized_name IS 'Automatically generated normalized version of ingredient name for consistent searching';
COMMENT ON COLUMN recipe_ingredients.quantity_text IS 'Original quantity string as entered by user (e.g., "2 cups", "1 large onion")';
COMMENT ON COLUMN recipe_ingredients.sort_order IS 'Order of ingredients as they appear in the recipe';

