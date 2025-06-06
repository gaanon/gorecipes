# PostgreSQL Migration Plan for GoRecipes

This document outlines the plan to migrate the GoRecipes application from BadgerDB (NoSQL) to PostgreSQL (SQL).

## Overview

The current application uses BadgerDB as a key-value store with denormalized data structures. By migrating to PostgreSQL, we can leverage proper relational data modeling, efficient querying, and advanced search capabilities while eliminating the need for workarounds like `FilterableIngredientNames`.

## Migration Steps

### 1. Database Schema Design

Design a PostgreSQL schema with the following tables:

- **`recipes`** - Core recipe data
  - `id` (UUID, Primary Key)
  - `name` (VARCHAR, NOT NULL)
  - `method` (TEXT, NOT NULL)
  - `photo_filename` (VARCHAR, NULLABLE)
  - `created_at` (TIMESTAMP WITH TIME ZONE, NOT NULL)
  - `updated_at` (TIMESTAMP WITH TIME ZONE, NOT NULL)

- **`ingredients`** - Master list of ingredients
  - `id` (UUID, Primary Key)
  - `name` (VARCHAR, NOT NULL, UNIQUE)
  - `normalized_name` (VARCHAR, NOT NULL) - for efficient searching
  - `created_at` (TIMESTAMP WITH TIME ZONE, NOT NULL)

- **`recipe_ingredients`** - Junction table linking recipes to ingredients
  - `id` (UUID, Primary Key)
  - `recipe_id` (UUID, Foreign Key to recipes.id)
  - `ingredient_id` (UUID, Foreign Key to ingredients.id)
  - `quantity_text` (VARCHAR) - original quantity string (e.g., "2 cups")
  - `sort_order` (INTEGER) - to maintain ingredient order in recipe

- **`meal_plan_entries`** - Meal planning data
  - `id` (UUID, Primary Key)
  - `recipe_id` (UUID, Foreign Key to recipes.id)
  - `date` (DATE, NOT NULL)
  - `created_at` (TIMESTAMP WITH TIME ZONE, NOT NULL)

**PostgreSQL Features to Utilize:**
- UUID primary keys for distributed-friendly IDs
- Full-text search indexes on ingredient names
- GIN indexes for efficient ingredient filtering
- Foreign key constraints for data integrity
- Partial indexes for performance optimization

### 2. Data Migration Script

- Export existing data from BadgerDB
- Parse ingredient strings to extract individual ingredients and normalize them
- Create a migration script that:
  - Populates the `ingredients` table with unique ingredients
  - Links recipes to ingredients through `recipe_ingredients` table
  - Migrates meal plan entries
- Handle duplicate ingredient names and normalize variations
- Preserve original quantity text while extracting searchable ingredient names

### 3. Update Backend Code

- Replace BadgerDB with PostgreSQL using `pgx` or `gorm`
- Rewrite data access layer:
  - Recipe CRUD operations with proper SQL joins
  - Ingredient search using PostgreSQL full-text search or LIKE queries
  - Efficient filtering by ingredients using EXISTS or JOIN queries
- Remove the `FilterableIngredientNames` field from the Recipe model
- Update ingredient autocomplete to query the `ingredients` table
- Implement proper transaction handling for data consistency

### 4. Modify Docker Configuration

- Add PostgreSQL service to `docker-compose.yml`
- Include environment variables for database connection
- Add volume for PostgreSQL data persistence
- Update backend service dependencies
- Ensure proper initialization order (PostgreSQL before backend)

### 5. Environment Handling

- Database connection configuration
- Migration script execution on startup
- Secure credential management using environment variables
- Support for different environments (development, production)

### 6. Enhanced Query Capabilities

Leverage PostgreSQL's superior querying for:
- Complex ingredient filtering (AND/OR operations)
- Fuzzy ingredient matching using similarity functions
- Recipe recommendations based on available ingredients
- Nutritional data aggregation (future enhancement)
- Advanced search with ranking and relevance scoring

### 7. Testing and Validation

- Unit tests for new data access layer
- Integration tests for migration scripts
- Performance testing for ingredient search queries
- Data integrity validation
- End-to-end testing of the complete application

### 8. Documentation and Deployment

- Updated setup instructions
- Migration guide for existing users
- Rollback procedures
- Performance tuning guidelines
- Backup and restore procedures

## Benefits of PostgreSQL Migration

1. **Better Data Modeling**: Proper relational structure with normalized data
2. **Efficient Querying**: Advanced SQL capabilities for complex queries
3. **Full-Text Search**: Built-in search capabilities without external dependencies
4. **Data Integrity**: Foreign key constraints and ACID transactions
5. **Scalability**: Better performance for complex queries and larger datasets
6. **Maintenance**: Standard SQL tooling and administration
7. **Future-Proof**: Easier to add new features like nutritional data, user management, etc.

## Implementation Order

1. Create database schema and migration scripts
2. Update Go models and data access layer
3. Modify Docker configuration
4. Test migration with existing data
5. Update frontend if needed for any API changes
6. Deploy and validate

## Rollback Strategy

In case of issues:
1. Keep BadgerDB data intact during migration
2. Maintain ability to switch back to BadgerDB code
3. Document exact steps to revert changes
4. Test rollback procedure in development environment

---

**Status**: Planning Phase
**Next Step**: Implement database schema design

