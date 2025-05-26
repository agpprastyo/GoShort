-- Drop trigger first
DROP TRIGGER IF EXISTS update_users_timestamp ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_timestamp();

-- Drop indexes
DROP INDEX IF EXISTS idx_users_name;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_role;

-- Drop the table
DROP TABLE IF EXISTS users;

-- Drop the enum type
-- Note: Must be done after dropping table that uses it
DROP TYPE IF EXISTS user_role;

-- Optional: Drop extension if no other objects depend on it
-- Commented out since other tables might use this extension
-- DROP EXTENSION IF EXISTS pgcrypto;