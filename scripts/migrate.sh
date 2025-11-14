#!/bin/bash
# File: scripts/migrate.sh
# Project: Terminal Velocity
# Description: Database migration runner
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2025-01-14

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-terminal_velocity}
DB_PASSWORD=${DB_PASSWORD:-changeme_in_production}
DB_NAME=${DB_NAME:-terminal_velocity}

MIGRATIONS_DIR="scripts/migrations"

# Usage information
usage() {
    echo "Usage: $0 [up|down|status|reset]"
    echo ""
    echo "Commands:"
    echo "  up      Apply all pending migrations"
    echo "  down    Rollback the last migration"
    echo "  status  Show migration status"
    echo "  reset   Reset all migrations (DANGEROUS)"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST     Database host (default: localhost)"
    echo "  DB_PORT     Database port (default: 5432)"
    echo "  DB_USER     Database user (default: terminal_velocity)"
    echo "  DB_PASSWORD Database password"
    echo "  DB_NAME     Database name (default: terminal_velocity)"
    exit 1
}

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Database connection function
psql_exec() {
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -c "$1"
}

psql_exec_file() {
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -f "$1"
}

# Create migrations table if it doesn't exist
init_migrations() {
    echo "Initializing migrations table..."
    if [ -f "${MIGRATIONS_DIR}/000_create_migrations_table.sql" ]; then
        psql_exec_file "${MIGRATIONS_DIR}/000_create_migrations_table.sql" > /dev/null 2>&1 || true
        echo -e "${GREEN}✓ Migrations table initialized${NC}"
    fi
}

# Get list of applied migrations
get_applied_migrations() {
    psql_exec "SELECT version FROM schema_migrations ORDER BY version;" | tr -d ' '
}

# Apply migration
apply_migration() {
    local file=$1
    local version=$(echo "$file" | grep -oP '^\d+')
    local name=$(basename "$file" .sql)

    echo -e "${BLUE}Applying migration $version: $name${NC}"

    if psql_exec_file "$file"; then
        psql_exec "INSERT INTO schema_migrations (version, name) VALUES ($version, '$name');" > /dev/null
        echo -e "${GREEN}✓ Migration $version applied${NC}"
        return 0
    else
        echo -e "${RED}✗ Migration $version failed${NC}"
        return 1
    fi
}

# Show migration status
show_status() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}               MIGRATION STATUS${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""

    init_migrations

    local applied=$(get_applied_migrations)

    for file in $(ls ${MIGRATIONS_DIR}/*.sql | sort); do
        local version=$(echo "$file" | grep -oP '\d+' | head -1)
        local name=$(basename "$file" .sql)

        # Skip the migrations table creation
        if [ "$version" = "000" ]; then
            continue
        fi

        if echo "$applied" | grep -q "^${version}$"; then
            echo -e "${GREEN}✓${NC} $version: $name"
        else
            echo -e "${YELLOW}○${NC} $version: $name (pending)"
        fi
    done
    echo ""
}

# Apply all pending migrations
migrate_up() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}               APPLYING MIGRATIONS${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""

    init_migrations

    local applied=$(get_applied_migrations)
    local count=0

    for file in $(ls ${MIGRATIONS_DIR}/*.sql | sort); do
        local version=$(echo "$file" | grep -oP '\d+' | head -1)

        # Skip the migrations table creation
        if [ "$version" = "000" ]; then
            continue
        fi

        # Skip already applied migrations
        if echo "$applied" | grep -q "^${version}$"; then
            continue
        fi

        if apply_migration "$file"; then
            count=$((count + 1))
        else
            echo -e "${RED}Migration failed. Stopping.${NC}"
            exit 1
        fi
    done

    echo ""
    if [ $count -eq 0 ]; then
        echo -e "${GREEN}All migrations are up to date${NC}"
    else
        echo -e "${GREEN}✓ Applied $count migration(s)${NC}"
    fi
}

# Reset all migrations (dangerous)
migrate_reset() {
    echo -e "${RED}⚠️  WARNING: This will reset ALL migrations!${NC}"
    echo -e "${RED}This action cannot be undone.${NC}"
    echo ""
    read -p "Are you sure? Type 'yes' to continue: " -r
    echo

    if [ "$REPLY" != "yes" ]; then
        echo "Aborted."
        exit 0
    fi

    echo "Dropping schema_migrations table..."
    psql_exec "DROP TABLE IF EXISTS schema_migrations CASCADE;" > /dev/null
    echo -e "${GREEN}✓ Migrations reset${NC}"
}

# Main command router
case "${1:-}" in
    up)
        migrate_up
        ;;
    status)
        show_status
        ;;
    reset)
        migrate_reset
        ;;
    *)
        usage
        ;;
esac
