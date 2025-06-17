-- 存储配置表（支持多用户）
CREATE TABLE IF NOT EXISTS storages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    mount_path TEXT NOT NULL,
    order_index INTEGER DEFAULT 0,
    driver TEXT NOT NULL,
    cache_expiration INTEGER DEFAULT 0,
    status TEXT DEFAULT '',
    addition TEXT DEFAULT '',
    remark TEXT DEFAULT '',
    modified DATETIME DEFAULT CURRENT_TIMESTAMP,
    disabled BOOLEAN DEFAULT FALSE,
    disable_index BOOLEAN DEFAULT FALSE,
    enable_sign BOOLEAN DEFAULT FALSE,
    
    -- Access control fields
    is_public BOOLEAN DEFAULT FALSE,
    allow_guest BOOLEAN DEFAULT FALSE,
    require_auth BOOLEAN DEFAULT TRUE,
    
    -- Sort fields
    order_by TEXT DEFAULT '',
    order_direction TEXT DEFAULT '',
    extract_folder TEXT DEFAULT '',
    
    -- Proxy fields
    web_proxy BOOLEAN DEFAULT FALSE,
    webdav_policy TEXT DEFAULT '',
    proxy_range BOOLEAN DEFAULT FALSE,
    down_proxy_url TEXT DEFAULT '',
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Unique constraint for user + mount_path
    UNIQUE(user_id, mount_path)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_storages_user_id ON storages(user_id);
CREATE INDEX IF NOT EXISTS idx_storages_mount_path ON storages(mount_path);
CREATE INDEX IF NOT EXISTS idx_storages_order ON storages(order_index);
CREATE INDEX IF NOT EXISTS idx_storages_driver ON storages(driver);
CREATE INDEX IF NOT EXISTS idx_storages_status ON storages(status);
CREATE INDEX IF NOT EXISTS idx_storages_public ON storages(is_public);
CREATE INDEX IF NOT EXISTS idx_storages_user_path ON storages(user_id, mount_path); 