#!/bin/bash

# Test script for migration functionality
# This script sets up a test environment and validates the migration process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test configuration
TEST_DIR="./test_migration"
TEST_BADGER_PATH="$TEST_DIR/badger_test"
TEST_EXPORT_DIR="$TEST_DIR/export"
TEST_POSTGRES_URL="postgres://postgres:postgres@localhost:5432/gorecipes_test?sslmode=disable"
MIGRATE_BINARY="./cmd/migrate/migrate"

cleanup() {
    log_info "Cleaning up test environment..."
    rm -rf "$TEST_DIR"
    # Drop test database if it exists
    psql "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
        -c "DROP DATABASE IF EXISTS gorecipes_test;" 2>/dev/null || true
}

setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create test directory
    mkdir -p "$TEST_DIR"
    mkdir -p "$TEST_BADGER_PATH"
    mkdir -p "$TEST_EXPORT_DIR"
    
    # Create test database
    log_info "Creating test PostgreSQL database..."
    psql "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
        -c "DROP DATABASE IF EXISTS gorecipes_test;"
    psql "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
        -c "CREATE DATABASE gorecipes_test;"
    
    # Run schema migration
    log_info "Running schema migration..."
    psql "$TEST_POSTGRES_URL" -f "./internal/database/migrations/001_initial_schema.sql"
    
    log_success "Test environment ready"
}

create_test_data() {
    log_info "Creating test data in BadgerDB..."
    
    # We'll use the existing application to create some test data
    # This is a simplified approach - in a real scenario, you might have existing data
    
    cat > "$TEST_DIR/create_test_data.go" << 'EOF'
package main

import (
    "encoding/json"
    "log"
    "time"
    "gorecipes/backend/internal/database"
    "gorecipes/backend/internal/models"
    "github.com/dgraph-io/badger/v4"
    "github.com/google/uuid"
)

func main() {
    dbPath := "./test_migration/badger_test"
    
    if err := database.InitDB(dbPath); err != nil {
        log.Fatalf("Failed to init DB: %v", err)
    }
    defer database.CloseDB()
    
    // Create test recipes
    recipes := []models.Recipe{
        {
            ID:   uuid.New().String(),
            Name: "Test Pancakes",
            Ingredients: []string{
                "2 cups all-purpose flour",
                "2 large eggs",
                "1.5 cups whole milk",
                "3 tbsp melted butter",
                "2 tbsp granulated sugar",
                "1 tsp salt",
                "2 tsp baking powder",
            },
            Method: "Mix ingredients and cook on griddle.",
            PhotoFilename: "pancakes.jpg",
            CreatedAt: time.Now().UTC(),
            UpdatedAt: time.Now().UTC(),
        },
        {
            ID:   uuid.New().String(),
            Name: "Simple Pasta",
            Ingredients: []string{
                "1 lb pasta",
                "2 cups marinara sauce",
                "1/2 cup grated parmesan cheese",
                "2 cloves garlic, minced",
                "2 tbsp olive oil",
                "1 tsp dried basil",
            },
            Method: "Cook pasta, heat sauce with garlic, combine and serve.",
            PhotoFilename: "pasta.jpg",
            CreatedAt: time.Now().UTC().Add(-24 * time.Hour),
            UpdatedAt: time.Now().UTC().Add(-24 * time.Hour),
        },
    }
    
    // Save recipes to BadgerDB
    err := database.DB.Update(func(txn *badger.Txn) error {
        for _, recipe := range recipes {
            recipeJSON, err := json.Marshal(recipe)
            if err != nil {
                return err
            }
            key := []byte("recipe:" + recipe.ID)
            if err := txn.Set(key, recipeJSON); err != nil {
                return err
            }
            log.Printf("Created test recipe: %s", recipe.Name)
        }
        return nil
    })
    
    if err != nil {
        log.Fatalf("Failed to create test data: %v", err)
    }
    
    // Create test meal plan entries
    mealPlanEntries := []models.MealPlanEntry{
        {
            ID:        uuid.New().String(),
            RecipeID:  recipes[0].ID,
            Date:      time.Now().UTC().AddDate(0, 0, 1), // Tomorrow
            CreatedAt: time.Now().UTC(),
        },
        {
            ID:        uuid.New().String(),
            RecipeID:  recipes[1].ID,
            Date:      time.Now().UTC().AddDate(0, 0, 2), // Day after tomorrow
            CreatedAt: time.Now().UTC(),
        },
    }
    
    err = database.DB.Update(func(txn *badger.Txn) error {
        for _, entry := range mealPlanEntries {
            entryJSON, err := json.Marshal(entry)
            if err != nil {
                return err
            }
            key := []byte("mealplanentry:" + entry.ID)
            if err := txn.Set(key, entryJSON); err != nil {
                return err
            }
            log.Printf("Created test meal plan entry for date: %s", entry.Date.Format("2006-01-02"))
        }
        return nil
    })
    
    if err != nil {
        log.Fatalf("Failed to create test meal plan data: %v", err)
    }
    
    log.Printf("Test data created successfully")
}
EOF
    
    # Build and run test data creator
    cd "$(dirname "$0")/.."
    go run "$TEST_DIR/create_test_data.go"
    
    log_success "Test data created in BadgerDB"
}

test_migration() {
    log_info "Testing migration process..."
    
    # Build migration binary
    log_info "Building migration binary..."
    go build -o "$MIGRATE_BINARY" ./cmd/migrate/main.go
    
    # Test export
    log_info "Testing export functionality..."
    "$MIGRATE_BINARY" \
        --badger-path="$TEST_BADGER_PATH" \
        --export-only \
        --export-dir="$TEST_EXPORT_DIR"
    
    # Verify export files
    if [ ! -f "$TEST_EXPORT_DIR/recipes.json" ]; then
        log_error "Export failed - recipes.json not found"
        return 1
    fi
    
    # Test dry run
    log_info "Testing dry run migration..."
    "$MIGRATE_BINARY" \
        --badger-path="$TEST_BADGER_PATH" \
        --postgres-url="$TEST_POSTGRES_URL" \
        --dry-run \
        --export-dir="$TEST_EXPORT_DIR"
    
    # Test actual migration
    log_info "Testing actual migration..."
    "$MIGRATE_BINARY" \
        --badger-path="$TEST_BADGER_PATH" \
        --postgres-url="$TEST_POSTGRES_URL" \
        --export-dir="$TEST_EXPORT_DIR"
    
    log_success "Migration completed"
}

verify_migration() {
    log_info "Verifying migration results..."
    
    # Check PostgreSQL data
    recipes_count=$(psql "$TEST_POSTGRES_URL" -t -c "SELECT COUNT(*) FROM recipes;" | tr -d ' ')
    ingredients_count=$(psql "$TEST_POSTGRES_URL" -t -c "SELECT COUNT(*) FROM ingredients;" | tr -d ' ')
    recipe_ingredients_count=$(psql "$TEST_POSTGRES_URL" -t -c "SELECT COUNT(*) FROM recipe_ingredients;" | tr -d ' ')
    meal_plan_count=$(psql "$TEST_POSTGRES_URL" -t -c "SELECT COUNT(*) FROM meal_plan_entries;" | tr -d ' ')
    
    log_info "Migration Results:"
    echo "  Recipes: $recipes_count (expected: 2)"
    echo "  Ingredients: $ingredients_count (expected: ~12)"
    echo "  Recipe-Ingredient relationships: $recipe_ingredients_count (expected: 13)"
    echo "  Meal plan entries: $meal_plan_count (expected: 2)"
    
    # Verify data integrity
    orphaned_count=$(psql "$TEST_POSTGRES_URL" -t -c "
        SELECT COUNT(*) FROM recipe_ingredients ri 
        LEFT JOIN recipes r ON ri.recipe_id = r.id 
        LEFT JOIN ingredients i ON ri.ingredient_id = i.id 
        WHERE r.id IS NULL OR i.id IS NULL;
    " | tr -d ' ')
    
    if [ "$orphaned_count" -gt 0 ]; then
        log_error "Data integrity issue: $orphaned_count orphaned recipe-ingredients"
        return 1
    fi
    
    # Check that ingredients were properly normalized
    sample_ingredients=$(psql "$TEST_POSTGRES_URL" -t -c "
        SELECT name FROM ingredients ORDER BY name LIMIT 5;
    ")
    
    log_info "Sample ingredients: $sample_ingredients"
    
    # Verify expected counts
    if [ "$recipes_count" -ne 2 ]; then
        log_error "Expected 2 recipes, got $recipes_count"
        return 1
    fi
    
    if [ "$meal_plan_count" -ne 2 ]; then
        log_error "Expected 2 meal plan entries, got $meal_plan_count"
        return 1
    fi
    
    log_success "Migration verification passed"
}

run_tests() {
    log_info "Starting migration tests..."
    
    # Check prerequisites
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL client (psql) not found"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        log_error "Go not found"
        exit 1
    fi
    
    # Test PostgreSQL connection
    if ! psql "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -c "SELECT 1;" > /dev/null 2>&1; then
        log_error "Cannot connect to PostgreSQL. Please ensure PostgreSQL is running and accessible with default credentials."
        exit 1
    fi
    
    setup_test_environment
    create_test_data
    test_migration
    verify_migration
    
    log_success "All migration tests passed!"
}

# Handle script arguments
case "${1:-test}" in
    test)
        run_tests
        ;;
    clean)
        cleanup
        log_success "Test environment cleaned up"
        ;;
    setup)
        setup_test_environment
        ;;
    *)
        echo "Usage: $0 [test|clean|setup]"
        echo "  test  - Run full migration test (default)"
        echo "  clean - Clean up test environment"
        echo "  setup - Set up test environment only"
        exit 1
        ;;
esac

# Always cleanup on exit
trap cleanup EXIT

