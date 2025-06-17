-- 系统设置表
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT DEFAULT '',
    help TEXT DEFAULT '',
    type TEXT DEFAULT 'string',
    options TEXT DEFAULT '',
    group_id INTEGER DEFAULT 0,
    flag INTEGER DEFAULT 0,
    index_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_settings_group ON settings(group_id);
CREATE INDEX IF NOT EXISTS idx_settings_flag ON settings(flag);
CREATE INDEX IF NOT EXISTS idx_settings_index_order ON settings(index_order); 