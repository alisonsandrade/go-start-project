-- Add new role reference column
ALTER TABLE users
ADD COLUMN role_id UUID;

-- Backfill existing users
UPDATE users u
SET role_id = r.id
FROM roles r
WHERE UPPER(u.role) = r.name;

-- Make mandatory
ALTER TABLE users
ALTER COLUMN role_id SET NOT NULL;

-- Foreign key
ALTER TABLE users
ADD CONSTRAINT fk_users_role
FOREIGN KEY (role_id)
REFERENCES roles(id);

-- Remove old column
ALTER TABLE users
DROP COLUMN role;
