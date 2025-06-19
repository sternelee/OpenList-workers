# 云盘下载功能示例

## 概述

这个示例演示了如何使用 alist RSS 管理功能的云盘离线下载特性。通过集成 alist 现有的离线下载系统，支持使用云盘服务 (115云盘、PikPak、迅雷网盘) 进行高速离线下载。

## 云盘下载优势

✅ **无需本地带宽**: 直接在云端完成下载，不占用家庭带宽
✅ **下载速度快**: 利用云盘服务器的高速网络
✅ **自动转存**: 下载完成后自动转存到目标目录
✅ **节省流量**: 特别适合家庭带宽有限的用户
✅ **24/7 运行**: 云端下载不受本地设备影响

## 支持的云盘服务

| 云盘服务 | 配置要求 | 特色功能 |
|---------|----------|----------|
| **115云盘** | 需要115存储驱动 | 大容量存储 |
| **PikPak** | 需要PikPak存储驱动 | 高速下载 |
| **迅雷网盘** | 需要Thunder存储驱动 | 资源丰富 |

## 运行示例

### 前置条件

1. **配置云盘存储**: 在 alist 中配置至少一个云盘存储驱动
2. **获取API Token**: 从 alist 管理界面获取访问令牌
3. **设置环境变量**: 设置 `ALIST_TOKEN` 环境变量

### 运行步骤

```bash
# 1. 设置环境变量
export ALIST_TOKEN="your_api_token_here"

# 2. 运行示例程序
go run cloud_download_example.go
```

### 示例输出

```
=== 获取可用的下载工具 ===
推荐工具: PikPak
可用工具:
  - Aria2 (local) - 高性能多协议下载工具，支持 HTTP、FTP、BitTorrent 等 [✅ 可用]
  - qBittorrent (local) - 功能强大的 BitTorrent 客户端，支持做种管理 [❌ 未配置]
  - PikPak网盘 (cloud) - 使用PikPak网盘的云端离线下载功能 [✅ 可用]
  - 115云盘 (cloud) - 使用115网盘的云端离线下载功能 [❌ 未配置]

✅ 选择云盘工具: PikPak网盘

=== 创建云盘自动下载规则 ===
✅ 云盘自动下载规则创建成功!
规则名称: 云盘自动下载示例
下载工具: PikPak
目标路径: /downloads/auto
临时路径: /temp/torrents

=== 使用说明 ===
1. 确保已配置相应的云盘存储驱动
2. RSS 订阅将自动使用云盘下载匹配的资源
3. 下载完成后文件会自动转存到目标目录
4. 可以在任务管理界面查看下载进度

🎉 云盘下载配置完成!
```

## API 使用说明

### 获取可用下载工具

```bash
curl -X GET "http://localhost:5244/api/admin/rss/download-tools" \
  -H "Authorization: Bearer $ALIST_TOKEN"
```

### 创建云盘下载规则

```bash
curl -X POST "http://localhost:5244/api/admin/rss/rules" \
  -H "Authorization: Bearer $ALIST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "PikPak自动下载",
    "must_contain": "1080p",
    "destination_path": "/downloads/auto",
    "download_tool": "PikPak",
    "delete_policy": "delete_on_upload_succeed",
    "torrent_temp_path": "/temp/torrents"
  }'
```

## 配置说明

### 下载工具配置

- `download_tool`: 选择的下载工具名称
- `delete_policy`: 删除策略
  - `delete_on_upload_succeed`: 上传成功后删除 (推荐)
  - `delete_never`: 永不删除
  - `delete_on_upload_failed`: 上传失败后删除

### 路径配置

- `destination_path`: 最终存储路径
- `torrent_temp_path`: 种子临时路径 (云盘下载专用)

## 常见问题

### Q1: 为什么没有可用的云盘工具？
**A**: 需要先在 alist 管理界面配置相应的云盘存储驱动。

### Q2: 云盘下载失败怎么办？
**A**: 检查云盘账号状态、存储空间和网络连接。

### Q3: 如何查看下载进度？
**A**: 在 alist 任务管理界面可以查看下载任务状态。

### Q4: 临时路径有什么用？
**A**: 云盘下载时先下载到临时路径，完成后再转存到目标路径，避免重复下载。

## 更多信息

- [完整API文档](../docs/rss_search_api.md)
- [alist官方文档](https://alist.nn.ci)
- [RSS功能说明](../README.md)