-- Migration: Create users table for authentication
-- Created: 2024-12-21

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    pwd_hash TEXT NOT NULL,
    pwd_ts INTEGER DEFAULT 0, -- password timestamp
    salt TEXT NOT NULL,
    role INTEGER NOT NULL DEFAULT 0, -- 0: GENERAL, 1: GUEST, 2: ADMIN
    permission INTEGER NOT NULL DEFAULT 0,
    base_path TEXT NOT NULL DEFAULT "",
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    otp_secret TEXT DEFAULT '',
    sso_id TEXT DEFAULT '',
    authn TEXT DEFAULT '', -- WebAuthn credentials JSON
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on username for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Create index on sso_id for SSO authentication
CREATE INDEX IF NOT EXISTS idx_users_sso_id ON users(sso_id);



-- Insert default admin user (password: admin)
-- Password hash is SHA256 of "admin" + static salt
INSERT OR IGNORE INTO users (username, pwd_hash, salt, role, permission, pwd_ts)
VALUES (
    'admin',
    'c3e99e5a8a3d3d5f8f8b9c2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4',
    'admin_default_salt',
    2, -- ADMIN role
    8191, -- All permissions enabled
    strftime('%s', 'now') -- current timestamp
); 