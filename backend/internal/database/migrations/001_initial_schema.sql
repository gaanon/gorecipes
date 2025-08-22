-- Migration: 001_initial_schema
-- Description: Create initial PostgreSQL schema for GoRecipes
-- Up Migration

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create recipes table
CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    method TEXT NOT NULL,
    photo_filename VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create ingredients table
CREATE TABLE ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    normalized_name VARCHAR(255) NOT NULL,
    normalized_name_tsvector TSVECTOR,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create recipe_ingredients junction table
CREATE TABLE recipe_ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity_text VARCHAR(500),
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(recipe_id, ingredient_id)
);

-- Create meal_plan_entries table
CREATE TABLE meal_plan_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(recipe_id, date)
);

-- Create performance indexes
CREATE INDEX idx_recipes_name ON recipes USING GIN (to_tsvector('english', name));
CREATE INDEX idx_recipes_created_at ON recipes(created_at DESC);
CREATE INDEX idx_recipes_updated_at ON recipes(updated_at DESC);

CREATE INDEX idx_ingredients_name ON ingredients(name);
CREATE INDEX idx_ingredients_normalized_name ON ingredients(normalized_name);
CREATE INDEX idx_ingredients_name_gin ON ingredients USING GIN (to_tsvector('english', name));
CREATE INDEX idx_ingredients_normalized_name_tsvector ON ingredients USING GIN(normalized_name_tsvector);

CREATE INDEX idx_recipe_ingredients_recipe_id ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredients_ingredient_id ON recipe_ingredients(ingredient_id);
CREATE INDEX idx_recipe_ingredients_sort_order ON recipe_ingredients(recipe_id, sort_order);

CREATE INDEX idx_meal_plan_entries_date ON meal_plan_entries(date DESC);
CREATE INDEX idx_meal_plan_entries_recipe_id ON meal_plan_entries(recipe_id);

-- Create utility functions
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION normalize_ingredient_name(input_name TEXT)
RETURNS TEXT AS $$
BEGIN
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

CREATE OR REPLACE FUNCTION set_normalized_ingredient_name()
RETURNS TRIGGER AS $$
BEGIN
    NEW.normalized_name = normalize_ingredient_name(NEW.name);
    NEW.normalized_name_tsvector = to_tsvector('english', NEW.normalized_name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
CREATE TRIGGER update_recipes_updated_at BEFORE UPDATE ON recipes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ingredients_updated_at BEFORE UPDATE ON ingredients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_set_normalized_ingredient_name
    BEFORE INSERT OR UPDATE ON ingredients
    FOR EACH ROW
    EXECUTE FUNCTION set_normalized_ingredient_name();

-- Create views
CREATE VIEW recipes_with_stats AS
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

-- Add comments
COMMENT ON TABLE recipes IS 'Core recipe information including name, cooking method, and photo';
COMMENT ON TABLE ingredients IS 'Master list of all ingredients used across recipes';
COMMENT ON TABLE recipe_ingredients IS 'Junction table linking recipes to their ingredients with quantities';
COMMENT ON TABLE meal_plan_entries IS 'Meal planning entries associating recipes with specific dates';

