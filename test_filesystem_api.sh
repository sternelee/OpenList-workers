#!/bin/bash

# OpenList Workers 文件系统 API 测试脚本

set -e

BASE_URL="http://localhost:8787"
USER_ID=1
CONFIG_ID=1

echo "=== OpenList Workers 文件系统 API 测试 ==="

# 健康检查
echo "1. 健康检查..."
curl -s "${BASE_URL}/health" | jq .

# 初始化系统
echo -e "\n2. 初始化系统..."
curl -s "${BASE_URL}/init" | jq .

# 获取用户驱动配置列表
echo -e "\n3. 获取用户驱动配置列表..."
curl -s "${BASE_URL}/api/drivers?user_id=${USER_ID}" | jq .

# 测试文件系统操作

# 列出根目录文件
echo -e "\n4. 列出根目录文件..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/&per_page=20" | jq .

# 列出目录（只显示目录）
echo -e "\n5. 列出目录（只显示目录）..."
curl -s "${BASE_URL}/api/fs/dirs?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/" | jq .

# 创建目录
echo -e "\n6. 创建目录 test_dir..."
curl -s -X POST "${BASE_URL}/api/fs/mkdir" \
  -d "user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/&dir_name=test_dir" | jq .

# 再次列出根目录查看新创建的目录
echo -e "\n7. 再次列出根目录查看新创建的目录..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/" | jq .

# 创建测试文件
echo -e "\n8. 上传测试文件..."
echo "Hello, OpenList Workers!" > test_file.txt
curl -s -X PUT "${BASE_URL}/api/fs/upload?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir&filename=test_file.txt" \
  --data-binary @test_file.txt | jq .

# 列出 test_dir 目录
echo -e "\n9. 列出 test_dir 目录..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir" | jq .

# 获取文件信息
echo -e "\n10. 获取文件信息..."
curl -s "${BASE_URL}/api/fs/get?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir/test_file.txt" | jq .

# 重命名文件
echo -e "\n11. 重命名文件..."
curl -s -X POST "${BASE_URL}/api/fs/rename" \
  -d "user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir/test_file.txt&new_name=renamed_file.txt" | jq .

# 再次列出 test_dir 目录查看重命名结果
echo -e "\n12. 再次列出 test_dir 目录查看重命名结果..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir" | jq .

# 创建另一个目录用于移动测试
echo -e "\n13. 创建 another_dir 目录..."
curl -s -X POST "${BASE_URL}/api/fs/mkdir" \
  -d "user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/&dir_name=another_dir" | jq .

# 移动文件
echo -e "\n14. 移动文件到 another_dir..."
curl -s -X POST "${BASE_URL}/api/fs/move" \
  -d "user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir/renamed_file.txt&dst_path=/another_dir" | jq .

# 列出 another_dir 目录查看移动结果
echo -e "\n15. 列出 another_dir 目录查看移动结果..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/another_dir" | jq .

# 复制文件
echo -e "\n16. 复制文件回 test_dir..."
curl -s -X POST "${BASE_URL}/api/fs/copy" \
  -d "user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/another_dir/renamed_file.txt&dst_path=/test_dir" | jq .

# 列出两个目录查看复制结果
echo -e "\n17. 列出 test_dir 目录查看复制结果..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir" | jq .

echo -e "\n18. 列出 another_dir 目录..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/another_dir" | jq .

# 删除文件
echo -e "\n19. 删除 another_dir 中的文件..."
curl -s -X POST "${BASE_URL}/api/fs/remove" \
  -d "user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/another_dir/renamed_file.txt" | jq .

# 删除目录
echo -e "\n20. 删除 another_dir 目录..."
curl -s -X POST "${BASE_URL}/api/fs/remove" \
  -d "user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/another_dir" | jq .

# 最终列出根目录
echo -e "\n21. 最终列出根目录..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/" | jq .

# 测试文件下载（重定向）
echo -e "\n22. 测试文件下载链接（重定向）..."
echo "下载链接: ${BASE_URL}/d/?user_id=${USER_ID}&config_id=${CONFIG_ID}&path=/test_dir/renamed_file.txt"

# 清理测试文件
rm -f test_file.txt

echo -e "\n=== 文件系统 API 测试完成 ==="

# 测试错误情况
echo -e "\n=== 错误情况测试 ==="

# 测试不存在的配置ID
echo -e "\n23. 测试不存在的配置ID..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&config_id=999&path=/" | jq .

# 测试未提供配置ID
echo -e "\n24. 测试未提供配置ID..."
curl -s "${BASE_URL}/api/fs/list?user_id=${USER_ID}&path=/" | jq .

# 测试不支持的操作（假设Local驱动不支持某些操作）
echo -e "\n25. 测试可能不支持的操作..."
curl -s -X POST "${BASE_URL}/api/fs/copy" \
  -d "user_id=${USER_ID}&config_id=2&path=/test&dst_path=/test2" | jq .

echo -e "\n=== 错误情况测试完成 ==="

echo -e "\n✅ 所有文件系统 API 测试完成"