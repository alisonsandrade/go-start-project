ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Users that already existed before email confirmation was introduced remain
-- valid accounts. New registrations and invitations start as unverified.
UPDATE users SET email_verified = TRUE WHERE email_verified = FALSE;