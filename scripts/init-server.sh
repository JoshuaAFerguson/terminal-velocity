#!/bin/bash
# File: scripts/init-server.sh
# Project: Terminal Velocity
# Description: Server initialization script - sets up database and universe
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
POSTGRES_USER=${POSTGRES_USER:-postgres}

SYSTEMS=${SYSTEMS:-100}
SEED=${SEED:-0}

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}       TERMINAL VELOCITY - SERVER INITIALIZATION${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Check if genmap binary exists
if [ ! -f "./genmap" ]; then
    echo -e "${YELLOW}Building genmap tool...${NC}"
    make build-tools || {
        echo -e "${RED}Error: Failed to build genmap tool${NC}"
        exit 1
    }
    echo -e "${GREEN}✓ genmap tool built${NC}"
    echo ""
fi

# Step 1: Create database and user (if needed)
echo -e "${BLUE}Step 1: Database Setup${NC}"
echo "----------------------------------------"
echo "Connecting to PostgreSQL as ${POSTGRES_USER}..."

# Check if database exists
if PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -c '\q' 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Database '${DB_NAME}' already exists${NC}"
    read -p "Do you want to reinitialize it? This will DELETE ALL DATA! [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Dropping existing database..."
        PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${POSTGRES_USER}" -c "DROP DATABASE IF EXISTS ${DB_NAME};"
    else
        echo "Aborted."
        exit 0
    fi
fi

# Create database and user
echo "Creating database and user..."
PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${POSTGRES_USER}" <<EOF
CREATE DATABASE ${DB_NAME};
CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';
GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};
EOF

echo -e "${GREEN}✓ Database created${NC}"
echo ""

# Step 2: Initialize schema
echo -e "${BLUE}Step 2: Schema Initialization${NC}"
echo "----------------------------------------"
echo "Running schema.sql..."

if [ ! -f "scripts/schema.sql" ]; then
    echo -e "${RED}Error: scripts/schema.sql not found${NC}"
    exit 1
fi

PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -f scripts/schema.sql

echo -e "${GREEN}✓ Schema initialized${NC}"
echo ""

# Step 3: Generate and populate universe
echo -e "${BLUE}Step 3: Universe Generation${NC}"
echo "----------------------------------------"
echo "Generating ${SYSTEMS} star systems..."
echo ""

./genmap \
    -systems "${SYSTEMS}" \
    -seed "${SEED}" \
    -save \
    -db-host "${DB_HOST}" \
    -db-port "${DB_PORT}" \
    -db-user "${DB_USER}" \
    -db-password "${DB_PASSWORD}" \
    -db-name "${DB_NAME}"

echo ""
echo -e "${GREEN}✓ Universe populated${NC}"
echo ""

# Step 4: Summary and next steps
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                    SETUP COMPLETE!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Database Information:"
echo "  Host:     ${DB_HOST}:${DB_PORT}"
echo "  Database: ${DB_NAME}"
echo "  User:     ${DB_USER}"
echo ""
echo "Universe Information:"
echo "  Systems:  ${SYSTEMS}"
echo "  Seed:     ${SEED}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "  1. Create your first player account:"
echo "     ./accounts create <username> <email>"
echo ""
echo "  2. (Optional) Add SSH key for passwordless login:"
echo "     ./accounts add-key <username> ~/.ssh/id_rsa.pub"
echo ""
echo "  3. Start the server:"
echo "     make run"
echo "     or"
echo "     ./server -config configs/config.yaml"
echo ""
echo "  4. Connect to the game:"
echo "     ssh -p 2222 <username>@${DB_HOST}"
echo ""
echo -e "${GREEN}Happy exploring! 🚀${NC}"
