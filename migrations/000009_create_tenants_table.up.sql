CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(150) NOT NULL,
    document VARCHAR(30) NOT NULL UNIQUE,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at);

INSERT INTO tenants (id, name, document, settings)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Default Tenant',
    'DEFAULT-TENANT',
    '{}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE users
    ADD CONSTRAINT fk_users_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE roles
    ADD CONSTRAINT fk_roles_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE refresh_tokens
    ADD CONSTRAINT fk_refresh_tokens_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE audit_logs
    ADD CONSTRAINT fk_audit_logs_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);
