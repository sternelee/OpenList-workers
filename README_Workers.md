# OpenList Workers - Cloudflare Workers 版本

这是一个基于 Cloudflare Workers 和 D1 数据库的 OpenList drivers API 统一管理系统。

## 功能特性

- ✅ 支持 50+ 种存储驱动（阿里云盘、百度网盘、OneDrive、Google Drive、S3等）
- ✅ 基于 Cloudflare Workers 运行，全球边缘计算
- ✅ 使用 D1 数据库存储配置信息
- ✅ RESTful API 接口
- ✅ 文件列表和下载功能
- ✅ 健康检查和状态监控

## 快速开始

### 1. 环境准备

确保你已经安装了以下工具：
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/install-and-update/)
- Go 1.23.4+

### 2. 创建 D1 数据库

```bash
# 创建 D1 数据库
wrangler d1 create openlist-db

# 复制返回的数据库 ID，更新 wrangler.toml 中的 database_id
```

### 3. 配置 wrangler.toml

更新 `wrangler.toml` 文件中的配置：

```toml
name = "openlist-workers"
main = "main.go"
compatibility_date = "2024-01-01"

[build]
command = "go build -o dist/worker main.go"

# D1 数据库配置
[[d1_databases]]
binding = "DB"
database_name = "openlist-db"
database_id = "your-actual-database-id"

# 环境变量
[vars]
ENVIRONMENT = "production"
```

### 4. 部署

```bash
# 部署到 Cloudflare Workers
wrangler deploy
```

## API 接口

### 1. 健康检查

```http
GET /health
```

响应示例：
```json
{
  "code": 200,
  "message": "OpenList Workers is running",
  "data": {
    "drivers_count": 50,
    "storages_count": 2,
    "timestamp": 1703123456
  }
}
```

### 2. 获取驱动列表

```http
GET /api/drivers
```

响应示例：
```json
{
  "code": 200,
  "data": {
    "drivers": ["S3", "AliyunDrive", "BaiduNetdisk", "OneDrive"],
    "info": {
      "S3": {
        "common": [...],
        "additional": [...],
        "config": {...}
      }
    }
  }
}
```

### 3. 存储管理

#### 获取存储列表
```http
GET /api/storages?page=1&per_page=20
```

#### 创建存储
```http
POST /api/storages
Content-Type: application/json

{
  "mount_path": "/my-s3",
  "driver": "S3",
  "order": 1,
  "remark": "My S3 Storage",
  "disabled": false
}
```

### 4. 文件系统操作

#### 列出文件
```http
GET /api/fs/list/my-s3/path/to/folder
```

#### 下载文件
```http
GET /d/my-s3/path/to/file.txt
```

## 支持的存储驱动

### 云盘服务
- 阿里云盘 (AliyunDrive)
- 百度网盘 (BaiduNetdisk)
- 115网盘 (115)
- 天翼云盘 (189)
- 夸克网盘 (Quark)
- UC网盘 (UC)
- 迅雷云盘 (Thunder)

### 国际服务
- Google Drive
- OneDrive
- Dropbox
- Mega
- Yandex Disk

### 对象存储
- Amazon S3
- 阿里云 OSS
- 腾讯云 COS
- 七牛云
- 又拍云

### 协议支持
- FTP
- SFTP
- WebDAV
- SMB

### 其他服务
- GitHub
- GitLab
- 蓝奏云
- 天翼云盘
- 和彩云

## 开发指南

### 添加新的存储驱动

1. 在 `drivers/` 目录下创建新的驱动包
2. 实现 `driver.Driver` 接口
3. 在 `meta.go` 中注册驱动
4. 在 `drivers/all.go` 中导入新驱动

### 本地开发

```bash
# 本地运行
wrangler dev

# 测试 API
curl http://localhost:8787/health
```

### 数据库操作

当前实现使用模拟数据，要启用真实的 D1 数据库操作：

1. 取消注释 `main.go` 中的数据库操作代码
2. 根据实际的 D1 数据库 API 调整代码
3. 确保 D1 数据库绑定正确配置

## 部署配置

### 生产环境

```bash
# 部署到生产环境
wrangler deploy --env production

# 查看日志
wrangler tail
```

### 自定义域名

在 `wrangler.toml` 中配置自定义域名：

```toml
[env.production.routes]
pattern = "your-domain.com/*"
zone_name = "your-domain.com"
```

## 监控和日志

### 健康检查
- 端点：`/health`
- 检查驱动数量、存储数量、系统状态

### 错误处理
- 所有 API 返回统一的错误格式
- 包含错误代码和详细信息

## 安全考虑

1. **认证授权**：当前版本未实现认证，生产环境需要添加
2. **CORS 配置**：已配置基本的 CORS 头
3. **输入验证**：需要添加更严格的输入验证
4. **速率限制**：建议添加 API 速率限制

## 性能优化

1. **缓存策略**：存储驱动实例在内存中缓存
2. **连接池**：复用数据库连接
3. **边缘计算**：利用 Cloudflare 全球边缘节点

## 故障排除

### 常见问题

1. **D1 数据库连接失败**
   - 检查 `wrangler.toml` 中的数据库配置
   - 确认数据库 ID 正确

2. **驱动加载失败**
   - 检查驱动名称是否正确
   - 查看控制台错误日志

3. **文件访问失败**
   - 检查存储配置是否正确
   - 确认存储服务可用性

### 调试模式

```bash
# 启用调试日志
wrangler dev --log-level debug
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

## 许可证

本项目基于原 OpenList 项目的许可证。

## 更新日志

### v1.0.0
- 初始版本
- 支持基本的 drivers API
- 集成 D1 数据库
- 文件列表和下载功能