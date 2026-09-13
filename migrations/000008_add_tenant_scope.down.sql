DROP INDEX IF EXISTS idx_audit_logs_tenant_id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_refresh_tokens_tenant_id;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_roles_tenant_id;
ALTER TABLE roles DROP COLUMN IF EXISTS tenant_id;

DROP INDEX IF EXISTS idx_users_tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
