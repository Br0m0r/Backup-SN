#!/bin/bash

# Database Migration Script
# This script applies all pending migrations to bring the database to the latest version

set -e

# Configuration
DB_PATH="${DATABASE_PATH:-./social_network.db}"
MIGRATIONS_DIR="./db/migrations"

echo "=== Database Migration Script ==="
echo "Database: $DB_PATH"
echo "Migrations: $MIGRATIONS_DIR"
echo ""

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo "Error: 'migrate' command not found."
    echo "Please install golang-migrate:"
    echo "  macOS/Linux: brew install golang-migrate"
    echo "  or: go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    echo "  Windows: scoop install migrate"
    exit 1
fi

# Run migrations
echo "Applying migrations..."
migrate -path "$MIGRATIONS_DIR" -database "sqlite3://$DB_PATH" up

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Migrations completed successfully!"
    
    # Show current version
    VERSION=$(migrate -path "$MIGRATIONS_DIR" -database "sqlite3://$DB_PATH" version 2>&1 | tail -n 1)
    echo "Current database version: $VERSION"
else
    echo ""
    echo "✗ Migration failed!"
    exit 1
fi
