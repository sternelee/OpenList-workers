# RSS 订阅管理 + 资源搜索 API 文档

## 🗂️ 功能概述

基于 qBittorrent 的设计理念，为 alist 实现了完整的 RSS 订阅管理和资源搜索功能：

- **RSS 订阅管理**: 支持文件夹组织、自动刷新、自动下载规则
- **资源搜索**: 插件化搜索引擎，支持多站点并发搜索
- **自动下载**: 基于关键词和正则表达式的智能过滤
- **多下载工具**: 支持 aria2、qBittorrent、Transmission 和云盘离线下载 (115云盘、PikPak、迅雷网盘)
- **深度集成**: 与 alist 现有的离线下载系统无缝结合

## 📡 RSS 订阅管理 API

### 文件夹管理

#### 获取文件夹列表
```http
GET /api/admin/rss/folders
```

#### 创建文件夹
```http
POST /api/admin/rss/folders
Content-Type: application/json

{
  "name": "动漫",
  "parent_path": "/电影"
}
```

#### 删除文件夹
```http
DELETE /api/admin/rss/folders/{id}
```

### 订阅源管理

#### 获取所有订阅
```http
GET /api/admin/rss/feeds
```

**响应示例:**
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "uid": "feed-uuid-123",
      "name": "DMHY 动漫花园",
      "url": "https://share.dmhy.org/topics/rss/rss.xml",
      "folder_id": 1,
      "refresh_interval": 300,
      "last_refresh": "2024-01-01T10:30:00Z",
      "is_enabled": true,
      "has_error": false,
      "folder": {
        "name": "动漫",
        "path": "/动漫"
      }
    }
  ]
}
```

#### 添加订阅源
```http
POST /api/admin/rss/feeds
Content-Type: application/json

{
  "name": "DMHY 动漫花园",
  "url": "https://share.dmhy.org/topics/rss/rss.xml",
  "folder_path": "/动漫",
  "refresh_interval": 300
}
```

#### 更新订阅源
```http
PUT /api/admin/rss/feeds/{id}
Content-Type: application/json

{
  "name": "新名称",
  "refresh_interval": 600,
  "is_enabled": false
}
```

#### 手动刷新订阅
```http
POST /api/admin/rss/feeds/{id}/refresh
```

#### 删除订阅源
```http
DELETE /api/admin/rss/feeds/{id}
```

### 文章管理

#### 获取文章列表
```http
GET /api/admin/rss/articles?feed_id=1&page=1&per_page=50&unread_only=true
```

**参数说明:**
- `feed_id`: 订阅源ID (可选)
- `page`: 页码，默认1
- `per_page`: 每页数量，默认50
- `unread_only`: 只显示未读文章

#### 标记文章已读
```http
POST /api/admin/rss/articles/{id}/read
```

#### 标记所有文章已读
```http
POST /api/admin/rss/articles/read-all?feed_id=1
```

### 自动下载规则

#### 获取规则列表
```http
GET /api/admin/rss/rules
```

#### 创建自动下载规则
```http
POST /api/admin/rss/rules
Content-Type: application/json

{
  "name": "下载进击的巨人",
  "must_contain": "进击的巨人",
  "must_not_contain": "预告",
  "use_regex": false,
  "episode_filter": "S04",
  "smart_filter": true,
  "affected_feeds": ["feed-uuid-123"],
  "destination_path": "/downloads/anime",
  "add_paused": false
}
```

#### 更新规则
```http
PUT /api/admin/rss/rules/{id}
```

#### 启用/禁用规则
```http
POST /api/admin/rss/rules/{id}/toggle
```

#### 删除规则
```http
DELETE /api/admin/rss/rules/{id}
```

#### 获取可用的下载工具
```http
GET /api/admin/rss/download-tools
```

**响应示例:**
```json
{
  "code": 200,
  "data": {
    "tools": [
      {
        "name": "aria2",
        "display_name": "Aria2",
        "type": "local",
        "is_configured": true,
        "is_available": true,
        "categories": ["all"],
        "description": "高性能多协议下载工具，支持 HTTP、FTP、BitTorrent 等"
      },
      {
        "name": "PikPak",
        "display_name": "PikPak网盘",
        "type": "cloud",
        "is_configured": true,
        "is_available": true,
        "categories": ["all"],
        "description": "使用PikPak网盘的云端离线下载功能"
      },
      {
        "name": "115 Cloud",
        "display_name": "115云盘",
        "type": "cloud",
        "is_configured": false,
        "is_available": false,
        "categories": ["all"],
        "description": "使用115网盘的云端离线下载功能"
      }
    ],
    "recommended_tool": "PikPak"
  }
}
```

**下载工具说明:**
- **local 类型**: 本地下载工具 (aria2, qBittorrent, Transmission)
- **cloud 类型**: 云盘离线下载 (115云盘, PikPak, 迅雷网盘)
- **云盘优势**: 无需本地带宽，下载速度快，支持直接转存到目标目录

**自动下载规则字段说明:**
- `download_tool`: 下载工具名称 (默认: "aria2")
- `delete_policy`: 删除策略 (默认: "delete_on_upload_succeed")
  - `delete_on_upload_succeed`: 上传成功后删除
  - `delete_never`: 永不删除
  - `delete_on_upload_failed`: 上传失败后删除
- `torrent_temp_path`: 种子临时路径 (仅用于云盘下载)

## 🔍 资源搜索 API

### 搜索插件管理

#### 获取插件列表
```http
GET /api/admin/search/plugins
```

**响应示例:**
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "thepiratebay",
      "display_name": "The Pirate Bay",
      "version": "1.0.0",
      "is_enabled": true,
      "categories": ["all", "audio", "video", "applications", "games"]
    }
  ]
}
```

#### 安装搜索插件
```http
POST /api/admin/search/plugins
Content-Type: application/json

{
  "name": "thepiratebay",
  "url": "https://example.com/plugins/thepiratebay.py"
}
```

#### 启用/禁用插件
```http
POST /api/admin/search/plugins/{name}/enable
POST /api/admin/search/plugins/{name}/disable
```

#### 卸载插件
```http
DELETE /api/admin/search/plugins/{name}
```

### 资源搜索

#### 执行搜索
```http
POST /api/admin/search/
Content-Type: application/json

{
  "query": "进击的巨人 S04",
  "plugins": ["thepiratebay", "nyaa"],
  "category": "video",
  "min_seeds": 5
}
```

**响应示例:**
```json
{
  "code": 200,
  "data": {
    "job_id": "search-job-uuid-456",
    "status": "running",
    "start_time": "2024-01-01T10:30:00Z"
  }
}
```

#### 查询搜索状态
```http
GET /api/admin/search/jobs/{job_id}
```

#### 获取搜索结果
```http
GET /api/admin/search/results/{search_id}?page=1&per_page=50
```

**响应示例:**
```json
{
  "code": 200,
  "data": {
    "results": [
      {
        "id": 1,
        "plugin_name": "thepiratebay",
        "title": "[Leopard-Raws] 进击的巨人 S04E01 [1080p]",
        "url": "https://thepiratebay.org/details/123456",
        "magnet_link": "magnet:?xt=urn:btih:...",
        "size": "1.2 GiB",
        "seeds": 150,
        "leechs": 30,
        "category": "video"
      }
    ],
    "total": 100,
    "page": 1,
    "per_page": 50
  }
}
```

#### 下载搜索结果
```http
POST /api/admin/search/download
Content-Type: application/json

{
  "result_id": 1,
  "destination_path": "/downloads/anime"
}
```

#### 批量下载
```http
POST /api/admin/search/batch-download
Content-Type: application/json

{
  "result_ids": [1, 2, 3],
  "destination_path": "/downloads/anime"
}
```

## 🔧 搜索插件开发

### 插件接口规范

搜索插件需要支持以下命令行参数：

```bash
# 获取插件信息
python3 plugin.py --info

# 执行搜索
python3 plugin.py --search "查询关键词" --category "video" --page 1
```

### 输出格式

#### 插件信息输出
```json
{
  "display_name": "The Pirate Bay",
  "version": "1.0.0",
  "categories": ["all", "audio", "video", "applications", "games"]
}
```

#### 搜索结果输出
```json
[
  {
    "title": "资源标题",
    "url": "详情页URL",
    "torrent_url": "种子文件URL",
    "magnet_link": "磁力链接",
    "size": "文件大小",
    "seeds": 150,
    "leechs": 30,
    "category": "video"
  }
]
```

## 🎯 使用场景示例

### 场景1: 自动追番设置

1. **添加 RSS 订阅源**
```bash
curl -X POST "http://localhost:5244/api/admin/rss/feeds" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "动漫花园",
    "url": "https://share.dmhy.org/topics/rss/rss.xml",
    "folder_path": "/动漫",
    "refresh_interval": 300
  }'
```

2. **设置自动下载规则**
```bash
curl -X POST "http://localhost:5244/api/admin/rss/rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "自动下载新番",
    "must_contain": "1080p",
    "must_not_contain": "预告|PV|CM",
    "use_regex": false,
    "affected_feeds": ["feed-uuid-123"],
    "destination_path": "/downloads/anime",
    "download_tool": "PikPak",
    "delete_policy": "delete_on_upload_succeed",
    "torrent_temp_path": "/temp/torrents"
    }'

# 使用本地下载工具示例 (aria2)
curl -X POST "http://localhost:5244/api/admin/rss/rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "aria2下载规则",
    "must_contain": "电影",
    "destination_path": "/downloads/movies",
    "download_tool": "aria2",
    "delete_policy": "delete_on_upload_succeed"
    }'
```

### 场景2: 手动搜索下载

1. **搜索资源**
```bash
curl -X POST "http://localhost:5244/api/admin/search/" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "进击的巨人 最终季",
    "category": "video",
    "min_seeds": 10
  }'
```

2. **下载选中资源**
```bash
curl -X POST "http://localhost:5244/api/admin/search/download" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "result_id": 123,
    "destination_path": "/downloads/anime"
  }'
```

### 场景3: 云盘离线下载配置

1. **获取可用的下载工具**
```bash
curl -X GET "http://localhost:5244/api/admin/rss/download-tools" \
  -H "Authorization: Bearer $TOKEN"
```

2. **配置云盘下载工具** (以PikPak为例)
```bash
# 先在 alist 系统设置中配置 PikPak 存储
curl -X POST "http://localhost:5244/api/admin/setting/set_pikpak" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_dir": "/pikpak_temp"
  }'
```

3. **创建使用云盘的自动下载规则**
```bash
curl -X POST "http://localhost:5244/api/admin/rss/rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "PikPak云盘自动下载",
    "must_contain": "4K|2160p",
    "destination_path": "/downloads/4k-movies",
    "download_tool": "PikPak",
    "delete_policy": "delete_never",
    "torrent_temp_path": "/pikpak_temp/torrents"
  }'
```

4. **云盘下载优势**
- ✅ 无需本地带宽：直接在云端完成下载
- ✅ 下载速度快：利用云盘服务器的高速网络
- ✅ 自动转存：下载完成后自动转存到目标目录
- ✅ 节省流量：特别适合家庭带宽有限的用户

**支持的云盘服务:**
- **115云盘**: 需要配置115存储驱动
- **PikPak**: 需要配置PikPak存储驱动
- **迅雷网盘**: 需要配置Thunder存储驱动

## ⚙️ 配置说明

### RSS 配置参数

- `refresh_interval`: 刷新间隔（秒），默认300秒
- `max_articles_per_feed`: 每个订阅源最大文章数，默认1000
- `auto_download_enabled`: 是否启用自动下载，默认true

### 搜索配置参数

- `plugin_directory`: 搜索插件目录，默认 `data/search_plugins`
- `search_timeout`: 搜索超时时间（秒），默认30秒
- `max_concurrent_searches`: 最大并发搜索数，默认5

## 🚨 注意事项

1. **权限控制**: RSS 和搜索功能需要管理员权限
2. **资源使用**: 搜索插件会消耗 CPU 和网络资源
3. **法律合规**: 请确保下载的内容符合当地法律法规
4. **插件安全**: 只安装来源可信的搜索插件
5. **存储空间**: 自动下载可能快速消耗存储空间
6. **云盘配置**: 使用云盘下载前需要先配置对应的存储驱动
7. **临时路径**: 云盘下载建议设置临时路径以避免直接下载到目标目录
8. **下载限制**: 不同云盘服务可能有下载速度和并发数限制

## 🔄 与现有功能集成

- **离线下载**: 支持多种下载工具
  - **本地工具**: aria2, qBittorrent, Transmission
  - **云盘工具**: 115云盘, PikPak, 迅雷网盘
- **存储管理**: 下载文件自动保存到指定的存储位置
- **任务管理**: 下载任务显示在现有的任务管理界面
- **用户权限**: 遵循现有的用户权限体系
- **API 风格**: 保持与现有 API 的一致性
- **智能选择**: 系统优先推荐已配置的云盘下载工具