#!/bin/bash

# OpenList Workers 离线下载 API 测试脚本

set -e

BASE_URL="http://localhost:8787"
USER_ID=1

echo "=== OpenList Workers 离线下载 API 测试 ==="

# 健康检查
echo "1. 健康检查..."
curl -s "${BASE_URL}/health" | jq .

# 初始化系统
echo -e "\n2. 初始化系统..."
curl -s "${BASE_URL}/init" | jq .

# 获取支持的离线下载工具列表
echo -e "\n3. 获取支持的离线下载工具列表..."
curl -s "${BASE_URL}/api/offline_download_tools" | jq .

# 获取用户的离线下载配置
echo -e "\n4. 获取用户的离线下载配置..."
curl -s "${BASE_URL}/api/user/offline_download/configs?user_id=${USER_ID}" | jq .

# 配置 Aria2 下载器
echo -e "\n5. 配置 Aria2 下载器..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_aria2?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "http://localhost:6800/jsonrpc",
    "secret": "my_secret_token"
  }' | jq .

# 配置 qBittorrent 下载器
echo -e "\n6. 配置 qBittorrent 下载器..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_qbittorrent?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://localhost:8080",
    "seedtime": "60"
  }' | jq .

# 配置 Transmission 下载器
echo -e "\n7. 配置 Transmission 下载器..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_transmission?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "uri": "http://localhost:9091/transmission/rpc",
    "seedtime": "120"
  }' | jq .

# 配置 115 云盘离线下载
echo -e "\n8. 配置 115 云盘离线下载..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_115?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_dir_path": "/downloads/115",
    "config_id": 1
  }' | jq .

# 配置 PikPak 离线下载
echo -e "\n9. 配置 PikPak 离线下载..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_pikpak?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_dir_path": "/downloads/pikpak",
    "config_id": 2
  }' | jq .

# 配置 Thunder 离线下载
echo -e "\n10. 配置 Thunder 离线下载..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_thunder?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_dir_path": "/downloads/thunder",
    "config_id": 3
  }' | jq .

# 再次获取用户的离线下载配置（查看配置后的结果）
echo -e "\n11. 再次获取用户的离线下载配置..."
curl -s "${BASE_URL}/api/user/offline_download/configs?user_id=${USER_ID}" | jq .

# 创建离线下载任务 - Aria2
echo -e "\n12. 创建 Aria2 离线下载任务..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/add_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "http://example.com/file1.zip",
      "http://example.com/file2.tar.gz"
    ],
    "config_id": 1,
    "dst_path": "/downloads",
    "tool": "aria2",
    "delete_policy": "keep"
  }' | jq .

# 创建离线下载任务 - 115
echo -e "\n13. 创建 115 离线下载任务..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/add_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "magnet:?xt=urn:btih:example123456789"
    ],
    "config_id": 1,
    "dst_path": "/downloads/115",
    "tool": "115",
    "delete_policy": "delete_on_complete"
  }' | jq .

# 创建离线下载任务 - qBittorrent
echo -e "\n14. 创建 qBittorrent 离线下载任务..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/add_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "magnet:?xt=urn:btih:another123456789"
    ],
    "config_id": 1,
    "dst_path": "/downloads/torrents",
    "tool": "qbittorrent",
    "delete_policy": "keep"
  }' | jq .

# 获取用户的离线下载任务列表
echo -e "\n15. 获取用户的离线下载任务列表..."
curl -s "${BASE_URL}/api/user/offline_download/tasks?user_id=${USER_ID}&page=1&per_page=10" | jq .

# 模拟更新任务状态 - 设置为运行中
echo -e "\n16. 更新任务状态为运行中..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/update_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 1,
    "status": "running",
    "progress": 25,
    "error": ""
  }' | jq .

# 模拟更新任务状态 - 设置为完成
echo -e "\n17. 更新任务状态为完成..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/update_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 1,
    "status": "completed",
    "progress": 100,
    "error": ""
  }' | jq .

# 模拟更新任务状态 - 设置为失败
echo -e "\n18. 更新任务状态为失败..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/update_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 2,
    "status": "failed",
    "progress": 0,
    "error": "Download failed: connection timeout"
  }' | jq .

# 再次获取任务列表查看状态更新
echo -e "\n19. 再次获取任务列表查看状态更新..."
curl -s "${BASE_URL}/api/user/offline_download/tasks?user_id=${USER_ID}" | jq .

# 删除一个任务
echo -e "\n20. 删除任务..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/delete_task?user_id=${USER_ID}&task_id=2" | jq .

# 最终获取任务列表
echo -e "\n21. 最终获取任务列表..."
curl -s "${BASE_URL}/api/user/offline_download/tasks?user_id=${USER_ID}" | jq .

echo -e "\n=== 离线下载 API 测试完成 ==="

# 测试错误情况
echo -e "\n=== 错误情况测试 ==="

# 测试配置不存在的驱动
echo -e "\n22. 测试配置不存在的驱动..."
curl -s -X POST "${BASE_URL}/api/admin/setting/set_115?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_dir_path": "/downloads/115",
    "config_id": 999
  }' | jq .

# 测试创建任务时缺少必需参数
echo -e "\n23. 测试创建任务时缺少必需参数..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/add_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [],
    "config_id": 1,
    "dst_path": "/downloads",
    "tool": "aria2"
  }' | jq .

# 测试更新不存在的任务
echo -e "\n24. 测试更新不存在的任务..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/update_task?user_id=${USER_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 999,
    "status": "completed",
    "progress": 100
  }' | jq .

# 测试删除不存在的任务
echo -e "\n25. 测试删除不存在的任务..."
curl -s -X POST "${BASE_URL}/api/user/offline_download/delete_task?user_id=${USER_ID}&task_id=999" | jq .

echo -e "\n=== 错误情况测试完成 ==="

echo -e "\n✅ 所有离线下载 API 测试完成"