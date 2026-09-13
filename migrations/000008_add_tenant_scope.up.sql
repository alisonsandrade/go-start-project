-- Add tenant ownership to all tenant-scoped records.
-- The default tenant keeps existing installations bootstrappable; authenticated
-- requests overwrite it through BaseModelTenant.BeforeCreate.
DO $$
DECLARE
    default_tenant UUID := '00000000-0000-0000-0000-000000000001';
BEGIN
    ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID;
    UPDATE users SET tenant_id = default_tenant WHERE tenant_id IS NULL;
    ALTER TABLE users ALTER COLUMN tenant_id SET DEFAULT '00000000-0000-0000-0000-000000000001'::uuid;
    ALTER TABLE users ALTER COLUMN tenant_id SET NOT NULL;
    CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);

    ALTER TABLE roles ADD COLUMN IF NOT EXISTS tenant_id UUID;
    UPDATE roles SET tenant_id = default_tenant WHERE tenant_id IS NULL;
    ALTER TABLE roles ALTER COLUMN tenant_id SET DEFAULT '00000000-0000-0000-0000-000000000001'::uuid;
    ALTER TABLE roles ALTER COLUMN tenant_id SET NOT NULL;
    CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);

    ALTER TABLE refresh_tokens ADD COLUMN IF NOT EXISTS tenant_id UUID;
    UPDATE refresh_tokens rt SET tenant_id = u.tenant_id FROM users u WHERE rt.user_id = u.id AND rt.tenant_id IS NULL;
    UPDATE refresh_tokens SET tenant_id = default_tenant WHERE tenant_id IS NULL;
    ALTER TABLE refresh_tokens ALTER COLUMN tenant_id SET DEFAULT '00000000-0000-0000-0000-000000000001'::uuid;
    ALTER TABLE refresh_tokens ALTER COLUMN tenant_id SET NOT NULL;
    CREATE INDEX IF NOT EXISTS idx_refresh_tokens_tenant_id ON refresh_tokens(tenant_id);

    ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS tenant_id UUID;
    UPDATE audit_logs SET tenant_id = default_tenant WHERE tenant_id IS NULL;
    ALTER TABLE audit_logs ALTER COLUMN tenant_id SET DEFAULT '00000000-0000-0000-0000-000000000001'::uuid;
    ALTER TABLE audit_logs ALTER COLUMN tenant_id SET NOT NULL;
    CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
END $$;
