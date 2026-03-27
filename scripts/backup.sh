#!/bin/bash

# gAPI Platform Backup Script
# Usage: ./scripts/backup.sh

set -e

BACKUP_DIR="${BACKUP_DIR:-./backups}"
DATE=$(date +%Y%m%d_%H%M%S)

echo "=== gAPI Platform Backup ==="
echo "Backup started at: $(date)"

# Create backup directory if not exists
mkdir -p "$BACKUP_DIR"

# Load environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

DB_PASSWORD="${DB_PASSWORD:-gapi123}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"

# Backup PostgreSQL
echo "Backing up PostgreSQL..."
docker exec gapi-postgres pg_dump -U "${DB_USER:-gapi}" "${DB_NAME:-gapi}" > "$BACKUP_DIR/db_${DATE}.sql"
echo "PostgreSQL backup saved: $BACKUP_DIR/db_${DATE}.sql"

# Backup Redis
echo "Backing up Redis..."
if [ -n "$REDIS_PASSWORD" ]; then
    docker exec gapi-redis redis-cli -a "$REDIS_PASSWORD" SAVE
else
    docker exec gapi-redis redis-cli SAVE
fi
docker cp gapi-redis:/data/dump.rdb "$BACKUP_DIR/redis_${DATE}.rdb"
echo "Redis backup saved: $BACKUP_DIR/redis_${DATE}.rdb"

# Backup RabbitMQ
echo "Backing up RabbitMQ..."
docker cp gapi-rabbitmq:/var/lib/rabbitmq/mnesia "$BACKUP_DIR/rabbitmq_${DATE}" || true
echo "RabbitMQ backup saved: $BACKUP_DIR/rabbitmq_${DATE}"

# Cleanup old backups (keep last 7 days)
echo "Cleaning up old backups..."
find "$BACKUP_DIR" -type f -mtime +7 -delete 2>/dev/null || true
find "$BACKUP_DIR" -type d -mtime +7 -empty -delete 2>/dev/null || true

echo ""
echo "=== Backup Completed ==="
echo "Backup completed at: $(date)"
echo "Backup files:"
ls -lh "$BACKUP_DIR" | tail -10
