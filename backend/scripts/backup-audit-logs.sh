#!/bin/bash
# ============================================================
# Audit Logs Backup Script
# Schedule: Run every 2 days via cron
# Usage: ./backup-audit-logs.sh
# ============================================================

set -e

# Configuration
BACKUP_DIR="/var/backups/gapi/audit-logs"
RETENTION_DAYS=30
DB_NAME="gapi"
DB_USER="gapi"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="/var/log/backup-audit-logs.log"

# Create backup directory if not exists
mkdir -p "$BACKUP_DIR"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "=========================================="
log "Starting audit logs backup"
log "=========================================="

# Export audit logs older than 7 days to CSV
log "Exporting audit logs to CSV..."

EXPORT_FILE="$BACKUP_DIR/audit_logs_${TIMESTAMP}.csv"

# Export logs from 7 days ago to 2 days ago (incremental backup)
psql -U "$DB_USER" -d "$DB_NAME" -c "\COPY (
    SELECT 
        id, user_id, username, action, action_group, resource_type, resource_id,
        request_method, request_path, request_ip, success, error_message,
        log_type, response_time_ms, created_at
    FROM audit_logs 
    WHERE created_at >= NOW() - INTERVAL '7 days'
    AND created_at < NOW() - INTERVAL '2 days'
    ORDER BY created_at DESC
) TO STDOUT WITH CSV HEADER" > "$EXPORT_FILE"

if [ -f "$EXPORT_FILE" ]; then
    FILE_SIZE=$(du -h "$EXPORT_FILE" | cut -f1)
    ROW_COUNT=$(wc -l < "$EXPORT_FILE")
    log "Backup created: $EXPORT_FILE (Size: $FILE_SIZE, Rows: $ROW_COUNT)"
else
    log "ERROR: Backup file not created"
    exit 1
fi

# Compress the backup
log "Compressing backup..."
gzip "$EXPORT_FILE"
log "Compressed: ${EXPORT_FILE}.gz"

# Delete logs older than 7 days from database
log "Deleting logs older than 7 days from database..."
DELETED_COUNT=$(psql -U "$DB_USER" -d "$DB_NAME" -t -c "
    WITH deleted AS (
        DELETE FROM audit_logs 
        WHERE created_at < NOW() - INTERVAL '7 days'
        RETURNING id
    )
    SELECT COUNT(*) FROM deleted
")
log "Deleted $DELETED_COUNT records from database"

# Clean up old backups (keep for 30 days)
log "Cleaning up backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -name "audit_logs_*.csv.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "audit_logs_*.csv" -mtime +$RETENTION_DAYS -delete

# Optimize table after deletion
log "Optimizing table..."
psql -U "$DB_USER" -d "$DB_NAME" -c "VACUUM ANALYZE audit_logs;"

# Summary
BACKUP_COUNT=$(find "$BACKUP_DIR" -name "audit_logs_*.gz" | wc -l)
log "=========================================="
log "Backup completed successfully"
log "Total backup files: $BACKUP_COUNT"
log "=========================================="
