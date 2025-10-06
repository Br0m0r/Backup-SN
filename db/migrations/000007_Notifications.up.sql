CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    related_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    notification_type  TEXT CHECK (notification_type IN ('message', 'follow_request', 'group_invite','comment')) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);