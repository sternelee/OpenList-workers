-- 存储配置表
CREATE TABLE IF NOT EXISTS storages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mount_path TEXT NOT NULL UNIQUE,
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_storages_mount_path ON storages(mount_path);
CREATE INDEX IF NOT EXISTS idx_storages_order ON storages(order_index);
CREATE INDEX IF NOT EXISTS idx_storages_driver ON storages(driver);
CREATE INDEX IF NOT EXISTS idx_storages_status ON storages(status); 