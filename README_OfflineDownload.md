# OpenList Workers 离线下载功能

基于用户驱动配置的强大离线下载系统，支持多种下载工具和云盘离线下载功能。

## 🎯 功能概述

OpenList Workers 的离线下载功能为用户提供了统一的下载管理平台，支持：

- **传统下载工具**: Aria2、qBittorrent、Transmission
- **云盘离线下载**: 115 云盘、PikPak、迅雷
- **统一任务管理**: 统一的 API 接口管理所有下载任务
- **用户隔离**: 每个用户独立的配置和任务空间

## 🔧 支持的下载工具

### 1. Aria2
- **类型**: HTTP/FTP/BitTorrent 下载器
- **配置**: URI + Secret
- **特点**: 轻量级、高性能、支持多协议

### 2. qBittorrent
- **类型**: BitTorrent 客户端
- **配置**: Web UI URL + 做种时间
- **特点**: 开源、功能丰富、Web 管理界面

### 3. Transmission
- **类型**: BitTorrent 客户端
- **配置**: RPC URI + 做种时间
- **特点**: 简洁、稳定、资源占用少

### 4. 115 云盘
- **类型**: 云盘离线下载
- **配置**: 临时目录路径 + 驱动配置ID
- **特点**: 高速下载、大容量、支持多种格式

### 5. PikPak
- **类型**: 云盘离线下载
- **配置**: 临时目录路径 + 驱动配置ID
- **特点**: 国际化服务、支持多平台

### 6. 迅雷 (Thunder)
- **类型**: 云盘离线下载
- **配置**: 临时目录路径 + 驱动配置ID
- **特点**: 国内优化、下载加速

## 📚 API 使用指南

### 基础配置 API

#### 获取支持的工具列表
```bash
curl -s "http://localhost:8787/api/offline_download_tools"
```

#### 获取用户离线下载配置
```bash
curl -s "http://localhost:8787/api/user/offline_download/configs?user_id=1"
```

### 工具配置 API

#### 配置 Aria2
```bash
curl -X POST "http://localhost:8787/api/admin/setting/set_aria2?user_id=1" \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "http://localhost:6800/jsonrpc",
    "secret": "my_secret_token"
  }'
```

#### 配置 qBittorrent
```bash
curl -X POST "http://localhost:8787/api/admin/setting/set_qbittorrent?user_id=1" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://localhost:8080",
    "seedtime": "60"
  }'
```

#### 配置 115 云盘
```bash
curl -X POST "http://localhost:8787/api/admin/setting/set_115?user_id=1" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_dir_path": "/downloads/115",
    "config_id": 1
  }'
```

### 任务管理 API

#### 创建下载任务
```bash
curl -X POST "http://localhost:8787/api/user/offline_download/add_task?user_id=1" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "http://example.com/file.zip",
      "magnet:?xt=urn:btih:example123456789"
    ],
    "config_id": 1,
    "dst_path": "/downloads",
    "tool": "aria2",
    "delete_policy": "keep"
  }'
```

#### 查询任务列表
```bash
curl -s "http://localhost:8787/api/user/offline_download/tasks?user_id=1&page=1&per_page=20"
```

#### 更新任务状态
```bash
curl -X POST "http://localhost:8787/api/user/offline_download/update_task?user_id=1" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 1,
    "status": "running",
    "progress": 50,
    "error": ""
  }'
```

#### 删除任务
```bash
curl -X POST "http://localhost:8787/api/user/offline_download/delete_task?user_id=1&task_id=1"
```

## 🏗️ 架构设计

### 数据结构

#### 离线下载配置
```go
type OfflineDownloadConfig struct {
    ID           uint   `json:"id"`
    UserID       uint   `json:"user_id"`       // 关联用户ID
    ToolName     string `json:"tool_name"`     // 工具名称
    Config       string `json:"config"`        // JSON配置
    TempDirPath  string `json:"temp_dir_path"` // 临时目录
    Enabled      bool   `json:"enabled"`       // 是否启用
    Created      string `json:"created"`       // 创建时间
    Modified     string `json:"modified"`      // 修改时间
}
```

#### 离线下载任务
```go
type OfflineDownloadTask struct {
    ID           uint   `json:"id"`
    UserID       uint   `json:"user_id"`       // 关联用户ID
    ConfigID     uint   `json:"config_id"`     // 驱动配置ID
    URLs         string `json:"urls"`          // URL列表(JSON)
    DstPath      string `json:"dst_path"`      // 目标路径
    Tool         string `json:"tool"`          // 使用的工具
    Status       string `json:"status"`        // 任务状态
    Progress     int    `json:"progress"`      // 进度百分比
    DeletePolicy string `json:"delete_policy"` // 删除策略
    Error        string `json:"error"`         // 错误信息
    Created      string `json:"created"`       // 创建时间
    Updated      string `json:"updated"`       // 更新时间
}
```

### 状态管理

#### 任务状态
- `pending`: 等待中
- `running`: 运行中
- `completed`: 已完成
- `failed`: 失败

#### 删除策略
- `keep`: 保留文件
- `delete_on_complete`: 完成后删除
- `delete_on_upload`: 上传后删除

### 用户隔离

每个用户拥有：
- 独立的下载工具配置
- 独立的下载任务队列
- 独立的临时目录空间
- 独立的权限控制

## 🛠️ 配置示例

### Aria2 配置
```json
{
  "uri": "http://localhost:6800/jsonrpc",
  "secret": "my_secret_token"
}
```

### qBittorrent 配置
```json
{
  "url": "http://localhost:8080",
  "seedtime": "60"
}
```

### 115 云盘配置
```json
{
  "temp_dir_path": "/downloads/115",
  "config_id": 1
}
```

## 🧪 测试

运行完整的离线下载功能测试：

```bash
./test_offline_download_api.sh
```

测试包括：
- 工具配置测试
- 任务创建测试
- 状态更新测试
- 错误处理测试

## 🔒 安全性

### 数据隔离
- 用户级别的配置隔离
- 任务权限验证
- 驱动配置验证

### 错误处理
- 参数验证
- 权限检查
- 异常捕获

## 📈 性能特点

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
3. 运行测试脚本：`./test_offline_download_api.sh`

### 生产环境
1. 配置 D1 数据库
2. 部署到 Cloudflare Workers
3. 配置环境变量和权限

## 🔮 未来计划

- [ ] 任务调度优化
- [ ] 下载速度限制
- [ ] 多文件批量处理
- [ ] WebSocket 实时状态推送
- [ ] 下载统计报表
- [ ] 更多云盘支持