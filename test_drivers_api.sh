#!/bin/bash

# OpenList Workers - Driver 配置管理 API 测试脚本
# 使用 D1 数据库保存和管理驱动配置

BASE_URL="http://localhost:8787"  # 本地测试地址，实际部署时替换为你的 Workers 域名

echo "🚀 OpenList Workers - Driver 配置管理 API 测试"
echo "=========================================="

# 初始化系统
echo "📋 1. 初始化系统..."
curl -s "$BASE_URL/init" | jq '.'
echo ""

# 健康检查
echo "❤️ 2. 健康检查..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# 获取所有驱动配置列表
echo "📋 3. 获取所有驱动配置列表..."
curl -s "$BASE_URL/api/drivers" | jq '.'
echo ""

# 获取启用的驱动配置
echo "✅ 4. 获取启用的驱动配置..."
curl -s "$BASE_URL/api/drivers?enabled=true" | jq '.'
echo ""

# 获取单个驱动配置（通过名称）
echo "🔍 5. 获取单个驱动配置（Local）..."
curl -s "$BASE_URL/api/admin/driver/get?name=Local" | jq '.'
echo ""

# 创建新的驱动配置
echo "➕ 6. 创建新的驱动配置（WebDAV）..."
curl -s -X POST "$BASE_URL/api/admin/driver/create" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WebDAV",
    "display_name": "WebDAV 存储",
    "description": "WebDAV 协议存储",
    "config": "{\"url\": \"\", \"username\": \"\", \"password\": \"\"}",
    "icon": "globe",
    "enabled": true,
    "order": 6
  }' | jq '.'
echo ""

# 更新驱动配置
echo "✏️ 7. 更新驱动配置（WebDAV）..."
curl -s -X POST "$BASE_URL/api/admin/driver/update" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WebDAV",
    "display_name": "WebDAV 网络存储",
    "description": "支持 WebDAV 协议的网络存储服务",
    "config": "{\"url\": \"https://example.com/webdav\", \"username\": \"user\", \"password\": \"pass\"}",
    "icon": "globe",
    "enabled": true,
    "order": 6
  }' | jq '.'
echo ""

# 禁用驱动配置
echo "⏸️ 8. 禁用驱动配置（WebDAV）..."
# 首先获取 WebDAV 的 ID
WEBDAV_ID=$(curl -s "$BASE_URL/api/admin/driver/get?name=WebDAV" | jq -r '.data.id')
curl -s -X POST "$BASE_URL/api/admin/driver/disable?id=$WEBDAV_ID" | jq '.'
echo ""

# 启用驱动配置
echo "▶️ 9. 启用驱动配置（WebDAV）..."
curl -s -X POST "$BASE_URL/api/admin/driver/enable?id=$WEBDAV_ID" | jq '.'
echo ""

# 再次获取所有驱动配置，查看变化
echo "📋 10. 查看所有驱动配置（包含新创建的）..."
curl -s "$BASE_URL/api/drivers" | jq '.'
echo ""

# 删除驱动配置
echo "🗑️ 11. 删除驱动配置（WebDAV）..."
curl -s -X POST "$BASE_URL/api/admin/driver/delete?id=$WEBDAV_ID" | jq '.'
echo ""

# 最终状态检查
echo "🏁 12. 最终状态检查..."
curl -s "$BASE_URL/api/drivers" | jq '.data | {drivers: .drivers, total: .total, enabled_count: (.configs | map(select(.enabled)) | length)}'
echo ""

echo "✨ 测试完成！"
echo ""
echo "💡 使用说明："
echo "  - 所有配置数据保存在 D1 数据库中"
echo "  - 支持完整的 CRUD 操作"
echo "  - 可以动态启用/禁用驱动"
echo "  - 兼容原有的 /api/drivers 接口"
echo "  - 新增 /api/admin/driver/* 管理接口"
echo ""
echo "🔧 API 端点："
echo "  GET    /api/drivers                      - 获取驱动列表（兼容旧版）"
echo "  GET    /api/admin/driver/list           - 获取驱动配置列表"
echo "  GET    /api/admin/driver/get            - 获取单个驱动配置"
echo "  POST   /api/admin/driver/create         - 创建驱动配置"
echo "  POST   /api/admin/driver/update         - 更新驱动配置"
echo "  POST   /api/admin/driver/delete         - 删除驱动配置"
echo "  POST   /api/admin/driver/enable         - 启用驱动配置"
echo "  POST   /api/admin/driver/disable        - 禁用驱动配置"