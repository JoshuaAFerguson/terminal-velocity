#!/bin/bash
# File: scripts/restore.sh
# Project: Terminal Velocity
# Description: Database restore from backup
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
BACKUP_DIR=${BACKUP_DIR:-backups}

# Usage information
usage() {
    echo "Usage: $0 [options] <backup_file>"
    echo ""
    echo "Options:"
    echo "  -l, --list              List available backups"
    echo "  -f, --force             Skip confirmation prompt"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST                 Database host (default: localhost)"
    echo "  DB_PORT                 Database port (default: 5432)"
    echo "  DB_USER                 Database user (default: terminal_velocity)"
    echo "  DB_PASSWORD             Database password"
    echo "  DB_NAME                 Database name (default: terminal_velocity)"
    echo "  BACKUP_DIR              Backup directory (default: backups)"
    echo ""
    echo "Examples:"
    echo "  $0 --list               # List available backups"
    echo "  $0 backup.sql.gz        # Restore from specific backup"
    echo "  $0 -f backup.sql.gz     # Restore without confirmation"
    exit 1
}

# List available backups
list_backups() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}               AVAILABLE BACKUPS${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""

    if [ ! -d "$BACKUP_DIR" ]; then
        echo -e "${YELLOW}No backup directory found: ${BACKUP_DIR}${NC}"
        exit 0
    fi

    BACKUPS=$(find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql*" -type f -printf '%T@ %p\n' | sort -rn | cut -d' ' -f2-)

    if [ -z "$BACKUPS" ]; then
        echo -e "${YELLOW}No backups found in ${BACKUP_DIR}${NC}"
        exit 0
    fi

    echo "Backups (newest first):"
    echo ""
    printf "%-40s %-15s %-20s\n" "Filename" "Size" "Date"
    echo "─────────────────────────────────────────────────────────────────────────"

    while IFS= read -r backup; do
        filename=$(basename "$backup")
        size=$(du -h "$backup" | cut -f1)
        date=$(stat -c %y "$backup" | cut -d'.' -f1)
        printf "%-40s %-15s %-20s\n" "$filename" "$size" "$date"
    done <<< "$BACKUPS"

    echo ""
    echo "To restore a backup:"
    echo "  $0 ${BACKUP_DIR}/<filename>"
    echo ""
}

# Parse command line arguments
FORCE=false
LIST=false
BACKUP_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -l|--list)
            LIST=true
            shift
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            if [ -z "$BACKUP_FILE" ]; then
                BACKUP_FILE="$1"
                shift
            else
                echo -e "${RED}Error: Too many arguments${NC}"
                usage
            fi
            ;;
    esac
done

# Handle --list option
if [ "$LIST" = true ]; then
    list_backups
    exit 0
fi

# Check if backup file is provided
if [ -z "$BACKUP_FILE" ]; then
    echo -e "${RED}Error: No backup file specified${NC}"
    echo ""
    usage
fi

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo -e "${RED}Error: Backup file not found: ${BACKUP_FILE}${NC}"
    exit 1
fi

# Determine if file is compressed
IS_COMPRESSED=false
if [[ "$BACKUP_FILE" == *.gz ]]; then
    IS_COMPRESSED=true
    # Check if gunzip is available
    if ! command -v gunzip &> /dev/null; then
        echo -e "${RED}Error: gunzip command not found. Cannot decompress backup.${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}       TERMINAL VELOCITY - DATABASE RESTORE${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Restore Configuration:"
echo "  Database:     ${DB_NAME}@${DB_HOST}:${DB_PORT}"
echo "  Backup File:  ${BACKUP_FILE}"
echo "  Compressed:   ${IS_COMPRESSED}"
echo "  File Size:    $(du -h "${BACKUP_FILE}" | cut -f1)"
echo ""

# Confirmation prompt (unless --force is used)
if [ "$FORCE" != true ]; then
    echo -e "${RED}⚠️  WARNING: This will DROP and RECREATE the database!${NC}"
    echo -e "${RED}All current data will be PERMANENTLY LOST!${NC}"
    echo ""
    read -p "Are you sure you want to continue? Type 'yes' to proceed: " -r
    echo

    if [ "$REPLY" != "yes" ]; then
        echo "Restore cancelled."
        exit 0
    fi
fi

# Test database connection
echo -e "${YELLOW}Testing database connection...${NC}"
if ! PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "postgres" -c '\q' 2>/dev/null; then
    echo -e "${RED}✗ Failed to connect to database${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Database connection successful${NC}"
echo ""

# Drop and recreate database
echo -e "${YELLOW}Dropping existing database...${NC}"
PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "postgres" <<EOF
DROP DATABASE IF EXISTS ${DB_NAME};
CREATE DATABASE ${DB_NAME};
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database recreated${NC}"
else
    echo -e "${RED}✗ Failed to recreate database${NC}"
    exit 1
fi
echo ""

# Restore backup
echo -e "${YELLOW}Restoring backup...${NC}"
if [ "$IS_COMPRESSED" = true ]; then
    gunzip -c "${BACKUP_FILE}" | PGPASSWORD="${DB_PASSWORD}" psql \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        -q 2>&1 | grep -i "error" || true
else
    PGPASSWORD="${DB_PASSWORD}" psql \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        -f "${BACKUP_FILE}" \
        -q 2>&1 | grep -i "error" || true
fi

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo -e "${GREEN}✓ Backup restored successfully${NC}"
else
    echo -e "${RED}✗ Restore failed (check errors above)${NC}"
    exit 1
fi

# Verify restore
echo ""
echo -e "${YELLOW}Verifying restore...${NC}"
TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | tr -d ' ')
echo "  Tables found: ${TABLE_COUNT}"

if [ "$TABLE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Restore verified${NC}"
else
    echo -e "${RED}✗ Warning: No tables found in restored database${NC}"
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                  RESTORE COMPLETE!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Restore Summary:"
echo "  Database:     ${DB_NAME}"
echo "  Backup File:  $(basename ${BACKUP_FILE})"
echo "  Tables:       ${TABLE_COUNT}"
echo ""
echo -e "${GREEN}Database has been restored successfully!${NC}"
echo ""
