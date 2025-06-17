# OpenList Workers 多用户驱动管理指南

## 概述

OpenList Workers 现已支持完整的多用户驱动管理系统，每个用户可以独立管理自己的存储配置，支持用户权限控制和公开存储访问。

## 🚀 核心特性

### 1. **用户隔离存储**
- ✅ 每个用户拥有独立的存储配置
- ✅ 用户级别的挂载路径管理
- ✅ 存储配置完全隔离，互不干扰

### 2. **权限控制**
- ✅ 基于用户的访问控制
- ✅ 公开存储支持（allow_guest）
- ✅ 匿名访问支持
- ✅ 管理员权限管理

### 3. **存储访问控制**
- ✅ `is_public`: 是否公开访问
- ✅ `allow_guest`: 是否允许访客访问
- ✅ `require_auth`: 是否需要认证

## 📊 数据库架构更新

### 存储表字段扩展
```sql
-- 新增字段
user_id INTEGER NOT NULL          -- 所属用户ID
is_public BOOLEAN DEFAULT FALSE   -- 是否公开访问
allow_guest BOOLEAN DEFAULT FALSE -- 是否允许访客访问
require_auth BOOLEAN DEFAULT TRUE -- 是否需要认证

-- 新约束
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
UNIQUE(user_id, mount_path)
```

## 🌐 API 端点

### 用户存储管理 API

#### 列出用户存储配置
```bash
GET /api/user/storages?page=1&per_page=20
Authorization: Bearer <user_token>
```

#### 创建用户存储
```bash
POST /api/user/storages/create
Authorization: Bearer <user_token>
Content-Type: application/json

{
  "mount_path": "/my-photos",
  "driver": "Virtual",
  "order_index": 0,
  "is_public": false,
  "allow_guest": false,
  "require_auth": true,
  "addition": "{\"files\":\"[{\\\"name\\\":\\\"photo1.jpg\\\",\\\"size\\\":2048,\\\"is_dir\\\":false,\\\"modified\\\":\\\"2023-01-01 12:00:00\\\"}]\"}",
  "remark": "我的相册"
}
```

#### 更新用户存储
```bash
PUT /api/user/storages/update?id=1
Authorization: Bearer <user_token>
Content-Type: application/json

{
  "mount_path": "/my-photos",
  "is_public": true,
  "allow_guest": true,
  "remark": "公开相册"
}
```

#### 删除用户存储
```bash
DELETE /api/user/storages/delete?id=1
Authorization: Bearer <user_token>
```

#### 测试用户存储连接
```bash
POST /api/user/storages/test
Authorization: Bearer <user_token>
Content-Type: application/json

{
  "driver": "Virtual",
  "mount_path": "/test",
  "addition": "{\"files\":\"[]\"}"
}
```

### 用户文件操作 API

#### 列出用户文件
```bash
GET /api/user/fs/list?path=/my-photos
Authorization: Bearer <user_token>
```

#### 下载文件（支持匿名访问公开存储）
```bash
GET /d/?path=/my-photos/photo1.jpg
# 或
GET /download/?path=/my-photos/photo1.jpg
```

### 管理员 API（保持原有功能）

管理员仍可通过 `/api/admin/storages/*` 端点管理所有用户的存储。

## 🔐 权限访问控制

### 访问级别

1. **用户私有存储**
   - 只有存储所有者可以访问
   - `require_auth = true`

2. **公开存储**
   - 所有人都可以访问
   - `is_public = true`

3. **访客存储**
   - 允许访客（未认证用户）访问
   - `allow_guest = true`

### 权限检查流程

```go
// 1. 检查用户拥有的存储
if storage.UserID == userID {
    return true // 用户自己的存储
}

// 2. 检查公开存储
if storage.IsPublic == true {
    return true // 公开存储
}

// 3. 检查访客权限
if storage.AllowGuest == true && userID == 0 {
    return true // 允许匿名访问
}

return false // 拒绝访问
```

## 💡 使用示例

### 1. 普通用户创建私有存储
```bash
# 用户登录
curl -X POST "https://your-domain.workers.dev/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "password": "password"}'

# 创建私有存储
curl -X POST "https://your-domain.workers.dev/api/user/storages/create" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "mount_path": "/private-docs",
    "driver": "Virtual",
    "is_public": false,
    "addition": "{\"files\":\"[{\\\"name\\\":\\\"secret.txt\\\",\\\"size\\\":1024,\\\"is_dir\\\":false}]\"}"
  }'
```

### 2. 创建公开相册
```bash
curl -X POST "https://your-domain.workers.dev/api/user/storages/create" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "mount_path": "/public-gallery",
    "driver": "Virtual",
    "is_public": true,
    "allow_guest": true,
    "addition": "{\"files\":\"[{\\\"name\\\":\\\"photo1.jpg\\\",\\\"size\\\":2048,\\\"is_dir\\\":false}]\"}"
  }'
```

### 3. 匿名访问公开资源
```bash
# 无需认证即可访问公开存储
curl "https://your-domain.workers.dev/d/?path=/public-gallery/photo1.jpg"
```

## 🏗️ 架构设计

### 驱动管理器层次

```
UserDriverManager
├── User 1 Drivers
│   ├── /private-docs -> Virtual Driver
│   └── /my-photos -> Virtual Driver
├── User 2 Drivers
│   ├── /documents -> S3 Driver
│   └── /backup -> OneDrive Driver
└── Public Storages (cached)
    └── /public-gallery -> Virtual Driver
```

### 服务层架构

```
UserDriverService
├── UserDriverManager (多用户驱动管理)
├── StorageRepository (数据访问层)
├── Permission Check (权限检查)
└── Public Storage Access (公开存储访问)
```

## 🔧 配置示例

### 虚拟驱动配置
```json
{
  "root_folder_path": "/",
  "files": "[
    {\"name\":\"文档\",\"size\":0,\"is_dir\":true,\"modified\":\"2023-01-01 12:00:00\"},
    {\"name\":\"照片\",\"size\":0,\"is_dir\":true,\"modified\":\"2023-01-01 12:00:00\"},
    {\"name\":\"readme.txt\",\"size\":1024,\"is_dir\":false,\"modified\":\"2023-01-01 12:00:00\"}
  ]"
}
```

### 存储权限配置
```json
{
  "mount_path": "/shared-docs",
  "driver": "Virtual",
  "is_public": true,     // 公开访问
  "allow_guest": true,   // 允许访客
  "require_auth": false, // 不需要认证
  "user_id": 1
}
```

## 🚀 部署和迁移

### 数据库迁移
```bash
# 运行用户存储字段迁移
./scripts/migrate.sh -f migrations/0008_add_user_storage_fields.sql

# 或者运行所有迁移
./scripts/migrate.sh -e production
```

### 环境变量
无需额外环境变量，使用现有的：
- `JWT_SECRET`: JWT密钥
- `DB`: Cloudflare D1数据库绑定

## ⚠️ 注意事项

1. **向后兼容性**: 现有存储会自动分配给第一个用户（通常是管理员）
2. **权限检查**: 所有文件访问都会进行权限验证
3. **性能优化**: 用户驱动按需加载，避免内存浪费
4. **安全考虑**: 用户只能管理自己的存储，管理员可以管理所有存储

## 🎯 使用场景

### 1. **个人云存储**
- 用户创建私有存储挂载点
- 管理个人文件和文档
- 支持多种存储后端

### 2. **团队协作**
- 创建公开存储供团队访问
- 分享文件给访客用户
- 不同用户管理不同项目存储

### 3. **内容分发**
- 公开存储作为CDN使用
- 匿名访问支持
- 高性能文件分发

## 📈 性能特性

- **懒加载**: 用户驱动仅在需要时加载
- **内存优化**: 按用户隔离驱动实例
- **缓存策略**: 公开存储结果缓存
- **并发安全**: 线程安全的驱动管理

OpenList Workers 的多用户驱动管理系统为企业和个人用户提供了完整的云存储解决方案，支持从私有存储到公开分享的各种使用场景。 