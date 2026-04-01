-- ============================================================
-- Audit Log Optimization - Phase 1
-- Date: 2026-04-01
-- Purpose: Add log_type field for filtering, optimize query performance
-- ============================================================

-- 1. Add log_type field for categorizing logs
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS log_type VARCHAR(20) DEFAULT 'operation';

-- Log type values:
-- 'operation' - Business operations (create, update, delete, payment, etc.)
-- 'access'      - Read operations (GET requests)
-- 'system'     - System operations (login, logout, config changes)

-- 2. Add response_time_ms for performance tracking
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS response_time_ms INTEGER DEFAULT 0;

-- 3. Create index on new fields for better query performance
CREATE INDEX IF NOT EXISTS idx_audit_log_type ON audit_logs(log_type);
CREATE INDEX IF NOT EXISTS idx_audit_method ON audit_logs(request_method);
CREATE INDEX IF NOT EXISTS idx_audit_log_type_created ON audit_logs(log_type, created_at DESC);

-- 4. Add constraint to limit body size (prevent data bloat)
ALTER TABLE audit_logs ADD CONSTRAINT chk_request_body_size 
CHECK (length(request_body) <= 50000);

ALTER TABLE audit_logs ADD CONSTRAINT chk_response_body_size 
CHECK (length(response_body) <= 50000);

-- 5. Update existing records to set log_type based on request_method
UPDATE audit_logs SET log_type = 'access' WHERE request_method = 'GET' AND log_type IS NULL;
UPDATE audit_logs SET log_type = 'operation' WHERE request_method != 'GET' AND log_type IS NULL;
UPDATE audit_logs SET log_type = 'operation' WHERE log_type IS NULL;

-- 6. Set default for log_type
ALTER TABLE audit_logs ALTER COLUMN log_type SET DEFAULT 'operation';
ALTER TABLE audit_logs ALTER COLUMN log_type SET NOT NULL;

-- ============================================================
-- Verification
-- ============================================================
-- SELECT log_type, COUNT(*) FROM audit_logs GROUP BY log_type;
-- SELECT 'audit_logs' as table_name, COUNT(*) as row_count FROM audit_logs;
