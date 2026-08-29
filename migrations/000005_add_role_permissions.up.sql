INSERT INTO permissions (code, description)
VALUES
    ('role:read', 'Read roles'),
    ('role:create', 'Create roles'),
    ('role:update', 'Update roles'),
    ('role:delete', 'Delete roles'),
    ('role:assign-permissions', 'Assign permissions to roles');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'ADMIN'
AND p.code IN (
    'role:read',
    'role:create',
    'role:update',
    'role:delete',
    'role:assign-permissions'
);
