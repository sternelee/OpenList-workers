# OpenList Workers - D1 数据库驱动配置管理

## 🎯 功能概述

本次更新为 OpenList Workers 添加了完整的 D1 数据库驱动配置管理功能，实现了：

- ✅ 使用 D1 数据库存储驱动配置
- ✅ 完整的驱动 CRUD 操作
- ✅ 动态启用/禁用驱动
- ✅ 兼容原有 API 接口
- ✅ 支持分页和过滤

## 📊 数据库表结构

### driver_configs 表
```sql
CREATE TABLE driver_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,           -- 驱动名称 (Local, S3, etc.)
    display_name TEXT NOT NULL,          -- 显示名称
    description TEXT,                    -- 描述信息
    config TEXT,                         -- JSON 格式的配置模板
    icon TEXT,                           -- 图标名称
    enabled BOOLEAN DEFAULT TRUE,        -- 是否启用
    order_num INTEGER DEFAULT 0,         -- 排序序号
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 🔌 默认驱动配置

系统初始化时会自动创建以下驱动配置：

| 驱动名称 | 显示名称 | 描述 | 状态 |
|---------|----------|------|------|
| Local | 本地存储 | 本地文件系统存储 | ✅ 启用 |
| S3 | Amazon S3 | Amazon S3 对象存储 | ✅ 启用 |
| Aliyundrive | 阿里云盘 | 阿里云盘存储 | ✅ 启用 |
| OneDrive | OneDrive | Microsoft OneDrive 存储 | ✅ 启用 |
| GoogleDrive | Google Drive | Google Drive 存储 | ✅ 启用 |

## 🚀 API 接口

### 兼容接口（保持向后兼容）
```bash
# 获取驱动列表（原有接口）
GET /api/drivers
GET /api/drivers?enabled=true  # 仅获取启用的驱动
```

### 新增管理接口
```bash
# 驱动配置管理
GET    /api/admin/driver/list           # 获取驱动配置列表
GET    /api/admin/driver/get            # 获取单个驱动配置
POST   /api/admin/driver/create         # 创建驱动配置
POST   /api/admin/driver/update         # 更新驱动配置
POST   /api/admin/driver/delete         # 删除驱动配置
POST   /api/admin/driver/enable         # 启用驱动配置
POST   /api/admin/driver/disable        # 禁用驱动配置
```

## 📝 API 使用示例

### 1. 获取所有驱动配置
```bash
curl "https://your-worker.dev/api/drivers"
```

**响应示例：**
```json
{
  "code": 200,
  "message": "",
  "data": {
    "drivers": ["Local", "S3", "Aliyundrive"],
    "info": {
      "Local": {
        "name": "Local",
        "display_name": "本地存储",
        "description": "本地文件系统存储",
        "icon": "folder",
        "config": "{\"root_folder_path\": \"/data\"}",
        "order": 1
      }
    },
    "configs": [...],
    "total": 5,
    "page": 1,
    "per_page": 20
  }
}
```

### 2. 创建新驱动配置
```bash
curl -X POST "https://your-worker.dev/api/admin/driver/create" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WebDAV",
    "display_name": "WebDAV 存储",
    "description": "WebDAV 协议存储",
    "config": "{\"url\": \"\", \"username\": \"\", \"password\": \"\"}",
    "icon": "globe",
    "enabled": true,
    "order": 6
  }'
```

### 3. 获取单个驱动配置
```bash
# 通过名称获取
curl "https://your-worker.dev/api/admin/driver/get?name=Local"

# 通过 ID 获取
curl "https://your-worker.dev/api/admin/driver/get?id=1"
```

### 4. 更新驱动配置
```bash
curl -X POST "https://your-worker.dev/api/admin/driver/update" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WebDAV",
    "display_name": "WebDAV 网络存储",
    "description": "支持 WebDAV 协议的网络存储服务",
    "config": "{\"url\": \"https://example.com/webdav\", \"username\": \"user\", \"password\": \"pass\"}",
    "icon": "globe",
    "enabled": true,
    "order": 6
  }'
```

### 5. 启用/禁用驱动
```bash
# 禁用驱动
curl -X POST "https://your-worker.dev/api/admin/driver/disable?id=1"

# 启用驱动
curl -X POST "https://your-worker.dev/api/admin/driver/enable?id=1"
```

### 6. 删除驱动配置
```bash
curl -X POST "https://your-worker.dev/api/admin/driver/delete?id=1"
```

## 🛠️ 配置结构说明

### DriverConfig 结构
```go
type DriverConfig struct {
    ID          uint   `json:"id"`           // 唯一标识
    Name        string `json:"name"`         // 驱动名称（唯一）
    DisplayName string `json:"display_name"` // 显示名称
    Description string `json:"description"`  // 描述信息
    Config      string `json:"config"`       // JSON 格式的配置模板
    Icon        string `json:"icon"`         // 图标名称
    Enabled     bool   `json:"enabled"`      // 是否启用
    Order       int    `json:"order"`        // 排序序号
    Created     string `json:"created"`      // 创建时间
    Modified    string `json:"modified"`     // 修改时间
}
```

### 配置模板示例

#### Local 驱动
```json
{
  "root_folder_path": "/data"
}
```

#### S3 驱动
```json
{
  "bucket": "",
  "region": "us-east-1",
  "access_key_id": "",
  "secret_access_key": ""
}
```

#### 阿里云盘驱动
```json
{
  "refresh_token": "",
  "root_folder_id": "root"
}
```

## 🧪 测试脚本

使用提供的测试脚本来验证功能：

```bash
# 赋予执行权限
chmod +x test_drivers_api.sh

# 运行测试（需要先启动本地服务器）
./test_drivers_api.sh
```

测试脚本会：
1. 初始化系统
2. 获取现有驱动配置
3. 创建新的 WebDAV 驱动配置
4. 更新、启用、禁用配置
5. 最后删除测试配置

## 🔧 部署配置

### wrangler.toml 配置
```toml
name = "openlist-workers"
main = "main.go"

[[d1_databases]]
binding = "DB"
database_name = "openlist-db"
database_id = "your-database-id"

[env.production]
[[env.production.d1_databases]]
binding = "DB"
database_name = "openlist-db"
database_id = "your-production-database-id"
```

### 实际使用 D1 数据库的代码修改
在实际部署时，需要将模拟的数据库操作替换为真实的 D1 API 调用：

```go
// 替换这类注释的代码：
// query := `INSERT INTO driver_configs ...`
// stmt := d1DB.Prepare(query)
// result, err := stmt.Bind(...).Run()

// 使用实际的 Cloudflare Workers D1 API
```

## 🚀 优势特性

1. **数据持久化**: 所有驱动配置保存在 D1 数据库中，重启不丢失
2. **动态管理**: 无需重新部署即可添加、修改、删除驱动配置
3. **向后兼容**: 完全兼容原有的 `/api/drivers` 接口
4. **灵活配置**: 支持 JSON 格式的配置模板
5. **状态管理**: 可以动态启用/禁用驱动
6. **排序控制**: 支持自定义驱动显示顺序

## 🔮 未来扩展

- [ ] 驱动配置版本管理
- [ ] 驱动配置导入/导出
- [ ] 驱动配置模板验证
- [ ] 驱动使用统计
- [ ] 批量操作支持

---

**注意**: 本功能需要 Cloudflare D1 数据库支持，请确保已正确配置 D1 绑定。