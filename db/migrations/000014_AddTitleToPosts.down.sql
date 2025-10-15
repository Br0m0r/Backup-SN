-- Remove title column from posts table
-- Note: SQLite doesn't support DROP COLUMN directly in older versions
-- This would require recreating the table in production
ALTER TABLE posts DROP COLUMN title;
