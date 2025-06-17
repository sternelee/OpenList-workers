# OpenList 模型迁移到 Cloudflare D1 数据库

## 概述

本文档记录了将 `internal/model` 目录下的模型迁移到 Cloudflare D1 数据库的过程和使用方法。

## 迁移的模型

### 主要数据模型

1. **User** - 用户模型，包含认证、权限等信息
2. **Storage** - 存储配置模型
3. **SettingItem** - 系统设置模型
4. **Meta** - 路径元数据模型
5. **SearchNode** - 搜索节点模型
6. **TaskItem** - 任务项模型
7. **SSHPublicKey** - SSH公钥模型

### 未迁移的模型

以下模型主要是接口定义和内存对象，不需要在数据库中持久化：
- `file.go` - 文件接口定义
- `obj.go` - 对象接口定义
- `object.go` - 对象实现
- `args.go` - 参数模型
- `req.go` - 请求模型
- `archive.go` - 归档模型

## 数据库迁移文件

创建了以下迁移文件：

1. `migrations/0001_create_users_table.sql` - 用户表（已更新）
2. `migrations/0002_create_storages_table.sql` - 存储配置表
3. `migrations/0003_create_settings_table.sql` - 系统设置表
4. `migrations/0004_create_metas_table.sql` - 路径元数据表
5. `migrations/0005_create_search_nodes_table.sql` - 搜索节点表
6. `migrations/0006_create_tasks_table.sql` - 任务表
7. `migrations/0007_create_ssh_keys_table.sql` - SSH密钥表

## Go 模型文件

在 `workers/models/` 目录下创建了以下模型文件：

- `user.go` - 用户模型（已更新）
- `storage.go` - 存储配置模型
- `setting.go` - 系统设置模型
- `meta.go` - 路径元数据模型
- `search.go` - 搜索模型
- `task.go` - 任务模型
- `ssh_key.go` - SSH密钥模型

## 数据库操作接口

在 `workers/db/` 目录下创建了完整的数据库操作层：

### 接口定义
- `repositories.go` - 仓库接口定义

### 仓库实现
- `user_repository.go` - 用户仓库实现
- `storage_repository.go` - 存储仓库实现
- `setting_repository.go` - 设置仓库实现
- `meta_repository.go` - 元数据仓库实现
- `search_node_repository.go` - 搜索节点仓库实现
- `task_repository.go` - 任务仓库实现
- `ssh_key_repository.go` - SSH密钥仓库实现

## 主要变更

### 用户模型变更
- 将 `PasswordHash` 字段改为 `PwdHash`
- 添加 `PwdTS` 字段记录密码时间戳
- 添加数据库标签 `db:"field_name"`
- 完善权限检查方法

### 通用变更
- 所有模型添加 `CreatedAt` 和 `UpdatedAt` 时间字段
- 添加适当的数据库索引
- 统一使用 `context.Context` 进行数据库操作
- 实现完整的 CRUD 操作

## 使用方法

### 1. 初始化数据库连接

```go
import (
    "github.com/sternelee/OpenList-workers/workers/db"
    "github.com/sternelee/OpenList-workers/workers/models"
)

// 创建仓库实例
repos := db.NewRepositories(sqlDB)
```

### 2. 用户操作示例

```go
// 创建用户
user := &models.User{
    Username: "testuser",
    Role:     models.GENERAL,
    Permission: 0,
}
user.SetPassword("password123")
err := repos.User.Create(ctx, user)

// 根据用户名查询
user, err := repos.User.GetByUsername(ctx, "testuser")

// 验证密码
err := user.ValidateRawPassword("password123")
```

### 3. 存储配置操作示例

```go
// 创建存储配置
storage := &models.Storage{
    MountPath: "/example",
    Driver:    "local",
    Status:    "work",
}
err := repos.Storage.Create(ctx, storage)

// 获取启用的存储
storages, err := repos.Storage.ListEnabled(ctx)
```

### 4. 设置操作示例

```go
// 设置配置值
err := repos.Setting.SetValue(ctx, "site_title", "My OpenList")

// 获取配置
setting, err := repos.Setting.GetByKey(ctx, "site_title")
```

## 数据库迁移执行

使用 Cloudflare Workers 的数据库迁移功能：

```bash
# 应用迁移
npx wrangler d1 migrations apply openlist-db

# 或者本地开发环境
npx wrangler d1 migrations apply openlist-db --local
```

## 注意事项

1. **时间戳处理**：D1 使用 SQLite，时间戳字段使用 `DATETIME` 类型
2. **外键约束**：D1 支持外键约束，SSH密钥表引用用户表
3. **索引优化**：为常用查询字段创建了索引
4. **事务支持**：建议在复杂操作中使用事务
5. **连接池**：Cloudflare Workers 自动管理 D1 连接

## 性能考虑

1. **查询优化**：使用适当的索引和查询条件
2. **分页查询**：实现了 `limit` 和 `offset` 分页
3. **批量操作**：考虑实现批量插入和更新方法
4. **缓存策略**：对于频繁访问的设置项可以考虑缓存

## 扩展性

数据库操作层设计采用了仓库模式，便于：
- 添加新的数据模型
- 扩展查询方法
- 实现缓存层
- 进行单元测试

这个迁移为 OpenList 项目提供了一个稳定、可扩展的数据持久化层。 