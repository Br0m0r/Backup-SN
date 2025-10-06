/* we use ON DELETE SET NULL to retain message history 
even if a user is deleted */


CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    receiver_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    image_url TEXT,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);