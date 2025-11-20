-- Remove image support from messages
DROP INDEX IF EXISTS idx_messages_image_path;

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE messages_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER,
    recipient_id INTEGER,
    content TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO messages_backup SELECT id, sender_id, recipient_id, content, is_read, created_at FROM messages;

DROP TABLE messages;

ALTER TABLE messages_backup RENAME TO messages;

CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_recipient_id ON messages(recipient_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_is_read ON messages(is_read);
