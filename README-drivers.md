# OpenList Workers 驱动系统使用指南

## 概述

OpenList Workers 现在支持完整的驱动系统，可以连接和管理多种存储后端。系统提供了灵活的架构，支持多种存储驱动，包括云存储、本地存储等。

## 核心功能

### 1. 驱动管理
- 动态驱动注册和加载
- 支持多种存储后端
- 热重载配置
- 驱动状态监控

### 2. 存储管理
- 存储配置的增删改查
- 存储状态管理（启用/禁用）
- 存储测试和验证
- 挂载路径管理

### 3. 文件操作
- 文件列表
- 文件下载
- 文件信息获取
- 路径解析和转发

## API 端点

### 驱动相关 API

#### 获取所有驱动信息
```bash
GET /api/drivers/info
```

#### 获取驱动名称列表
```bash
GET /api/drivers
```

#### 获取特定驱动信息
```bash
GET /api/drivers/info/?driver=Virtual
```

### 存储管理 API（需要管理员权限）

#### 列出所有存储
```bash
GET /api/admin/storages?page=1&per_page=20
Authorization: Bearer <admin_token>
```

#### 创建存储
```bash
POST /api/admin/storages/create
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "mount_path": "/test",
  "driver": "Virtual",
  "order_index": 0,
  "addition": "{\"files\":\"[{\\\"name\\\":\\\"test.txt\\\",\\\"size\\\":1024,\\\"is_dir\\\":false,\\\"modified\\\":\\\"2023-01-01 12:00:00\\\"}]\"}",
  "remark": "测试虚拟存储"
}
```

#### 测试存储连接
```bash
POST /api/admin/storages/test
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "driver": "Virtual",
  "mount_path": "/test",
  "addition": "{\"files\":\"[{\\\"name\\\":\\\"test.txt\\\",\\\"size\\\":1024,\\\"is_dir\\\":false,\\\"modified\\\":\\\"2023-01-01 12:00:00\\\"}]\"}"
}
```

#### 启用/禁用存储
```bash
POST /api/admin/storages/enable?id=1
POST /api/admin/storages/disable?id=1
Authorization: Bearer <admin_token>
```

### 文件操作 API

#### 列出文件
```bash
GET /api/fs/list?path=/test
Authorization: Bearer <token>
```

#### 获取文件信息
```bash
GET /api/fs/get?path=/test/file.txt
Authorization: Bearer <token>
```

#### 下载文件
```bash
GET /d/?path=/test/file.txt
GET /download/?path=/test/file.txt
```

## 驱动类型

### Virtual 驱动
虚拟驱动用于测试和演示目的，支持通过JSON配置虚拟文件列表。

**配置示例：**
```json
{
  "root_folder_path": "/",
  "files": "[{\"name\":\"文档.txt\",\"size\":1024,\"is_dir\":false,\"modified\":\"2023-01-01 12:00:00\"},{\"name\":\"图片\",\"size\":0,\"is_dir\":true,\"modified\":\"2023-01-01 12:00:00\"}]"
}
```

### 扩展驱动
系统设计为可扩展架构，可以轻松添加新的驱动类型：
- 阿里云盘
- OneDrive
- Google Drive
- S3兼容存储
- 本地存储
- FTP/SFTP
- WebDAV
- 等等

## 开发指南

### 添加新驱动

1. 创建驱动包
```go
// workers/drivers/mydrive/driver.go
package mydrive

import (
    "context"
    "github.com/sternelee/OpenList-workers/workers/drivers"
)

type MyDrive struct {
    drivers.BaseDriver
    Addition
}

type Addition struct {
    drivers.RootPath
    APIKey string `json:"api_key" required:"true"`
    Secret string `json:"secret" required:"true"`
}

func (d *MyDrive) Config() drivers.DriverConfig {
    return drivers.DriverConfig{
        Name:        "MyDrive",
        DefaultRoot: "/",
    }
}

// 实现其他必需的接口方法...
```

2. 注册驱动
```go
// workers/drivers/mydrive/meta.go
func init() {
    drivers.RegisterDriver(func() drivers.Driver {
        return &MyDrive{}
    })
}
```

3. 导入驱动
```go
// main.go
import _ "github.com/sternelee/OpenList-workers/workers/drivers/mydrive"
```

### 数据库集成

驱动系统完全集成了 Cloudflare D1 数据库：
- 存储配置持久化
- 状态管理
- 权限控制
- 审计日志

## 部署说明

### 环境变量
- `JWT_SECRET`: JWT 密钥
- `DB`: Cloudflare D1 数据库绑定

### 数据库迁移
运行迁移脚本创建必要的数据表：
```bash
./scripts/migrate.sh -e production
```

### Wrangler 配置
确保 `wrangler.toml` 包含必要的配置：
```toml
[[d1_databases]]
binding = "DB"
database_name = "openlist-db"
database_id = "your-d1-database-id"
```

## 安全考虑

1. **权限控制**: 存储管理需要管理员权限
2. **认证**: 文件操作需要有效的 JWT token
3. **路径验证**: 防止路径遍历攻击
4. **配置加密**: 敏感配置信息应加密存储
5. **访问日志**: 记录所有文件访问操作

## 性能优化

1. **缓存策略**: 支持文件列表缓存
2. **并发控制**: 限制同时操作数量
3. **懒加载**: 按需加载驱动实例
4. **连接池**: 复用数据库连接

## 故障排除

### 常见问题

1. **驱动初始化失败**
   - 检查存储配置是否正确
   - 验证网络连接
   - 查看错误日志

2. **文件访问失败**
   - 确认用户权限
   - 检查挂载路径
   - 验证驱动状态

3. **性能问题**
   - 启用缓存
   - 优化查询
   - 监控资源使用

### 日志查看
```bash
# Cloudflare Workers 日志
wrangler tail

# 健康检查
curl https://your-domain.workers.dev/health
```

## 总结

OpenList Workers 驱动系统提供了完整的存储管理解决方案，支持多种存储后端，具备良好的扩展性和性能。通过标准化的 API 接口，可以轻松集成各种存储服务，为用户提供统一的文件管理体验。 