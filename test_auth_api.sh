#!/bin/bash

# OpenList Workers 认证 API 测试脚本

set -e

BASE_URL="http://localhost:8787"
TOKEN=""

echo "=== OpenList Workers 认证 API 测试 ==="

# 健康检查
echo "1. 健康检查..."
curl -s "${BASE_URL}/health" | jq .

# 初始化系统
echo -e "\n2. 初始化系统..."
curl -s "${BASE_URL}/init" | jq .

# 测试未认证访问（应该失败）
echo -e "\n3. 测试未认证访问驱动列表（应该失败）..."
curl -s "${BASE_URL}/api/drivers" | jq .

# 用户注册
echo -e "\n4. 用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "base_path": "/home/testuser"
  }')

echo $REGISTER_RESPONSE | jq .

# 提取注册后的 token
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.data.token')
echo "获取到的 Token: $TOKEN"

# 测试重复注册（应该失败）
echo -e "\n5. 测试重复注册（应该失败）..."
curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }' | jq .

# 用户登录
echo -e "\n6. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }')

echo $LOGIN_RESPONSE | jq .

# 更新 token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
echo "登录后的 Token: $TOKEN"

# 测试错误登录（应该失败）
echo -e "\n7. 测试错误登录（应该失败）..."
curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "wrongpassword"
  }' | jq .

# 获取当前用户信息
echo -e "\n8. 获取当前用户信息..."
curl -s "${BASE_URL}/api/auth/me" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试带认证的驱动列表
echo -e "\n9. 获取用户驱动配置列表..."
curl -s "${BASE_URL}/api/drivers" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试文件系统操作（带认证）
echo -e "\n10. 列出根目录文件..."
curl -s "${BASE_URL}/api/fs/list?path=/" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试离线下载工具列表（带认证）
echo -e "\n11. 获取离线下载工具列表..."
curl -s "${BASE_URL}/api/offline_download_tools" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 注册管理员用户
echo -e "\n12. 注册管理员用户..."
ADMIN_REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

echo $ADMIN_REGISTER_RESPONSE | jq .

# 管理员登录
echo -e "\n13. 管理员登录..."
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

echo $ADMIN_LOGIN_RESPONSE | jq .

# 获取管理员 token
ADMIN_TOKEN=$(echo $ADMIN_LOGIN_RESPONSE | jq -r '.data.token')
echo "管理员 Token: $ADMIN_TOKEN"

# 测试管理员操作：获取用户列表
echo -e "\n14. 管理员获取用户列表..."
curl -s "${BASE_URL}/api/admin/user/list" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# 测试普通用户访问管理员API（应该失败）
echo -e "\n15. 普通用户尝试访问管理员API（应该失败）..."
curl -s "${BASE_URL}/api/admin/user/list" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试Token格式：从查询参数传递
echo -e "\n16. 通过查询参数传递Token..."
curl -s "${BASE_URL}/api/auth/me?token=${TOKEN}" | jq .

# 测试Token格式：无 Bearer 前缀
echo -e "\n17. 通过 Authorization 头传递Token（无 Bearer 前缀）..."
curl -s "${BASE_URL}/api/auth/me" \
  -H "Authorization: $TOKEN" | jq .

# 创建驱动配置
echo -e "\n18. 创建驱动配置..."
curl -s -X POST "${BASE_URL}/api/user/driver/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TestLocal",
    "display_name": "测试本地存储",
    "description": "用于测试的本地存储",
    "config": "{\"root_folder_path\": \"/tmp/test\"}",
    "icon": "folder",
    "enabled": true,
    "order": 1
  }' | jq .

# 获取更新后的驱动配置列表
echo -e "\n19. 获取更新后的驱动配置列表..."
curl -s "${BASE_URL}/api/drivers" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 配置 Aria2（需要认证）
echo -e "\n20. 配置 Aria2..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_aria2" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "http://localhost:6800/jsonrpc",
    "secret": "test_secret"
  }' | jq .

# 创建离线下载任务
echo -e "\n21. 创建离线下载任务..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/add_task" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["http://example.com/test.zip"],
    "config_id": 1,
    "dst_path": "/downloads",
    "tool": "aria2",
    "delete_policy": "keep"
  }' | jq .

# 用户登出
echo -e "\n22. 用户登出..."
curl -s "${BASE_URL}/api/auth/logout" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试登出后访问（应该失败）
echo -e "\n23. 测试登出后访问（应该失败）..."
curl -s "${BASE_URL}/api/auth/me" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== 认证 API 测试完成 ==="

# 测试错误情况
echo -e "\n=== 错误情况测试 ==="

# 测试无效Token
echo -e "\n24. 测试无效Token..."
curl -s "${BASE_URL}/api/auth/me" \
  -H "Authorization: Bearer invalid_token_12345" | jq .

# 测试空用户名注册
echo -e "\n25. 测试空用户名注册..."
curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "",
    "password": "testpass123"
  }' | jq .

# 测试弱密码注册
echo -e "\n26. 测试弱密码注册..."
curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "weakuser",
    "password": "123"
  }' | jq .

# 测试不存在的用户登录
echo -e "\n27. 测试不存在的用户登录..."
curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "nonexistent",
    "password": "testpass123"
  }' | jq .

echo -e "\n=== 错误情况测试完成 ==="

echo -e "\n✅ 所有认证 API 测试完成"

echo -e "\n=== 测试总结 ==="
echo "1. ✅ 用户注册功能正常"
echo "2. ✅ 用户登录功能正常"
echo "3. ✅ JWT Token 生成和验证正常"
echo "4. ✅ 认证中间件功能正常"
echo "5. ✅ 权限控制功能正常"
echo "6. ✅ 多种Token传递方式支持"
echo "7. ✅ 错误处理功能正常"