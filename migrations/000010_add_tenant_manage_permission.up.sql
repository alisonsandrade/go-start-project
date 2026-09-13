INSERT INTO permissions (code, description)
VALUES ('tenant:manage', 'Create, read, update and delete tenants')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'ADMIN'
  AND p.code = 'tenant:manage'
ON CONFLICT (role_id, permission_id) DO NOTHING;
