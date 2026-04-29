#!/bin/bash
# Database backup script for mtsn2kolut-super-app
# Usage: ./deploy/scripts/backup.sh [backup_directory]

set -euo pipefail

# Configuration
BACKEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/services/core-api"
DB_SCRIPTS_DIR="$BACKEND_DIR/db/scripts"
BACKUP_DIR="${1:-/backups/mtsn2kolut}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Load environment variables
if [ -f "$BACKEND_DIR/.env" ]; then
    export $(grep -v '^#' "$BACKEND_DIR/.env" | xargs)
fi

# Check if DATABASE_URL is set
if [ -z "${DATABASE_URL:-}" ]; then
    echo "ERROR: DATABASE_URL environment variable is not set"
    echo "Make sure .env file exists in services/core-api/ or export DATABASE_URL"
    exit 1
fi

# Perform backup
BACKUP_FILE="$BACKUP_DIR/pusaka_backup_$TIMESTAMP.sql.gz"
echo "Starting database backup to $BACKUP_FILE..."

if pg_dump --clean --if-exists --no-owner --no-privileges "$DATABASE_URL" | gzip > "$BACKUP_FILE"; then
    echo "Backup completed successfully: $BACKUP_FILE"
    
    # Verify backup file was created and is not empty
    if [ ! -s "$BACKUP_FILE" ]; then
        echo "ERROR: Backup file is empty"
        rm -f "$BACKUP_FILE"
        exit 1
    fi
    
    # Clean up old backups
    echo "Cleaning up backups older than $RETENTION_DAYS days..."
    find "$BACKUP_DIR" -name "pusaka_backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete
    
    # List remaining backups
    echo "Remaining backups:"
    ls -lh "$BACKUP_DIR"/pusaka_backup_*.sql.gz 2>/dev/null || echo "No backups found"
else
    echo "ERROR: Backup failed"
    exit 1
fi
