/* Add 'invited' status to group_members status check constraint */

-- SQLite doesn't support ALTER TABLE for CHECK constraints
-- We need to recreate the table with the updated constraint

-- Create new table with updated constraint
CREATE TABLE group_members_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT CHECK (role IN ('admin', 'member')) NOT NULL DEFAULT 'member',
    status TEXT CHECK (status IN ('pending', 'accepted', 'invited')) NOT NULL DEFAULT 'accepted',
    joined_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(group_id, user_id)
);

-- Copy data from old table
INSERT INTO group_members_new (id, group_id, user_id, role, status, joined_at)
SELECT id, group_id, user_id, role, status, joined_at
FROM group_members;

-- Drop old table
DROP TABLE group_members;

-- Rename new table
ALTER TABLE group_members_new RENAME TO group_members;

-- Recreate indexes
CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE INDEX idx_group_members_role ON group_members(role);
CREATE INDEX idx_group_members_status ON group_members(status);
