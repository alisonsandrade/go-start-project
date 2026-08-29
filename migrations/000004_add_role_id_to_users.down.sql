-- Restore old column
ALTER TABLE users
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'USER';

-- Restore values
UPDATE users u
SET role = r.name
FROM roles r
WHERE u.role_id = r.id;

-- Drop FK
ALTER TABLE users
DROP CONSTRAINT IF EXISTS fk_users_role;

-- Drop new column
ALTER TABLE users
DROP COLUMN role_id;
