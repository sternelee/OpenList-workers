-- 搜索节点表
CREATE TABLE IF NOT EXISTS search_nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent TEXT NOT NULL,
    name TEXT NOT NULL,
    is_dir BOOLEAN DEFAULT FALSE,
    size INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_search_nodes_parent ON search_nodes(parent);
CREATE INDEX IF NOT EXISTS idx_search_nodes_name ON search_nodes(name);
CREATE INDEX IF NOT EXISTS idx_search_nodes_is_dir ON search_nodes(is_dir);
CREATE INDEX IF NOT EXISTS idx_search_nodes_parent_name ON search_nodes(parent, name); 