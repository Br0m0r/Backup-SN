-- Remove user search indexes

DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_first_name;
DROP INDEX IF EXISTS idx_users_last_name;
DROP INDEX IF EXISTS idx_users_full_name;
