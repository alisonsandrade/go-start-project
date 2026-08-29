DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN (
        'role:read',
        'role:create',
        'role:update',
        'role:delete',
        'role:assign-permissions'
    )
);

DELETE FROM permissions
WHERE code IN (
    'role:read',
    'role:create',
    'role:update',
    'role:delete',
    'role:assign-permissions'
);
