#!/bin/bash
# File: scripts/backup.sh
# Project: Terminal Velocity
# Description: Automated database backup with retention policies
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

# Backup configuration
BACKUP_DIR=${BACKUP_DIR:-backups}
RETENTION_DAYS=${RETENTION_DAYS:-7}
RETENTION_COUNT=${RETENTION_COUNT:-10}
COMPRESS=${COMPRESS:-true}

# Usage information
usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -d, --dir DIR           Backup directory (default: backups)"
    echo "  -r, --retention DAYS    Keep backups for N days (default: 7)"
    echo "  -c, --count COUNT       Keep last N backups (default: 10)"
    echo "  -n, --no-compress       Don't compress backup files"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST                 Database host (default: localhost)"
    echo "  DB_PORT                 Database port (default: 5432)"
    echo "  DB_USER                 Database user (default: terminal_velocity)"
    echo "  DB_PASSWORD             Database password"
    echo "  DB_NAME                 Database name (default: terminal_velocity)"
    echo ""
    echo "Examples:"
    echo "  $0                      # Create backup with defaults"
    echo "  $0 -d /var/backups      # Custom backup directory"
    echo "  $0 -r 30 -c 50          # Keep 30 days and 50 backups"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        -r|--retention)
            RETENTION_DAYS="$2"
            shift 2
            ;;
        -c|--count)
            RETENTION_COUNT="$2"
            shift 2
            ;;
        -n|--no-compress)
            COMPRESS=false
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

# Check if pg_dump is available
if ! command -v pg_dump &> /dev/null; then
    echo -e "${RED}Error: pg_dump command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql"

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}       TERMINAL VELOCITY - DATABASE BACKUP${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Backup Configuration:"
echo "  Database:    ${DB_NAME}@${DB_HOST}:${DB_PORT}"
echo "  Backup Dir:  ${BACKUP_DIR}"
echo "  Filename:    $(basename $BACKUP_FILE)"
echo "  Compress:    ${COMPRESS}"
echo "  Retention:   ${RETENTION_DAYS} days / ${RETENTION_COUNT} backups"
echo ""

# Create backup
echo -e "${YELLOW}Creating backup...${NC}"
PGPASSWORD="${DB_PASSWORD}" pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --no-owner \
    --no-privileges \
    --format=plain \
    --file="${BACKUP_FILE}" \
    --verbose 2>&1 | grep -v "NOTICE:" | tail -5

if [ $? -eq 0 ]; then
    BACKUP_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo -e "${GREEN}✓ Backup created: ${BACKUP_FILE} (${BACKUP_SIZE})${NC}"
else
    echo -e "${RED}✗ Backup failed${NC}"
    exit 1
fi

# Compress backup if enabled
if [ "$COMPRESS" = true ]; then
    echo ""
    echo -e "${YELLOW}Compressing backup...${NC}"
    gzip -f "${BACKUP_FILE}"
    BACKUP_FILE="${BACKUP_FILE}.gz"
    COMPRESSED_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo -e "${GREEN}✓ Backup compressed: ${BACKUP_FILE} (${COMPRESSED_SIZE})${NC}"
fi

# Apply retention policies
echo ""
echo -e "${YELLOW}Applying retention policies...${NC}"

# Count current backups
TOTAL_BACKUPS=$(find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql*" -type f | wc -l)
echo "  Current backups: ${TOTAL_BACKUPS}"

# Remove backups older than RETENTION_DAYS
DELETED_AGE=0
while IFS= read -r old_file; do
    if [ -n "$old_file" ]; then
        echo "  Deleting old backup: $(basename "$old_file")"
        rm -f "$old_file"
        DELETED_AGE=$((DELETED_AGE + 1))
    fi
done < <(find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql*" -type f -mtime +${RETENTION_DAYS})

if [ $DELETED_AGE -gt 0 ]; then
    echo -e "${GREEN}✓ Deleted ${DELETED_AGE} backup(s) older than ${RETENTION_DAYS} days${NC}"
fi

# Keep only last RETENTION_COUNT backups
CURRENT_COUNT=$(find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql*" -type f | wc -l)
if [ $CURRENT_COUNT -gt $RETENTION_COUNT ]; then
    DELETED_COUNT=0
    # Sort by modification time (oldest first) and delete excess
    while IFS= read -r old_file; do
        if [ -n "$old_file" ]; then
            echo "  Deleting excess backup: $(basename "$old_file")"
            rm -f "$old_file"
            DELETED_COUNT=$((DELETED_COUNT + 1))
        fi
    done < <(find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql*" -type f -printf '%T+ %p\n' | sort | head -n -${RETENTION_COUNT} | cut -d' ' -f2-)

    if [ $DELETED_COUNT -gt 0 ]; then
        echo -e "${GREEN}✓ Deleted ${DELETED_COUNT} excess backup(s) (keeping last ${RETENTION_COUNT})${NC}"
    fi
fi

# Show final backup count
FINAL_COUNT=$(find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql*" -type f | wc -l)
TOTAL_SIZE=$(du -sh "${BACKUP_DIR}" | cut -f1)

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                  BACKUP COMPLETE!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Backup Summary:"
echo "  Latest Backup:  $(basename ${BACKUP_FILE})"
echo "  Total Backups:  ${FINAL_COUNT}"
echo "  Total Size:     ${TOTAL_SIZE}"
echo ""
echo -e "${YELLOW}To restore this backup:${NC}"
if [ "$COMPRESS" = true ]; then
    echo "  gunzip -c ${BACKUP_FILE} | psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME}"
else
    echo "  psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -f ${BACKUP_FILE}"
fi
echo ""
