DROP INDEX IF EXISTS idx_audit_logs_deleted_at;
DROP INDEX IF EXISTS idx_refresh_tokens_deleted_at;
DROP INDEX IF EXISTS idx_permissions_deleted_at;
DROP INDEX IF EXISTS idx_roles_deleted_at;

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE refresh_tokens
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE permissions
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE roles
    DROP COLUMN IF EXISTS deleted_at;
