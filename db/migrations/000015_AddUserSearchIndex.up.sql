-- Add indexes for faster user search
-- These indexes will speed up LIKE queries on username, first_name, last_name

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_first_name ON users(first_name);
CREATE INDEX IF NOT EXISTS idx_users_last_name ON users(last_name);

-- Composite index for full name search
CREATE INDEX IF NOT EXISTS idx_users_full_name ON users(first_name, last_name);
