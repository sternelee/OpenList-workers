# OpenList Cloudflare Workers 部署指南

这是 OpenList 项目的 Cloudflare Workers 版本，使用 D1 数据库存储用户登录信息。

## 功能特性

- 🚀 运行在 Cloudflare Workers 上
- 🗄️ 使用 D1 数据库存储用户数据
- 🔐 JWT 认证系统
- 👥 用户管理功能
- 🛡️ 基于角色的权限控制

## 环境要求

- Node.js (>=16)
- Go (>=1.23)
- TinyGo
- wrangler CLI

## 安装工具

### 1. 安装 wrangler
```bash
npm install -g wrangler
```

### 2. 安装 TinyGo
```bash
# macOS
brew install tinygo

# 或者从官网下载: https://tinygo.org/getting-started/install/
```

### 3. 安装 workers-assets-gen
```bash
make install-tools
```

## 部署步骤

### 1. 克隆项目
```bash
git clone <repository-url>
cd OpenList-workers
```

### 2. 登录 Cloudflare
```bash
wrangler login
```

### 3. 创建 D1 数据库
```bash
make db-create
```

这将创建一个名为 `openlist-db` 的 D1 数据库。复制输出中的 `database_id` 并更新 `wrangler.toml` 文件中的相应字段。

### 4. 更新配置文件
编辑 `wrangler.toml` 文件，将 `database_id` 替换为实际的数据库 ID：

```toml
[[d1_databases]]
binding = "DB"
database_name = "openlist-db"
database_id = "your-actual-database-id-here"
```

### 5. 运行数据库迁移
```bash
# 本地开发环境
make db-migrate-local

# 生产环境
make db-migrate-remote
```

### 6. 本地开发
```bash
make dev
```

访问 `http://localhost:8787/ping` 确认服务正常运行。

### 7. 部署到生产环境
```bash
make deploy
```

## API 端点

### 公共端点
- `GET /ping` - 健康检查

### 认证端点
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出（需要认证）

### 用户端点
- `GET /api/me` - 获取当前用户信息（需要认证）

### 管理员端点
- `GET /api/admin/users` - 获取用户列表（需要管理员权限）

## 使用示例

### 登录
```bash
curl -X POST http://localhost:8787/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}'
```

### 获取当前用户信息
```bash
curl -X GET http://localhost:8787/api/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 获取用户列表（管理员）
```bash
curl -X GET http://localhost:8787/api/admin/users \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

## 默认用户

系统会自动创建一个默认管理员用户：
- 用户名: `admin`
- 密码: `admin`

⚠️ **重要**: 部署到生产环境后，请立即更改默认管理员密码！

## 环境变量

可以在 `wrangler.toml` 中设置以下环境变量：

```toml
[vars]
JWT_SECRET = "your-jwt-secret-here"  # JWT 签名密钥
ENVIRONMENT = "production"           # 环境标识
```

## 数据库操作

### 查询本地数据库
```bash
make db-query-local
```

### 查询远程数据库
```bash
make db-query-remote
```

### 直接执行 SQL
```bash
# 本地
wrangler d1 execute openlist-db --local --command "SELECT * FROM users;"

# 远程
wrangler d1 execute openlist-db --remote --command "SELECT * FROM users;"
```

## 故障排除

### 1. 构建失败
确保已安装 TinyGo 和 workers-assets-gen：
```bash
make install-tools
```

### 2. 数据库连接失败
检查 `wrangler.toml` 中的 `database_id` 是否正确。

### 3. 权限问题
确保已通过 `wrangler login` 登录到 Cloudflare。

## 开发

### 项目结构
```
├── main.go                    # 主入口文件
├── wrangler.toml             # Cloudflare Workers 配置
├── Makefile                  # 构建脚本
├── migrations/               # 数据库迁移文件
│   └── 0001_create_users_table.sql
├── workers/                  # Workers 相关代码
│   ├── auth/                # 认证相关
│   ├── db/                  # 数据库操作
│   ├── handlers/            # HTTP 处理器
│   └── models/              # 数据模型
└── pkg/                     # 工具包
    └── utils/
        └── random/          # 随机字符串生成
```

### 添加新功能
1. 在 `workers/handlers/` 中添加新的处理器
2. 在 `main.go` 中注册路由
3. 如需数据库操作，在 `workers/db/` 中添加相应方法

## 许可证

本项目基于原 OpenList 项目的许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！ 