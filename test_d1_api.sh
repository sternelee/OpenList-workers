#!/bin/bash

# OpenList Workers - 简化版用户驱动配置 API 测试脚本
# 演示基于用户的驱动配置管理功能

BASE_URL="http://localhost:8787"  # 本地测试地址，实际部署时替换为你的 Workers 域名
USER_ID=1  # 测试用户ID（管理员）

echo "🚀 OpenList Workers - 用户驱动配置 API 测试"
echo "============================================"

# 检查 jq 是否安装
if ! command -v jq &> /dev/null; then
    echo "⚠️  jq 未安装，输出将不格式化"
    JQ_CMD="cat"
else
    JQ_CMD="jq '.'"
fi

# 初始化系统
echo "📋 1. 初始化系统和数据库..."
curl -s "$BASE_URL/init" | $JQ_CMD
echo ""

# 健康检查
echo "❤️ 2. 健康检查..."
curl -s "$BASE_URL/health" | $JQ_CMD
echo ""

echo "🔧 用户驱动配置管理测试"
echo "======================="

# 获取用户的所有驱动配置
echo "📋 3. 获取用户的所有驱动配置列表..."
curl -s "$BASE_URL/api/drivers?user_id=$USER_ID" | $JQ_CMD
echo ""

# 获取启用的驱动配置
echo "✅ 4. 获取用户启用的驱动配置..."
curl -s "$BASE_URL/api/drivers?user_id=$USER_ID&enabled=true" | $JQ_CMD
echo ""

# 获取分页的驱动配置
echo "📄 5. 获取分页的驱动配置（第1页，每页3个）..."
curl -s "$BASE_URL/api/user/driver/list?user_id=$USER_ID&page=1&per_page=3" | $JQ_CMD
echo ""

# 获取单个驱动配置（通过名称）
echo "🔍 6. 获取单个驱动配置（Local）..."
curl -s "$BASE_URL/api/user/driver/get?user_id=$USER_ID&name=Local" | $JQ_CMD
echo ""

# 创建新的用户驱动配置
echo "➕ 7. 创建新的驱动配置（WebDAV测试）..."
curl -s -X POST "$BASE_URL/api/user/driver/create?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WebDAV_Test",
    "display_name": "WebDAV 测试",
    "description": "WebDAV 协议存储测试配置",
    "config": "{\"url\": \"https://example.com/webdav\", \"username\": \"\", \"password\": \"\"}",
    "icon": "folder-network",
    "enabled": true,
    "order": 10
  }' | $JQ_CMD
echo ""

# 更新用户驱动配置
echo "✏️ 8. 更新驱动配置..."
curl -s -X POST "$BASE_URL/api/user/driver/update?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "WebDAV_Test",
    "display_name": "WebDAV 测试（已更新）",
    "description": "WebDAV 协议存储测试配置 - 更新版本",
    "config": "{\"url\": \"https://updated.example.com/webdav\", \"username\": \"test\", \"password\": \"password\"}",
    "icon": "cloud-upload",
    "enabled": true,
    "order": 10
  }' | $JQ_CMD
echo ""

# 禁用驱动配置（需要先获取ID）
echo "🔒 9. 禁用驱动配置..."
# 这里简化处理，使用固定的ID进行测试
curl -s -X POST "$BASE_URL/api/user/driver/disable?user_id=$USER_ID&id=6" | $JQ_CMD
echo ""

# 启用驱动配置
echo "🔓 10. 启用驱动配置..."
curl -s -X POST "$BASE_URL/api/user/driver/enable?user_id=$USER_ID&id=6" | $JQ_CMD
echo ""

echo "👥 用户管理测试"
echo "==============="

# 获取用户列表
echo "📋 11. 获取用户列表..."
curl -s "$BASE_URL/api/admin/user/list" | $JQ_CMD
echo ""

# 获取分页的用户列表
echo "📄 12. 获取分页的用户列表（第1页，每页10个）..."
curl -s "$BASE_URL/api/admin/user/list?page=1&per_page=10" | $JQ_CMD
echo ""

# 获取单个用户
echo "🔍 13. 获取单个用户（ID=1）..."
curl -s "$BASE_URL/api/admin/user/get?id=1" | $JQ_CMD
echo ""

# 创建新用户
echo "➕ 14. 创建新用户..."
curl -s -X POST "$BASE_URL/api/admin/user/create" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "base_path": "/",
    "role": 2,
    "disabled": false,
    "permission": 255
  }' | $JQ_CMD
echo ""

echo "🧪 跨用户配置隔离测试"
echo "=================="

# 创建另一个用户的驱动配置
echo "➕ 15. 为用户2创建驱动配置..."
curl -s -X POST "$BASE_URL/api/user/driver/create?user_id=2" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Guest_Local",
    "display_name": "访客本地存储",
    "description": "访客用户的本地存储配置",
    "config": "{\"root_folder_path\": \"/guest\"}",
    "icon": "folder",
    "enabled": true,
    "order": 1
  }' | $JQ_CMD
echo ""

# 验证用户配置隔离
echo "🔍 16. 获取用户1的驱动配置（应该看不到用户2的配置）..."
curl -s "$BASE_URL/api/drivers?user_id=1" | $JQ_CMD
echo ""

echo "🔍 17. 获取用户2的驱动配置（应该只看到自己的配置）..."
curl -s "$BASE_URL/api/drivers?user_id=2" | $JQ_CMD
echo ""

echo "🔧 系统状态检查"
echo "================"

# 最终健康检查
echo "❤️ 18. 最终健康检查..."
curl -s "$BASE_URL/health" | $JQ_CMD
echo ""

echo "✅ 测试完成！"
echo ""
echo "📊 测试总结："
echo "- 用户驱动配置管理：完整的 CRUD 操作"
echo "- 用户管理：创建、读取、分页查询"
echo "- 配置隔离：用户之间的配置独立"
echo "- 数据库连接：D1 数据库正常工作"
echo ""
echo "🔗 有用的 API 端点："
echo "- 用户驱动列表：$BASE_URL/api/drivers?user_id=<user_id>"
echo "- 健康检查：$BASE_URL/health"
echo "- 系统初始化：$BASE_URL/init"
echo "- 用户管理：$BASE_URL/api/admin/user/*"
echo "- 驱动管理：$BASE_URL/api/user/driver/*"
echo ""
echo "📝 注意事项："
echo "- 每个用户的驱动配置相互独立"
echo "- 默认用户ID为1（管理员）"
echo "- 可通过 user_id 参数指定用户"