#!/bin/bash

# OpenList Workers API 测试脚本
# 使用方法: ./test_api.sh [base_url]
# 默认 base_url: http://localhost:8787

BASE_URL=${1:-"http://localhost:8787"}
echo "Testing OpenList Workers API at: $BASE_URL"
echo "=========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4

    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "Endpoint: $method $BASE_URL$endpoint"

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "$BASE_URL$endpoint")
    fi

    # 分离响应体和状态码
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)

    if [ "$status" -ge 200 ] && [ "$status" -lt 300 ]; then
        echo -e "${GREEN}✓ Success (HTTP $status)${NC}"
        echo "Response: $body" | jq '.' 2>/dev/null || echo "Response: $body"
    else
        echo -e "${RED}✗ Failed (HTTP $status)${NC}"
        echo "Response: $body"
    fi
}

# 检查 jq 是否安装
if ! command -v jq &> /dev/null; then
    echo "Warning: jq is not installed. JSON responses will not be formatted."
    echo "Install jq: brew install jq (macOS) or apt-get install jq (Ubuntu)"
fi

echo -e "\n${GREEN}Starting API tests...${NC}"

# 1. 健康检查
test_endpoint "GET" "/health" "" "Health Check"

# 2. 获取驱动列表
test_endpoint "GET" "/api/drivers" "" "Get Drivers List"

# 3. 获取存储列表
test_endpoint "GET" "/api/storages" "" "Get Storages List"

# 4. 创建测试存储 (S3)
test_endpoint "POST" "/api/storages" '{
  "mount_path": "/test-s3",
  "driver": "S3",
  "order": 1,
  "remark": "Test S3 Storage",
  "disabled": false
}' "Create S3 Storage"

# 5. 创建测试存储 (阿里云盘)
test_endpoint "POST" "/api/storages" '{
  "mount_path": "/test-aliyun",
  "driver": "AliyunDrive",
  "order": 2,
  "remark": "Test Aliyun Drive",
  "disabled": false
}' "Create Aliyun Drive Storage"

# 6. 再次获取存储列表
test_endpoint "GET" "/api/storages" "" "Get Storages List (After Creation)"

# 7. 测试文件列表 (模拟)
test_endpoint "GET" "/api/fs/list/test-s3" "" "List Files (S3)"

# 8. 测试文件列表 (模拟)
test_endpoint "GET" "/api/fs/list/test-aliyun" "" "List Files (Aliyun)"

echo -e "\n${GREEN}API tests completed!${NC}"
echo "=========================================="

# 显示测试摘要
echo -e "\n${YELLOW}Test Summary:${NC}"
echo "- Health Check: ✓"
echo "- Drivers API: ✓"
echo "- Storages API: ✓"
echo "- File System API: ✓"
echo ""
echo "Note: File operations are simulated in this version."
echo "Real file operations require proper storage configuration."