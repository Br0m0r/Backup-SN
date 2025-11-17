-- Post Service Database Schema
-- Owns: Posts, comments, and post privacy

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    image_path TEXT,
    privacy_level TEXT CHECK (privacy_level IN ('public', 'private', 'almost_private')) NOT NULL DEFAULT 'public',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_created_at ON posts(created_at);
CREATE INDEX idx_posts_privacy_level ON posts(privacy_level);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);

CREATE TABLE post_viewers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    UNIQUE(post_id, user_id)
);

CREATE INDEX idx_post_viewers_post_id ON post_viewers(post_id);
CREATE INDEX idx_post_viewers_user_id ON post_viewers(user_id);

-- Cache table for user data (denormalized for performance)
CREATE TABLE user_cache (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    avatar_path TEXT,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_user_cache_username ON user_cache(username);
