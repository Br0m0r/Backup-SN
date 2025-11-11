-- User Service Database Schema
-- Owns: User profiles and following relationships

CREATE TABLE user_profiles (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    avatar_path TEXT,
    nickname TEXT,
    about_me TEXT,
    is_public_profile BOOLEAN NOT NULL DEFAULT 1,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE follows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follower_id INTEGER NOT NULL,
    following_id INTEGER NOT NULL,
    status TEXT CHECK (status IN ('pending', 'accepted')) NOT NULL DEFAULT 'accepted',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(follower_id, following_id)
);

CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_following_id ON follows(following_id);
CREATE INDEX idx_follows_status ON follows(status);
