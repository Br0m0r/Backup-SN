-- Add group_id column to posts table for group posts
ALTER TABLE posts ADD COLUMN group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE;

-- Create index for efficient group post queries
CREATE INDEX idx_posts_group_id ON posts(group_id);
