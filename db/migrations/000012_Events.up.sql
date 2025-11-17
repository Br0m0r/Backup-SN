/* Group events */

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    creator_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    event_time DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_events_group_id ON events(group_id);
CREATE INDEX idx_events_creator_id ON events(creator_id);
CREATE INDEX idx_events_event_time ON events(event_time);
CREATE INDEX idx_events_created_at ON events(created_at);
