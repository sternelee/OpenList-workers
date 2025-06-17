-- 添加用户存储支持的迁移脚本
-- 如果storages表已存在但没有用户字段，这个脚本会添加必要的字段

-- 检查并添加user_id字段
ALTER TABLE storages ADD COLUMN IF NOT EXISTS user_id INTEGER NOT NULL DEFAULT 1;

-- 检查并添加访问控制字段
ALTER TABLE storages ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT FALSE;
ALTER TABLE storages ADD COLUMN IF NOT EXISTS allow_guest BOOLEAN DEFAULT FALSE;
ALTER TABLE storages ADD COLUMN IF NOT EXISTS require_auth BOOLEAN DEFAULT TRUE;

-- 添加外键约束 (如果不存在)
-- 注意：SQLite中添加外键约束比较复杂，这里仅作为注释说明
-- 实际环境中可能需要重新创建表来添加外键约束

-- 创建新索引
CREATE INDEX IF NOT EXISTS idx_storages_user_id ON storages(user_id);
CREATE INDEX IF NOT EXISTS idx_storages_public ON storages(is_public);
CREATE INDEX IF NOT EXISTS idx_storages_user_path ON storages(user_id, mount_path);

-- 删除旧的唯一约束并创建新的 (仅针对mount_path的唯一约束)
-- 注意：SQLite中修改约束比较复杂，这里仅作为注释说明
-- 新的唯一约束应该是 UNIQUE(user_id, mount_path)

-- 更新现有数据，设置默认用户ID
-- 假设第一个用户ID为1，管理员用户
UPDATE storages SET user_id = 1 WHERE user_id = 0 OR user_id IS NULL;

-- 设置一些合理的默认值
UPDATE storages SET is_public = FALSE WHERE is_public IS NULL;
UPDATE storages SET allow_guest = FALSE WHERE allow_guest IS NULL;
UPDATE storages SET require_auth = TRUE WHERE require_auth IS NULL; 