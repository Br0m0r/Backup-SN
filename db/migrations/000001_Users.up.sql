CREATE TABLE users (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    Avatar_url TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    about_me TEXT ,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    is_public_profile BOOLEAN NOT NULL DEFAULT 1
);
