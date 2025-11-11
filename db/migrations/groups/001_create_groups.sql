-- Group Service Database Schema
-- Owns: Groups, members, events, and event responses

CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    image_url TEXT,
    creator_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE group_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT CHECK (role IN ('admin', 'member')) NOT NULL DEFAULT 'member',
    status TEXT CHECK (status IN ('pending', 'accepted')) NOT NULL DEFAULT 'accepted',
    joined_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    UNIQUE(group_id, user_id)
);

CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE INDEX idx_group_members_role ON group_members(role);
CREATE INDEX idx_group_members_status ON group_members(status);

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    creator_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    event_time DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX idx_events_group_id ON events(group_id);
CREATE INDEX idx_events_creator_id ON events(creator_id);
CREATE INDEX idx_events_event_time ON events(event_time);
CREATE INDEX idx_events_created_at ON events(created_at);

CREATE TABLE event_responses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    response TEXT CHECK (response IN ('going', 'not_going', 'interested')) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    UNIQUE(event_id, user_id)
);

CREATE INDEX idx_event_responses_event_id ON event_responses(event_id);
CREATE INDEX idx_event_responses_user_id ON event_responses(user_id);
CREATE INDEX idx_event_responses_response ON event_responses(response);

-- Cache table for user data (denormalized for performance)
CREATE TABLE user_cache (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    avatar_path TEXT,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_user_cache_username ON user_cache(username);
