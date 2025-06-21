# OpenList Workers - 完整的云存储管理平台

基于 Cloudflare Workers 和 D1 数据库的轻量级云存储管理系统，支持多用户、多存储驱动配置、完整的文件系统操作、强大的离线下载功能和JWT认证系统。

## 🌟 主要特性

### 核心功能
- **JWT 认证系统**: 完整的用户注册、登录、权限管理
- **多用户支持**: 每个用户拥有独立的驱动配置和文件空间
- **多驱动支持**: 支持本地存储、S3、阿里云盘、OneDrive、Google Drive 等
- **完整文件系统**: 支持文件和目录的增删改查、上传下载等操作
- **离线下载**: 支持 Aria2、qBittorrent、Transmission、115、PikPak、Thunder 等多种下载工具
- **D1 数据库**: 持久化存储用户和驱动配置数据
- **开发友好**: 支持开发和生产环境分离

### 认证特性
- **JWT Token 认证**: 基于 JWT 的无状态认证机制
- **用户注册**: 自助注册功能，支持用户名和密码验证
- **安全登录**: 密码哈希存储，登录状态管理
- **权限控制**: 基于角色的访问控制（RBAC）
- **多种Token传递**: 支持 Authorization 头和查询参数
- **Token过期管理**: 24小时自动过期，安全可靠

### 离线下载特性
- **多工具支持**: Aria2、qBittorrent、Transmission、115 云盘、PikPak、迅雷
- **任务管理**: 创建、查询、更新、删除离线下载任务
- **进度跟踪**: 实时更新下载进度和状态
- **云盘集成**: 支持 115、PikPak、Thunder 等云盘的离线下载功能
- **用户隔离**: 每个用户的下载配置和任务完全独立

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

### 用户注册和登录
```bash
# 注册新用户
curl -X POST http://localhost:8787/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "myuser",
    "password": "mypassword",
    "base_path": "/home/myuser"
  }'

# 用户登录
curl -X POST http://localhost:8787/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "myuser",
    "password": "mypassword"
  }'
```

## 📚 API 文档

### 认证 API

#### 用户注册
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "myuser",
  "password": "mypassword",
  "base_path": "/home/myuser"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "User registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "myuser",
      "role": 0,
      "base_path": "/home/myuser"
    }
  }
}
```

#### 用户登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "myuser",
  "password": "mypassword"
}
```

#### 获取当前用户信息
```http
GET /api/auth/me
Authorization: Bearer <token>
```

#### 用户登出
```http
POST /api/auth/logout
Authorization: Bearer <token>
```

### 用户管理 API（需要管理员权限）

#### 获取用户列表
```http
GET /api/admin/user/list?page=1&per_page=20
Authorization: Bearer <admin_token>
```

#### 创建用户
```http
POST /api/admin/user/create
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "newpassword",
  "role": 0,
  "base_path": "/home/newuser"
}
```

### 驱动配置 API（需要认证）

#### 获取用户驱动配置
```http
GET /api/drivers
Authorization: Bearer <token>
```

#### 创建驱动配置
```http
POST /api/user/driver/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "MyS3",
  "display_name": "我的S3存储",
  "description": "个人S3存储配置",
  "config": "{\"bucket\": \"my-bucket\", \"region\": \"us-east-1\"}",
  "icon": "cloud",
  "enabled": true
}
```

### 文件系统 API（需要认证）

#### 列出文件和目录
```http
GET /api/fs/list?path=/&page=1&per_page=20
Authorization: Bearer <token>
```

#### 创建目录
```http
POST /api/fs/mkdir
Authorization: Bearer <token>
Content-Type: application/x-www-form-urlencoded

path=/&dir_name=new_folder
```

#### 上传文件
```http
PUT /api/fs/upload?path=/&filename=test.txt
Authorization: Bearer <token>
Content-Type: application/octet-stream

[文件内容]
```

#### 下载文件
```http
GET /d/?path=/test.txt
Authorization: Bearer <token>
```

### 离线下载 API（需要认证）

#### 获取支持的下载工具
```http
GET /api/offline_download_tools
Authorization: Bearer <token>
```

#### 获取用户离线下载配置
```http
GET /api/user/offline_download/configs
Authorization: Bearer <token>
```

#### 配置 Aria2 下载器
```http
POST /api/admin/setting/set_aria2
Authorization: Bearer <token>
Content-Type: application/json

{
  "uri": "http://localhost:6800/jsonrpc",
  "secret": "my_secret_token"
}
```

#### 添加离线下载任务
```http
POST /api/user/offline_download/add_task
Authorization: Bearer <token>
Content-Type: application/json

{
  "urls": [
    "http://example.com/file.zip",
    "magnet:?xt=urn:btih:example123456789"
  ],
  "config_id": 1,
  "dst_path": "/downloads",
  "tool": "aria2",
  "delete_policy": "keep"
}
```

#### 获取离线下载任务列表
```http
GET /api/user/offline_download/tasks?page=1&per_page=20
Authorization: Bearer <token>
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

## 🔐 认证机制

### JWT Token 结构
```json
{
  "user_id": 1,
  "username": "myuser",
  "role": 0,
  "exp": 1703097600,
  "iat": 1703011200
}
```

### Token 传递方式

1. **Authorization Header（推荐）**:
   ```http
   Authorization: Bearer <token>
   ```

2. **Authorization Header（简化）**:
   ```http
   Authorization: <token>
   ```

3. **查询参数**:
   ```http
   GET /api/auth/me?token=<token>
   ```

### 权限级别
- **0 - GENERAL**: 普通用户，只能访问自己的资源
- **1 - GUEST**: 访客用户（通常被禁用）
- **2 - ADMIN**: 管理员用户，可以访问所有管理功能

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

**offline_download_configs 表**:
```sql
CREATE TABLE offline_download_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    tool_name TEXT NOT NULL,
    config TEXT,
    temp_dir_path TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created TEXT,
    modified TEXT,
    UNIQUE(user_id, tool_name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**offline_download_tasks 表**:
```sql
CREATE TABLE offline_download_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    config_id INTEGER NOT NULL,
    urls TEXT NOT NULL,
    dst_path TEXT NOT NULL,
    tool TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    progress INTEGER NOT NULL DEFAULT 0,
    delete_policy TEXT,
    error TEXT,
    created TEXT,
    updated TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (config_id) REFERENCES driver_configs(id) ON DELETE CASCADE
);
```

### JWT 配置
```go
const (
    JWT_SECRET     = "openlist-workers-secret-key-2024"
    JWT_EXPIRATION = 24 * time.Hour // 24小时过期
)
```

## 🧪 测试

### 运行认证功能测试
```bash
./test_auth_api.sh
```

### 运行离线下载功能测试
```bash
./test_offline_download_api.sh
```

### 运行文件系统功能测试
```bash
./test_filesystem_api.sh
```

测试内容包括：
- 用户注册和登录
- JWT Token 验证
- 权限控制测试
- 文件系统操作
- 离线下载功能
- 错误处理测试

## 🔒 安全性

### 认证安全
- JWT Token 24小时自动过期
- 密码使用 SHA256 哈希 + 盐值存储
- 用户数据完全隔离
- 基于角色的权限控制

### 数据隔离
- 用户级别的配置隔离
- 任务权限验证
- 驱动配置验证

### 错误处理
- 参数验证
- 权限检查
- 异常捕获
- 安全的错误信息返回

## 📈 性能特点

### 认证优化
- JWT 无状态设计
- 内存缓存用户信息
- 快速权限验证

### 内存优化
- 驱动实例缓存
- 配置缓存机制
- 任务状态缓存

### 并发支持
- 多用户并发操作
- 多任务并行处理
- 无状态设计

## 🚀 部署说明

### 开发环境
1. 启动开发服务器：`wrangler dev`
2. 初始化系统：`curl -X POST http://localhost:8787/init`
3. 注册用户：`curl -X POST http://localhost:8787/api/auth/register -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}'`
4. 运行测试脚本：`./test_auth_api.sh`

### 生产环境
1. 配置 D1 数据库
2. 更新 JWT_SECRET 为安全的密钥
3. 部署到 Cloudflare Workers
4. 配置环境变量和权限

### 安全建议
1. 在生产环境中修改默认的 JWT_SECRET
2. 设置适当的 CORS 策略
3. 启用 HTTPS
4. 定期更新用户密码
5. 监控异常访问

## 🔮 未来计划

- [ ] OAuth2 第三方登录支持
- [ ] 2FA 双因素认证
- [ ] 用户权限细粒度控制
- [ ] API 访问速率限制
- [ ] 审计日志功能
- [ ] 多租户支持
- [ ] WebSocket 实时通信
- [ ] 移动端适配

## 💡 使用示例

### 完整的用户流程示例

1. **用户注册**:
```bash
curl -X POST http://localhost:8787/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"alice123","base_path":"/home/alice"}'
```

2. **用户登录并获取Token**:
```bash
TOKEN=$(curl -s -X POST http://localhost:8787/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"alice123"}' | jq -r '.data.token')
```

3. **配置云存储驱动**:
```bash
curl -X POST http://localhost:8787/api/user/driver/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MyS3",
    "display_name": "我的S3存储",
    "config": "{\"bucket\":\"my-bucket\",\"region\":\"us-east-1\"}"
  }'
```

4. **使用文件系统功能**:
```bash
# 列出文件
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8787/api/fs/list?path=/"

# 创建目录
curl -X POST http://localhost:8787/api/fs/mkdir \
  -H "Authorization: Bearer $TOKEN" \
  -d "path=/&dir_name=documents"
```

5. **配置离线下载**:
```bash
# 配置Aria2
curl -X POST http://localhost:8787/api/admin/setting/set_aria2 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"uri":"http://localhost:6800/jsonrpc","secret":"mysecret"}'

# 添加下载任务
curl -X POST http://localhost:8787/api/user/offline_download/add_task \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "urls":["http://example.com/file.zip"],
    "config_id":1,
    "dst_path":"/downloads",
    "tool":"aria2"
  }'
```

## 📖 常见问题

### Q: 如何重置用户密码？
A: 目前需要管理员通过 `/api/admin/user/update` 接口重置。

### Q: Token 过期后如何处理？
A: 用户需要重新登录获取新的 Token。

### Q: 如何添加新的云存储驱动？
A: 通过 `/api/user/driver/create` 接口添加新的驱动配置。

### Q: 支持哪些文件操作？
A: 支持列表、创建目录、重命名、移动、复制、删除、上传、下载等操作。

### Q: 离线下载支持哪些协议？
A: 支持 HTTP/HTTPS、FTP、BitTorrent/磁力链接等多种协议。