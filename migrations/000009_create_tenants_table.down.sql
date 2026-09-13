ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS fk_audit_logs_tenant;
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_tokens_tenant;
ALTER TABLE roles DROP CONSTRAINT IF EXISTS fk_roles_tenant;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_tenant;
DROP TABLE IF EXISTS tenants;
