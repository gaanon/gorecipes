# Database Schema Documentation

This directory contains the PostgreSQL database schema and related files for the GoRecipes application.

## Files Overview

- `schema.sql` - Complete database schema with all tables, indexes, functions, and views
- `migrations/001_initial_schema.sql` - Initial migration to create the schema
- `migrations/001_initial_schema_down.sql` - Rollback migration
- `queries.sql` - Common SQL queries that will be used in the Go application

## Database Design

### Tables

#### `recipes`
Stores core recipe information:
- `id` (UUID) - Primary key
- `name` (VARCHAR) - Recipe name
- `method` (TEXT) - Cooking instructions
- `photo_filename` (VARCHAR) - Optional photo file
- `created_at`, `updated_at` (TIMESTAMP) - Audit fields

#### `ingredients`
Master list of all ingredients:
- `id` (UUID) - Primary key
- `name` (VARCHAR) - Original ingredient name (unique)
- `normalized_name` (VARCHAR) - Automatically normalized for searching
- `created_at` (TIMESTAMP) - When ingredient was first added

#### `recipe_ingredients`
Junction table linking recipes to ingredients:
- `id` (UUID) - Primary key
- `recipe_id` (UUID) - Foreign key to recipes
- `ingredient_id` (UUID) - Foreign key to ingredients
- `quantity_text` (VARCHAR) - Original quantity string ("2 cups", "1 large")
- `sort_order` (INTEGER) - Order ingredients appear in recipe

#### `meal_plan_entries`
Meal planning data:
- `id` (UUID) - Primary key
- `recipe_id` (UUID) - Foreign key to recipes
- `date` (DATE) - Planned cooking date
- `created_at` (TIMESTAMP) - When plan was created

### Key Features

#### Automatic Normalization
The `normalize_ingredient_name()` function automatically:
- Converts to lowercase
- Removes common descriptors (fresh, dried, chopped, etc.)
- Removes units and quantities
- Standardizes whitespace

#### Full-Text Search
GIN indexes enable efficient full-text search on:
- Recipe names
- Ingredient names
- Both original and normalized ingredient names

#### Performance Indexes
- Recipe lookups by date
- Ingredient searches
- Recipe-ingredient joins
- Meal plan date ranges

### Views

#### `recipes_with_stats`
Combines recipes with ingredient counts for API responses.

#### `ingredient_usage_stats`
Shows how frequently each ingredient is used across recipes.

## Migration Strategy

The migration from BadgerDB will:
1. Export all existing recipes and meal plans
2. Parse ingredient strings to extract individual ingredients
3. Populate the normalized relational structure
4. Maintain data integrity through foreign key constraints

## Usage Examples

See `queries.sql` for common query patterns that will be implemented in the Go application:
- Recipe search by ingredients
- Ingredient autocomplete
- Meal plan queries
- Recipe recommendations based on available ingredients

## Performance Considerations

- GIN indexes for full-text search
- B-tree indexes for common lookup patterns
- Foreign key constraints for data integrity
- Automatic timestamp updates via triggers
- Efficient pagination support

## Development Setup

1. Run the initial migration: `psql -f migrations/001_initial_schema.sql`
3. Test queries: Use examples from `queries.sql`

## Rollback

To rollback the schema:
```sql
psql -f migrations/001_initial_schema_down.sql
```

