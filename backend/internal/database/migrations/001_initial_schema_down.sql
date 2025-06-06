-- Migration: 001_initial_schema (DOWN)
-- Description: Rollback initial PostgreSQL schema for GoRecipes
-- Down Migration

-- Drop views
DROP VIEW IF EXISTS recipes_with_stats;
DROP VIEW IF EXISTS ingredient_usage_stats;

-- Drop triggers
DROP TRIGGER IF EXISTS update_recipes_updated_at ON recipes;
DROP TRIGGER IF EXISTS trigger_set_normalized_ingredient_name ON ingredients;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS normalize_ingredient_name(TEXT);
DROP FUNCTION IF EXISTS set_normalized_ingredient_name();

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS meal_plan_entries;
DROP TABLE IF EXISTS recipe_ingredients;
DROP TABLE IF EXISTS ingredients;
DROP TABLE IF EXISTS recipes;

-- Drop extension (only if no other tables use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";

