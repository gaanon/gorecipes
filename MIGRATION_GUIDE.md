# Migration Guide: BadgerDB to PostgreSQL

This guide walks you through migrating your GoRecipes application from BadgerDB to PostgreSQL.

## Prerequisites

1. **Backup your data**: Always backup your BadgerDB data before starting migration
2. **PostgreSQL installed**: Ensure PostgreSQL is running and accessible
3. **Go 1.24+**: Required for building the migration tools

## Migration Steps

### Step 1: Prepare PostgreSQL Database

1. Create a new PostgreSQL database:
```sql
CREATE DATABASE gorecipes;
CREATE USER gorecipes WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE gorecipes TO gorecipes;
```

2. Run the schema migration:
```bash
psql -d gorecipes -f backend/internal/database/migrations/001_initial_schema.sql
```

### Step 2: Build Migration Tools

```bash
cd backend
./scripts/migrate.sh build
```

### Step 3: Export Existing Data (Recommended)

First, perform a dry run to export your data without touching PostgreSQL:

```bash
./scripts/migrate.sh export --badger-path ./path/to/your/badgerdb
```

This creates JSON files in `./migration_export/` that you can inspect:
- `recipes.json` - All your recipes
- `ingredients.json` - Extracted and normalized ingredients
- `recipe_ingredients.json` - Recipe-ingredient relationships
- `meal_plan_entries.json` - Meal planning data
- `migration_summary.json` - Summary statistics

### Step 4: Dry Run Migration

Test the migration without writing to PostgreSQL:

```bash
./scripts/migrate.sh dry-run \
  --badger-path ./path/to/your/badgerdb \
  --postgres-url "postgres://gorecipes:password@localhost:5432/gorecipes?sslmode=disable"
```

### Step 5: Perform Full Migration

⚠️ **Warning**: This will write data to PostgreSQL

```bash
./scripts/migrate.sh migrate \
  --badger-path ./path/to/your/badgerdb \
  --postgres-url "postgres://gorecipes:password@localhost:5432/gorecipes?sslmode=disable"
```

### Step 6: Verify Migration

Check that your data migrated correctly:

```bash
./scripts/migrate.sh verify \
  --postgres-url "postgres://gorecipes:password@localhost:5432/gorecipes?sslmode=disable"
```

## What Gets Migrated

### Recipes
- ✅ Recipe ID, name, method, photo filename
- ✅ Created and updated timestamps
- ✅ All existing recipes preserved

### Ingredients
- ✅ Extracted from ingredient strings in recipes
- ✅ Automatically normalized for better searching
- ✅ Deduplicated ("tomato" and "tomatoes" become one ingredient)
- ✅ Original quantity text preserved in recipe-ingredient relationships

### Recipe-Ingredient Relationships
- ✅ Links recipes to their ingredients
- ✅ Preserves original quantity text ("2 cups flour")
- ✅ Maintains ingredient order in recipes

### Meal Plan Entries
- ✅ All meal planning data
- ✅ Recipe-date associations
- ✅ Created timestamps

## Data Transformation

### Ingredient Processing

The migration tool intelligently processes ingredient strings:

**Before (BadgerDB)**:
```json
{
  "ingredients": [
    "2 cups all-purpose flour",
    "3 large eggs, beaten",
    "1 cup whole milk"
  ]
}
```

**After (PostgreSQL)**:
```sql
-- ingredients table
INSERT INTO ingredients (name, normalized_name) VALUES
  ('flour', 'flour'),
  ('eggs', 'eggs'),
  ('milk', 'milk');

-- recipe_ingredients table
INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity_text, sort_order) VALUES
  ('recipe-uuid', 'flour-uuid', '2 cups all-purpose flour', 1),
  ('recipe-uuid', 'eggs-uuid', '3 large eggs, beaten', 2),
  ('recipe-uuid', 'milk-uuid', '1 cup whole milk', 3);
```

### Removed Fields

- `FilterableIngredientNames` - No longer needed with proper SQL queries
- Ingredient keys in BadgerDB - Replaced with proper foreign key relationships

## Troubleshooting

### Common Issues

1. **Connection Errors**
   ```
   Failed to connect to PostgreSQL
   ```
   - Check PostgreSQL is running
   - Verify connection URL format
   - Check user permissions

2. **Permission Errors**
   ```
   permission denied for table recipes
   ```
   - Ensure the database user has proper permissions
   - Run: `GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gorecipes;`

3. **Missing Dependencies**
   ```
   cannot find package
   ```
   - Run: `go mod tidy` in the backend directory

### Recovery

If migration fails:

1. **Check export files**: Your data is safely exported in JSON format
2. **Clear PostgreSQL**: Drop and recreate tables if needed
3. **Fix issues**: Address any errors shown in migration output
4. **Retry**: Run migration again

### Rollback to BadgerDB

If you need to rollback:

1. Keep your original BadgerDB data intact
2. Switch back to the old codebase
3. Your BadgerDB data remains unchanged

## Performance Notes

- **Small datasets** (< 1000 recipes): Migration completes in seconds
- **Large datasets** (> 10000 recipes): May take several minutes
- **Memory usage**: Minimal - data is processed in batches
- **Disk space**: Requires space for export files (temporary)

## Post-Migration

### Update Docker Configuration

After successful migration, update your `docker-compose.yml` to use PostgreSQL (covered in Step 4 of the main migration plan).

### Update Application Code

The backend code needs to be updated to use PostgreSQL instead of BadgerDB (covered in Step 3 of the main migration plan).

### Cleanup

1. Remove BadgerDB data directory (after confirming migration success)
2. Remove migration export files (optional)
3. Update backup procedures to use PostgreSQL

## Support

If you encounter issues:

1. Check the migration logs for specific error messages
2. Verify your PostgreSQL connection and permissions
3. Ensure all prerequisites are met
4. Check the export files to confirm data was read correctly from BadgerDB

## Next Steps

After successful migration:

1. Update the backend code to use PostgreSQL
2. Modify Docker configuration
3. Test the application thoroughly
4. Set up PostgreSQL backups
5. Monitor performance and optimize as needed

