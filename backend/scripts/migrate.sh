#!/bin/bash

# Migration script for BadgerDB to PostgreSQL
# This script provides convenient commands for the migration process

set -e

# Default values
BADGER_PATH="/app/data/badgerdb"
POSTGRES_URL="postgres://gorecipes:password@localhost:5432/gorecipes?sslmode=disable"
EXPORT_DIR="./migration_export"
MIGRATE_BINARY="./cmd/migrate/migrate"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_usage() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  build         Build the migration binary"
    echo "  export        Export data from BadgerDB to JSON files"
    echo "  dry-run       Perform a dry run migration (export only)"
    echo "  migrate       Perform the full migration to PostgreSQL"
    echo "  verify        Verify the migration results"
    echo "  help          Show this help message"
    echo ""
    echo "Options:"
    echo "  --badger-path PATH     Path to BadgerDB database (default: $BADGER_PATH)"
    echo "  --postgres-url URL     PostgreSQL connection URL (default: $POSTGRES_URL)"
    echo "  --export-dir DIR       Directory for export files (default: $EXPORT_DIR)"
    echo ""
    echo "Examples:"
    echo "  $0 build"
    echo "  $0 export --badger-path ./data/badger"
    echo "  $0 migrate --postgres-url 'postgres://user:pass@localhost/db'"
    echo "  $0 dry-run"
}

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

build_migration_binary() {
    log_info "Building migration binary..."
    cd "$(dirname "$0")/.."
    go build -o "$MIGRATE_BINARY" ./cmd/migrate/main.go
    
    if [ $? -eq 0 ]; then
        log_success "Migration binary built successfully: $MIGRATE_BINARY"
    else
        log_error "Failed to build migration binary"
        exit 1
    fi
}

export_data() {
    log_info "Exporting data from BadgerDB..."
    
    if [ ! -f "$MIGRATE_BINARY" ]; then
        log_warning "Migration binary not found. Building it first..."
        build_migration_binary
    fi
    
    "$MIGRATE_BINARY" \
        --badger-path="$BADGER_PATH" \
        --export-only \
        --export-dir="$EXPORT_DIR"
    
    if [ $? -eq 0 ]; then
        log_success "Data exported successfully to: $EXPORT_DIR"
        log_info "Export files:"
        ls -la "$EXPORT_DIR"/
    else
        log_error "Export failed"
        exit 1
    fi
}

dry_run_migration() {
    log_info "Performing dry run migration..."
    
    if [ ! -f "$MIGRATE_BINARY" ]; then
        log_warning "Migration binary not found. Building it first..."
        build_migration_binary
    fi
    
    "$MIGRATE_BINARY" \
        --badger-path="$BADGER_PATH" \
        --postgres-url="$POSTGRES_URL" \
        --dry-run \
        --export-dir="$EXPORT_DIR"
    
    if [ $? -eq 0 ]; then
        log_success "Dry run completed successfully"
        log_info "Check export files in: $EXPORT_DIR"
    else
        log_error "Dry run failed"
        exit 1
    fi
}

run_migration() {
    log_info "Starting full migration to PostgreSQL..."
    log_warning "This will modify your PostgreSQL database!"
    
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Migration cancelled"
        exit 0
    fi
    
    if [ ! -f "$MIGRATE_BINARY" ]; then
        log_warning "Migration binary not found. Building it first..."
        build_migration_binary
    fi
    
    "$MIGRATE_BINARY" \
        --badger-path="$BADGER_PATH" \
        --postgres-url="$POSTGRES_URL" \
        --export-dir="$EXPORT_DIR"
    
    if [ $? -eq 0 ]; then
        log_success "Migration completed successfully!"
    else
        log_error "Migration failed"
        exit 1
    fi
}

verify_migration() {
    log_info "Verifying migration results..."
    
    # Check if PostgreSQL is accessible
    psql "$POSTGRES_URL" -c "SELECT 1;" > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        log_error "Cannot connect to PostgreSQL database"
        exit 1
    fi
    
    # Get counts from PostgreSQL
    log_info "Checking data in PostgreSQL..."
    
    recipes_count=$(psql "$POSTGRES_URL" -t -c "SELECT COUNT(*) FROM recipes;" | tr -d ' ')
    ingredients_count=$(psql "$POSTGRES_URL" -t -c "SELECT COUNT(*) FROM ingredients;" | tr -d ' ')
    recipe_ingredients_count=$(psql "$POSTGRES_URL" -t -c "SELECT COUNT(*) FROM recipe_ingredients;" | tr -d ' ')
    meal_plan_count=$(psql "$POSTGRES_URL" -t -c "SELECT COUNT(*) FROM meal_plan_entries;" | tr -d ' ')
    
    log_info "PostgreSQL Data Counts:"
    echo "  Recipes: $recipes_count"
    echo "  Ingredients: $ingredients_count"
    echo "  Recipe-Ingredient relationships: $recipe_ingredients_count"
    echo "  Meal plan entries: $meal_plan_count"
    
    # Check for any data integrity issues
    orphaned_recipe_ingredients=$(psql "$POSTGRES_URL" -t -c "
        SELECT COUNT(*) FROM recipe_ingredients ri 
        LEFT JOIN recipes r ON ri.recipe_id = r.id 
        LEFT JOIN ingredients i ON ri.ingredient_id = i.id 
        WHERE r.id IS NULL OR i.id IS NULL;
    " | tr -d ' ')
    
    orphaned_meal_plans=$(psql "$POSTGRES_URL" -t -c "
        SELECT COUNT(*) FROM meal_plan_entries mpe 
        LEFT JOIN recipes r ON mpe.recipe_id = r.id 
        WHERE r.id IS NULL;
    " | tr -d ' ')
    
    if [ "$orphaned_recipe_ingredients" -gt 0 ] || [ "$orphaned_meal_plans" -gt 0 ]; then
        log_warning "Data integrity issues found:"
        [ "$orphaned_recipe_ingredients" -gt 0 ] && echo "  Orphaned recipe-ingredients: $orphaned_recipe_ingredients"
        [ "$orphaned_meal_plans" -gt 0 ] && echo "  Orphaned meal plan entries: $orphaned_meal_plans"
    else
        log_success "No data integrity issues found"
    fi
    
    # Sample some data
    log_info "Sample migrated data:"
    psql "$POSTGRES_URL" -c "
        SELECT r.name, COUNT(ri.ingredient_id) as ingredient_count 
        FROM recipes r 
        LEFT JOIN recipe_ingredients ri ON r.id = ri.recipe_id 
        GROUP BY r.id, r.name 
        ORDER BY r.created_at DESC 
        LIMIT 5;
    "
}

# Parse command line arguments
COMMAND="$1"
shift || true

while [[ $# -gt 0 ]]; do
    case $1 in
        --badger-path)
            BADGER_PATH="$2"
            shift 2
            ;;
        --postgres-url)
            POSTGRES_URL="$2"
            shift 2
            ;;
        --export-dir)
            EXPORT_DIR="$2"
            shift 2
            ;;
        -*)
            log_error "Unknown option: $1"
            print_usage
            exit 1
            ;;
        *)
            log_error "Unknown argument: $1"
            print_usage
            exit 1
            ;;
    esac
done

# Execute command
case $COMMAND in
    build)
        build_migration_binary
        ;;
    export)
        export_data
        ;;
    dry-run)
        dry_run_migration
        ;;
    migrate)
        run_migration
        ;;
    verify)
        verify_migration
        ;;
    help|--help|-h|"")
        print_usage
        ;;
    *)
        log_error "Unknown command: $COMMAND"
        print_usage
        exit 1
        ;;
esac

