/* Primary key here are 2 ids followers and followed    
 to ensure a user can't follow the same user more than once */


CREATE TABLE follows (
    follower_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (follower_id, followed_id)
);