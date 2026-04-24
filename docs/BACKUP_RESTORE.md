# Database Backup and Restore System

**Feature**: Automated Database Backup and Restore
**Phase**: 20
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Backup and Restore system provides automated database backup capabilities with retention policies, compression, and safe restore functionality. The system ensures data safety through regular backups and provides simple recovery mechanisms.

### Key Features

- **Automated Backups**: Scheduled database dumps via cron
- **Compression**: gzip compression for space efficiency
- **Retention Policies**: Days and count-based retention
- **Safe Restore**: Confirmation prompts and validation
- **Progress Tracking**: Visual feedback for large operations
- **Multiple Formats**: Plain SQL and compressed formats
- **Backup Listing**: Easy viewing of available backups
- **Cleanup Automation**: Automatic removal of old backups

---

## Architecture

### Components

1. **Backup Script** (`scripts/backup.sh`)
   - Database dumping with pg_dump
   - Compression with gzip
   - Retention enforcement
   - Automatic cleanup

2. **Restore Script** (`scripts/restore.sh`)
   - Backup listing
   - Safe restore with confirmation
   - Database recreation
   - Verification checks

3. **Cron Integration** (`scripts/crontab.example`)
   - Scheduled backup execution
   - Log management
   - Error notifications

### Data Flow

**Backup Flow**:
```
Trigger Backup (cron/manual)
         ↓
Create Backup Directory
         ↓
Execute pg_dump
         ↓
Compress with gzip
         ↓
Apply Retention Policies
         ↓
Delete Old Backups
         ↓
Report Summary
```

**Restore Flow**:
```
Select Backup File
         ↓
Confirmation Prompt
         ↓
Test Database Connection
         ↓
Drop Existing Database
         ↓
Create Fresh Database
         ↓
Restore from Backup
         ↓
Verify Restoration
         ↓
Report Success
```

---

## Implementation Details

### Backup Script

**Script Location**: `/home/user/terminal-velocity/scripts/backup.sh`

**Usage**:
```bash
# Basic backup with defaults
./scripts/backup.sh

# Custom backup directory
./scripts/backup.sh -d /var/backups

# Custom retention: 30 days and 50 backups
./scripts/backup.sh -r 30 -c 50

# No compression
./scripts/backup.sh -n
```

**Configuration Options**:

| Option | Description | Default |
|--------|-------------|---------|
| `-d, --dir` | Backup directory | backups |
| `-r, --retention` | Keep backups for N days | 7 |
| `-c, --count` | Keep last N backups | 10 |
| `-n, --no-compress` | Disable compression | false |

**Environment Variables**:
```bash
DB_HOST=localhost         # Database host
DB_PORT=5432             # Database port
DB_USER=terminal_velocity # Database user
DB_PASSWORD=<password>   # Database password
DB_NAME=terminal_velocity # Database name
```

**Backup Process**:
```bash
#!/bin/bash

# Generate filename with timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql"

# Create backup using pg_dump
PGPASSWORD="${DB_PASSWORD}" pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --no-owner \
    --no-privileges \
    --format=plain \
    --file="${BACKUP_FILE}" \
    --verbose

# Compress backup
if [ "$COMPRESS" = true ]; then
    gzip -f "${BACKUP_FILE}"
    BACKUP_FILE="${BACKUP_FILE}.gz"
fi
```

**Retention Enforcement**:
```bash
# Remove backups older than RETENTION_DAYS
find "${BACKUP_DIR}" \
    -name "${DB_NAME}_*.sql*" \
    -type f \
    -mtime +${RETENTION_DAYS} \
    -delete

# Keep only last RETENTION_COUNT backups
find "${BACKUP_DIR}" \
    -name "${DB_NAME}_*.sql*" \
    -type f \
    -printf '%T+ %p\n' | \
    sort | \
    head -n -${RETENTION_COUNT} | \
    cut -d' ' -f2- | \
    xargs rm -f
```

### Restore Script

**Script Location**: `/home/user/terminal-velocity/scripts/restore.sh`

**Usage**:
```bash
# List available backups
./scripts/restore.sh --list

# Restore from specific backup
./scripts/restore.sh backups/terminal_velocity_20250115_143022.sql.gz

# Force restore without confirmation
./scripts/restore.sh -f backup.sql.gz
```

**Configuration Options**:

| Option | Description |
|--------|-------------|
| `-l, --list` | List available backups |
| `-f, --force` | Skip confirmation prompt |

**List Backups**:
```bash
function list_backups() {
    BACKUPS=$(find "${BACKUP_DIR}" \
        -name "${DB_NAME}_*.sql*" \
        -type f \
        -printf '%T@ %p\n' | \
        sort -rn | \
        cut -d' ' -f2-)

    echo "Backups (newest first):"
    printf "%-40s %-15s %-20s\n" "Filename" "Size" "Date"
    echo "────────────────────────────────────────────────────────"

    while IFS= read -r backup; do
        filename=$(basename "$backup")
        size=$(du -h "$backup" | cut -f1)
        date=$(stat -c %y "$backup" | cut -d'.' -f1)
        printf "%-40s %-15s %-20s\n" "$filename" "$size" "$date"
    done <<< "$BACKUPS"
}
```

**Restore Process**:
```bash
# Confirmation prompt (unless --force)
if [ "$FORCE" != true ]; then
    echo "⚠️  WARNING: This will DROP and RECREATE the database!"
    echo "All current data will be PERMANENTLY LOST!"
    read -p "Type 'yes' to proceed: " -r

    if [ "$REPLY" != "yes" ]; then
        echo "Restore cancelled."
        exit 0
    fi
fi

# Test database connection
PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "postgres" \
    -c '\q'

# Drop and recreate database
PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "postgres" <<EOF
DROP DATABASE IF EXISTS ${DB_NAME};
CREATE DATABASE ${DB_NAME};
EOF

# Restore backup
if [ "$IS_COMPRESSED" = true ]; then
    gunzip -c "${BACKUP_FILE}" | \
        PGPASSWORD="${DB_PASSWORD}" psql \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        -q
else
    PGPASSWORD="${DB_PASSWORD}" psql \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        -f "${BACKUP_FILE}" \
        -q
fi

# Verify restore
TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")

echo "Tables found: ${TABLE_COUNT}"
```

---

## Automation

### Cron Configuration

**Example Crontab** (`scripts/crontab.example`):
```bash
# Daily backups at 2 AM
0 2 * * * /path/to/terminal-velocity/scripts/backup.sh >> /var/log/terminal-velocity/backup.log 2>&1

# Weekly cleanup (Sunday at 3 AM)
0 3 * * 0 /path/to/terminal-velocity/scripts/backup.sh -r 30 -c 50 >> /var/log/terminal-velocity/backup.log 2>&1

# Monthly full backup (1st of month at 1 AM)
0 1 1 * * /path/to/terminal-velocity/scripts/backup.sh -d /var/backups/monthly >> /var/log/terminal-velocity/backup.log 2>&1
```

**Install Crontab**:
```bash
# Edit user crontab
crontab -e

# Add backup schedule
0 2 * * * /home/user/terminal-velocity/scripts/backup.sh

# Verify crontab
crontab -l
```

### Backup Rotation Strategy

**Recommended Strategy**:

1. **Daily Backups**: Keep for 7 days
   ```bash
   0 2 * * * ./backup.sh -r 7 -c 10
   ```

2. **Weekly Backups**: Keep for 30 days
   ```bash
   0 2 * * 0 ./backup.sh -d backups/weekly -r 30 -c 5
   ```

3. **Monthly Backups**: Keep for 365 days
   ```bash
   0 2 1 * * ./backup.sh -d backups/monthly -r 365 -c 12
   ```

---

## Backup Best Practices

### Backup Security

**1. Secure Credentials**:
```bash
# Use environment variables
export DB_PASSWORD='secure_password'

# Or use .pgpass file
echo "localhost:5432:terminal_velocity:terminal_velocity:password" > ~/.pgpass
chmod 600 ~/.pgpass
```

**2. Encrypt Backups** (optional):
```bash
# Encrypt after backup
gpg --encrypt --recipient admin@example.com backup.sql.gz

# Decrypt before restore
gpg --decrypt backup.sql.gz.gpg > backup.sql.gz
```

**3. Off-Site Storage**:
```bash
# Sync to remote server
rsync -avz backups/ remote:/backups/terminal-velocity/

# Or use cloud storage
aws s3 sync backups/ s3://terminal-velocity-backups/
```

### Backup Monitoring

**Check Backup Success**:
```bash
# Check last backup
ls -lht backups/ | head -n 2

# Verify backup size (should be > 100KB)
[ $(stat -c%s "backup.sql.gz") -gt 102400 ] && echo "OK" || echo "FAIL"

# Test restore in development
./restore.sh -f latest_backup.sql.gz
```

**Alert on Failures**:
```bash
# Add to backup script
if [ $? -ne 0 ]; then
    echo "Backup failed!" | mail -s "Backup Failure" admin@example.com
fi
```

---

## Troubleshooting

### Common Issues

**Problem**: pg_dump command not found
**Solution**:
```bash
# Install PostgreSQL client
sudo apt-get install postgresql-client
# Or on macOS
brew install postgresql
```

**Problem**: Permission denied
**Solution**:
```bash
# Make scripts executable
chmod +x scripts/backup.sh scripts/restore.sh

# Check directory permissions
chmod 755 backups/
```

**Problem**: Backup file empty or very small
**Solution**:
- Verify database connection
- Check PostgreSQL is running
- Review error messages in output
- Ensure user has proper database permissions

**Problem**: Restore fails with errors
**Solution**:
- Check PostgreSQL version compatibility
- Ensure database exists
- Verify backup file not corrupted
- Review error messages for specific issues

---

## Recovery Procedures

### Disaster Recovery Plan

**1. Regular Testing**:
```bash
# Monthly: Test restore on development system
./restore.sh -f latest_backup.sql.gz

# Verify data integrity
psql -d terminal_velocity -c "SELECT COUNT(*) FROM players;"
```

**2. Emergency Restore**:
```bash
# Stop server
systemctl stop terminal-velocity

# Restore from latest backup
./restore.sh latest_backup.sql.gz

# Verify restore
./restore.sh --list

# Start server
systemctl start terminal-velocity
```

**3. Point-in-Time Recovery**:
```bash
# List backups by time
./restore.sh --list

# Restore from specific time
./restore.sh backups/terminal_velocity_20250115_120000.sql.gz
```

---

## API Reference

### Backup Script

```bash
backup.sh [OPTIONS]

OPTIONS:
  -d, --dir DIR         Backup directory (default: backups)
  -r, --retention DAYS  Keep backups for N days (default: 7)
  -c, --count COUNT     Keep last N backups (default: 10)
  -n, --no-compress     Don't compress backup files
  -h, --help            Show help message

ENVIRONMENT:
  DB_HOST              Database host (default: localhost)
  DB_PORT              Database port (default: 5432)
  DB_USER              Database user (default: terminal_velocity)
  DB_PASSWORD          Database password
  DB_NAME              Database name (default: terminal_velocity)
```

### Restore Script

```bash
restore.sh [OPTIONS] <backup_file>

OPTIONS:
  -l, --list           List available backups
  -f, --force          Skip confirmation prompt
  -h, --help           Show help message

ENVIRONMENT:
  DB_HOST              Database host (default: localhost)
  DB_PORT              Database port (default: 5432)
  DB_USER              Database user (default: terminal_velocity)
  DB_PASSWORD          Database password
  DB_NAME              Database name (default: terminal_velocity)
  BACKUP_DIR           Backup directory (default: backups)
```

---

## Related Documentation

- [Admin System](./ADMIN_SYSTEM.md) - Database management
- [Deployment Guide](./DEPLOYMENT.md) - Production setup

---

## File Locations

**Scripts**:
- `scripts/backup.sh` - Backup script
- `scripts/restore.sh` - Restore script
- `scripts/crontab.example` - Example cron configuration

**Documentation**:
- `docs/BACKUP_RESTORE.md` - This file
- `ROADMAP.md` - Phase 20 details

---

**For questions about backups and recovery, contact the development team or system administrator.**
