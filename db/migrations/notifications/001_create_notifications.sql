-- Notification Service Database Schema
-- Owns: All notifications

CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT CHECK (type IN ('follow', 'follow_request', 'group_invite', 'group_request', 'event', 'message', 'comment', 'post')) NOT NULL,
    related_id INTEGER,
    content TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- Cache table for user data (denormalized for performance)
CREATE TABLE user_cache (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    avatar_path TEXT,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_user_cache_username ON user_cache(username);
