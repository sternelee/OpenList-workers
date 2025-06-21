# OpenList Workers - 完整的云存储管理平台

基于 Cloudflare Workers 和 D1 数据库的轻量级云存储管理系统，支持多用户、多存储驱动配置和完整的文件系统操作。

## 🌟 主要特性

### 核心功能
- **多用户支持**: 每个用户拥有独立的驱动配置和文件空间
- **多驱动支持**: 支持本地存储、S3、阿里云盘、OneDrive、Google Drive 等
- **完整文件系统**: 支持文件和目录的增删改查、上传下载等操作
- **D1 数据库**: 持久化存储用户和驱动配置数据
- **开发友好**: 支持开发和生产环境分离

### 架构优势
- **无服务器**: 基于 Cloudflare Workers，自动扩缩容
- **高性能**: 全球 CDN 加速，毫秒级响应
- **低成本**: 按需付费，小规模使用几乎免费
- **安全**: 用户数据隔离，配置独立管理

## 🚀 快速开始

### 环境准备
```bash
# 克隆项目
git clone https://github.com/yourusername/OpenList-workers.git
cd OpenList-workers

# 安装依赖
npm install
```

### 本地开发
```bash
# 启动开发服务器
wrangler dev

# 访问 http://localhost:8787
```

### 初始化系统
```bash
# 初始化数据库和默认数据
curl -X POST http://localhost:8787/init
```

## 📚 API 文档

### 用户管理 API

#### 获取用户列表
```http
GET /api/admin/user/list?page=1&per_page=20
```

#### 创建用户
```http
POST /api/admin/user/create
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123",
  "role": 2,
  "base_path": "/",
  "permission": 0x30FF
}
```

#### 更新用户
```http
POST /api/admin/user/update
Content-Type: application/json

{
  "id": 1,
  "username": "updateduser",
  "password": "newpassword",
  "disabled": false
}
```

#### 删除用户
```http
POST /api/admin/user/delete?id=1
```

### 驱动配置管理 API

#### 获取用户驱动配置列表
```http
GET /api/drivers?user_id=1&enabled=true
```

#### 创建驱动配置
```http
POST /api/user/driver/create
Content-Type: application/json

{
  "name": "MyS3",
  "display_name": "我的 S3 存储",
  "description": "私人 S3 存储配置",
  "config": "{\"bucket\":\"my-bucket\",\"region\":\"us-east-1\",\"access_key_id\":\"xxx\",\"secret_access_key\":\"xxx\"}",
  "icon": "cloud",
  "enabled": true,
  "order": 1
}
```

#### 更新驱动配置
```http
POST /api/user/driver/update
Content-Type: application/json

{
  "id": 1,
  "name": "MyS3",
  "display_name": "更新的 S3 存储",
  "config": "{\"bucket\":\"new-bucket\",\"region\":\"us-east-1\"}",
  "enabled": true
}
```

#### 删除驱动配置
```http
POST /api/user/driver/delete?id=1
```

#### 启用/禁用驱动配置
```http
POST /api/user/driver/enable?id=1
POST /api/user/driver/disable?id=1
```

### 文件系统 API

#### 列出文件和目录
```http
GET /api/fs/list?user_id=1&config_id=1&path=/&page=1&per_page=20
```

#### 获取文件信息
```http
GET /api/fs/get?user_id=1&config_id=1&path=/file.txt
```

#### 创建目录
```http
POST /api/fs/mkdir
Content-Type: application/x-www-form-urlencoded

user_id=1&config_id=1&path=/&dir_name=new_folder
```

#### 重命名文件/目录
```http
POST /api/fs/rename
Content-Type: application/x-www-form-urlencoded

user_id=1&config_id=1&path=/old_name.txt&new_name=new_name.txt
```

#### 移动文件/目录
```http
POST /api/fs/move
Content-Type: application/x-www-form-urlencoded

user_id=1&config_id=1&path=/source/file.txt&dst_path=/destination/
```

#### 复制文件/目录
```http
POST /api/fs/copy
Content-Type: application/x-www-form-urlencoded

user_id=1&config_id=1&path=/source/file.txt&dst_path=/destination/
```

#### 删除文件/目录
```http
POST /api/fs/remove
Content-Type: application/x-www-form-urlencoded

user_id=1&config_id=1&path=/file.txt
```

#### 上传文件
```http
PUT /api/fs/upload?user_id=1&config_id=1&path=/folder&filename=upload.txt
Content-Type: application/octet-stream

[文件内容]
```

#### 下载文件
```http
GET /d/?user_id=1&config_id=1&path=/file.txt
```

### 系统 API

#### 健康检查
```http
GET /health
```

#### 系统初始化
```http
POST /init
```

## 🛠️ 配置说明

### 数据库配置

#### D1 数据库表结构

**users 表**:
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    pwd_hash TEXT NOT NULL,
    salt TEXT NOT NULL,
    base_path TEXT NOT NULL DEFAULT '/',
    role INTEGER NOT NULL DEFAULT 2,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    permission INTEGER NOT NULL DEFAULT 0,
    sso_id TEXT,
    otp_secret TEXT,
    authn TEXT
);
```

**driver_configs 表**:
```sql
CREATE TABLE driver_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    display_name TEXT,
    description TEXT,
    config TEXT,
    icon TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    order_num INTEGER NOT NULL DEFAULT 0,
    created TEXT,
    modified TEXT,
    UNIQUE(user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 驱动配置示例

#### 本地存储
```json
{
  "root_folder_path": "/data"
}
```

#### Amazon S3
```json
{
  "bucket": "my-bucket",
  "region": "us-east-1",
  "access_key_id": "AKIAIOSFODNN7EXAMPLE",
  "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  "endpoint": "https://s3.amazonaws.com"
}
```

#### 阿里云盘
```json
{
  "refresh_token": "your_refresh_token",
  "root_folder_id": "root"
}
```

#### OneDrive
```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "redirect_uri": "http://localhost:8787/callback"
}
```

#### Google Drive
```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "redirect_uri": "http://localhost:8787/callback"
}
```

## 🧪 测试

### 运行测试脚本
```bash
# 基础 API 测试
chmod +x test_d1_api.sh
./test_d1_api.sh

# 文件系统 API 测试
chmod +x test_filesystem_api.sh
./test_filesystem_api.sh
```

### 测试覆盖范围
- ✅ 用户管理 CRUD 操作
- ✅ 驱动配置管理
- ✅ 文件系统基本操作
- ✅ 权限验证
- ✅ 错误处理
- ✅ 多用户隔离

## 🚢 部署

### Cloudflare Workers 部署

1. **配置 wrangler.toml**:
```toml
name = "openlist-workers"
main = "main.go"
compatibility_date = "2024-01-01"

[[d1_databases]]
binding = "DB"
database_name = "openlist-db"
database_id = "your-database-id"
```

2. **创建 D1 数据库**:
```bash
wrangler d1 create openlist-db
```

3. **执行数据库迁移**:
```bash
wrangler d1 execute openlist-db --file=schema.sql
```

4. **部署应用**:
```bash
wrangler deploy
```

### 环境变量
```toml
[vars]
ENVIRONMENT = "production"
```

## 🔒 安全特性

### 用户隔离
- 每个用户只能访问自己的驱动配置
- 文件操作限制在用户配置的驱动范围内
- 数据库层面的外键约束确保数据一致性

### 权限控制
- 管理员用户可以管理所有用户
- 普通用户只能管理自己的配置
- 支持角色基础的权限控制

### 数据保护
- 密码使用 salt + hash 存储
- 敏感配置信息存储在 D1 数据库中
- 支持 2FA 认证（预留接口）

## 📊 性能特性

### 缓存机制
- 驱动实例缓存，避免重复初始化
- 用户配置内存缓存
- 智能缓存失效机制

### 资源优化
- 按需加载驱动
- 连接池复用
- 最小化内存使用

## 🔧 开发指南

### 项目结构
```
OpenList-workers/
├── main.go                    # 主应用程序
├── d1_database.go            # 生产环境数据库
├── d1_database_dev.go        # 开发环境数据库
├── test_d1_api.sh           # API 测试脚本
├── test_filesystem_api.sh    # 文件系统测试脚本
├── wrangler.toml            # Cloudflare Workers 配置
├── README_D1_Complete.md    # 完整文档
├── README_FileSystem.md     # 文件系统文档
└── README_Workers.md        # Workers 文档
```

### 添加新驱动
1. 在 `initDefaultData()` 中添加驱动配置
2. 确保驱动名称与 OpenList 支持的驱动匹配
3. 提供正确的配置 JSON 格式
4. 测试驱动兼容性

### 调试指南
```bash
# 查看日志
wrangler tail

# 本地调试
wrangler dev --local

# 数据库查询
wrangler d1 execute openlist-db --command "SELECT * FROM users;"
```

## 🤝 贡献指南

### 提交代码
1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 创建 Pull Request

### 报告问题
- 使用 GitHub Issues
- 提供详细的错误信息和重现步骤
- 包含环境信息

## 📄 许可证

MIT License - 详见 LICENSE 文件

## 🔗 相关链接

- [Cloudflare Workers 文档](https://developers.cloudflare.com/workers/)
- [D1 数据库文档](https://developers.cloudflare.com/d1/)
- [OpenList 项目](https://github.com/OpenListTeam/OpenList)

## 📞 支持

- GitHub Issues: 技术问题和 bug 报告
- Discussions: 使用问题和建议
- Email: 商业支持和合作

---

**OpenList Workers** - 让云存储管理变得简单高效！ 🚀