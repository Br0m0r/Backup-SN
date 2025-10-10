#!/bin/bash

# Run migrations for SQLite database
# This script applies all pending migrations in order

DB_PATH="${DATABASE_PATH:-/app/social_network.db}"
MIGRATIONS_DIR="/app/migrations"

echo "Starting database migrations..."
echo "Database: $DB_PATH"
echo "Migrations directory: $MIGRATIONS_DIR"

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Database not found, creating new database..."
    touch "$DB_PATH"
fi

# Create schema_migrations table if it doesn't exist
sqlite3 "$DB_PATH" <<EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    dirty INTEGER NOT NULL DEFAULT 0
);
EOF

# Get current migration version
CURRENT_VERSION=$(sqlite3 "$DB_PATH" "SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = 0")
echo "Current migration version: $CURRENT_VERSION"

# Apply pending migrations
for file in $(ls $MIGRATIONS_DIR/*_*.up.sql | sort); do
    # Extract version number from filename (e.g., 000001 from 000001_Users.up.sql)
    VERSION=$(basename "$file" | cut -d'_' -f1 | sed 's/^0*//')
    
    if [ -z "$VERSION" ]; then
        VERSION=0
    fi
    
    if [ "$VERSION" -gt "$CURRENT_VERSION" ]; then
        echo "Applying migration $VERSION: $(basename $file)"
        
        # Mark as dirty in case of failure
        sqlite3 "$DB_PATH" "INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES ($VERSION, 1)"
        
        # Run migration
        if sqlite3 "$DB_PATH" < "$file"; then
            # Mark as clean (successful)
            sqlite3 "$DB_PATH" "UPDATE schema_migrations SET dirty = 0 WHERE version = $VERSION"
            echo "✓ Migration $VERSION applied successfully"
        else
            echo "✗ Migration $VERSION failed!"
            exit 1
        fi
    fi
done

echo "All migrations completed successfully!"
