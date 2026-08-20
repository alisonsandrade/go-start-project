-- Permissions: the catalog of fine-grained actions in the system.
CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        VARCHAR(100) UNIQUE NOT NULL,
    description VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Roles: named sets of permissions. System roles cannot be modified or deleted.
CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(255),
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Join table linking roles to permissions (many-to-many).
CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Seed: permissions catalog
INSERT INTO permissions (code, description) VALUES
    ('user:read',   'Read a user profile'),
    ('user:create', 'Create a user'),
    ('user:update', 'Update a user'),
    ('user:delete', 'Delete a user'),
    ('user:list',   'List all users');

-- Seed: system roles (protected against modification/deletion)
INSERT INTO roles (name, description, is_system) VALUES
    ('ADMIN', 'System administrator with full access', TRUE),
    ('USER',  'Default user with basic access',        TRUE);

-- Seed: ADMIN gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'ADMIN';

-- Seed: USER gets only user:read
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'user:read'
WHERE r.name = 'USER';
