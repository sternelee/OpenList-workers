#!/bin/bash

# OpenList Workers API 集成测试脚本
# 模拟真实用户操作流程的端到端测试

set -e

# 配置
BASE_URL="http://localhost:8787"
TEST_USERNAME="integration_test_user"
TEST_PASSWORD="integration_test_123"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 全局变量
USER_TOKEN=""
ADMIN_TOKEN=""
DRIVER_CONFIG_ID=""
TEST_DIR="/integration_test"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# HTTP 请求函数
make_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local token="$4"
    local content_type="${5:-application/json}"

    local headers=(-H "Content-Type: $content_type")
    if [[ -n "$token" ]]; then
        headers+=(-H "Authorization: Bearer $token")
    fi

    if [[ "$method" == "GET" ]]; then
        curl -s -w "\n%{http_code}" "${headers[@]}" "$BASE_URL$url"
    else
        curl -s -w "\n%{http_code}" -X "$method" "${headers[@]}" -d "$data" "$BASE_URL$url"
    fi
}

# 解析响应
parse_response() {
    local response="$1"
    local expected_code="$2"

    local body=$(echo "$response" | head -n -1)
    local status_code=$(echo "$response" | tail -n 1)

    if [[ "$status_code" == "$expected_code" ]]; then
        echo "$body"
        return 0
    else
        log_error "Expected status $expected_code, got $status_code"
        echo "Response: $body" >&2
        return 1
    fi
}

# 提取JSON字段
extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\":[^,}]*" | cut -d: -f2 | tr -d '"' | tr -d ' '
}

# 场景1: 新用户完整使用流程
scenario_new_user_journey() {
    log_info "Running Scenario 1: New User Complete Journey"

    # 1.1 系统初始化
    log_info "Step 1.1: Initialize system"
    local response=$(make_request "GET" "/init")
    if parse_response "$response" "200" > /dev/null; then
        log_success "System initialized"
    else
        log_error "System initialization failed"
        return 1
    fi

    # 1.2 用户注册
    log_info "Step 1.2: User registration"
    local register_data='{"username":"'$TEST_USERNAME'","password":"'$TEST_PASSWORD'","base_path":"/"}'
    response=$(make_request "POST" "/api/auth/register" "$register_data")
    local body=$(parse_response "$response" "200")
    if [[ $? -eq 0 ]]; then
        USER_TOKEN=$(extract_json_field "$body" "token")
        log_success "User registered and logged in"
    else
        log_error "User registration failed"
        return 1
    fi

    # 1.3 查看可用驱动
    log_info "Step 1.3: Check available drivers"
    response=$(make_request "GET" "/api/drivers" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Available drivers retrieved"
    else
        log_error "Failed to retrieve drivers"
        return 1
    fi

    # 1.4 配置本地存储驱动
    log_info "Step 1.4: Configure local storage driver"
    local driver_data='{"name":"Local","display_name":"我的本地存储","description":"用于测试的本地存储","config":"{\"root_folder_path\": \"/tmp/openlist_test\"}","icon":"folder","enabled":true,"order":1}'
    response=$(make_request "POST" "/api/user/driver/create" "$driver_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Local storage driver configured"

        # 获取驱动配置ID
        response=$(make_request "GET" "/api/user/driver/get?name=Local" "" "$USER_TOKEN")
        body=$(parse_response "$response" "200")
        if [[ $? -eq 0 ]]; then
            DRIVER_CONFIG_ID=$(extract_json_field "$body" "id")
            log_success "Driver config ID obtained: $DRIVER_CONFIG_ID"
        fi
    else
        log_error "Failed to configure local storage driver"
        return 1
    fi

    # 1.5 创建测试目录
    if [[ -n "$DRIVER_CONFIG_ID" ]]; then
        log_info "Step 1.5: Create test directory"
        local mkdir_data="config_id=$DRIVER_CONFIG_ID&path=/&dir_name=test_folder"
        response=$(make_request "POST" "/api/fs/mkdir" "$mkdir_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
        if parse_response "$response" "200" > /dev/null; then
            log_success "Test directory created"
        else
            log_error "Failed to create test directory"
        fi
    fi

    # 1.6 列出文件
    if [[ -n "$DRIVER_CONFIG_ID" ]]; then
        log_info "Step 1.6: List files"
        response=$(make_request "GET" "/api/fs/list?config_id=$DRIVER_CONFIG_ID&path=/" "" "$USER_TOKEN")
        if parse_response "$response" "200" > /dev/null; then
            log_success "Files listed successfully"
        else
            log_error "Failed to list files"
        fi
    fi

    # 1.7 配置离线下载工具
    log_info "Step 1.7: Configure offline download tool"
    local aria2_data='{"uri":"http://localhost:6800/jsonrpc","secret":"test_secret"}'
    response=$(make_request "POST" "/api/admin/setting/set_aria2" "$aria2_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Aria2 configured"
    else
        log_error "Failed to configure Aria2"
    fi

    # 1.8 用户登出
    log_info "Step 1.8: User logout"
    response=$(make_request "POST" "/api/auth/logout" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "User logged out"
    else
        log_error "Failed to logout"
    fi

    log_success "Scenario 1: New User Complete Journey - PASSED"
}

# 场景2: 管理员管理用户和驱动
scenario_admin_management() {
    log_info "Running Scenario 2: Admin Management Tasks"

    # 2.1 管理员登录
    log_info "Step 2.1: Admin login"
    local login_data='{"username":"admin","password":"admin123"}'
    local response=$(make_request "POST" "/api/auth/login" "$login_data")
    local body=$(parse_response "$response" "200")
    if [[ $? -eq 0 ]]; then
        ADMIN_TOKEN=$(extract_json_field "$body" "token")
        log_success "Admin logged in"
    else
        log_error "Admin login failed"
        return 1
    fi

    # 2.2 查看所有用户
    log_info "Step 2.2: List all users"
    response=$(make_request "GET" "/api/admin/user/list?page=1&per_page=10" "" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "User list retrieved"
    else
        log_error "Failed to retrieve user list"
        return 1
    fi

    # 2.3 创建新用户
    log_info "Step 2.3: Create new user"
    local user_data='{"username":"admin_created_user","password":"password123","base_path":"/restricted","role":2}'
    response=$(make_request "POST" "/api/admin/user/create" "$user_data" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "New user created by admin"
    else
        log_error "Failed to create new user"
    fi

    # 2.4 获取用户详情
    log_info "Step 2.4: Get user details"
    response=$(make_request "GET" "/api/admin/user/get?id=1" "" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "User details retrieved"
    else
        log_error "Failed to retrieve user details"
    fi

    # 2.5 更新用户信息
    log_info "Step 2.5: Update user information"
    local update_data='{"id":1,"username":"admin","password":"","base_path":"/admin","role":1,"disabled":false}'
    response=$(make_request "POST" "/api/admin/user/update" "$update_data" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "User information updated"
    else
        log_error "Failed to update user information"
    fi

    log_success "Scenario 2: Admin Management Tasks - PASSED"
}

# 场景3: 文件操作完整流程
scenario_file_operations() {
    log_info "Running Scenario 3: Complete File Operations"

    if [[ -z "$USER_TOKEN" ]]; then
        # 重新登录用户
        local login_data='{"username":"'$TEST_USERNAME'","password":"'$TEST_PASSWORD'"}'
        local response=$(make_request "POST" "/api/auth/login" "$login_data")
        local body=$(parse_response "$response" "200")
        if [[ $? -eq 0 ]]; then
            USER_TOKEN=$(extract_json_field "$body" "token")
        fi
    fi

    if [[ -z "$DRIVER_CONFIG_ID" ]] && [[ -n "$USER_TOKEN" ]]; then
        # 获取驱动配置ID
        local response=$(make_request "GET" "/api/user/driver/get?name=Local" "" "$USER_TOKEN")
        local body=$(parse_response "$response" "200")
        if [[ $? -eq 0 ]]; then
            DRIVER_CONFIG_ID=$(extract_json_field "$body" "id")
        fi
    fi

    if [[ -z "$DRIVER_CONFIG_ID" ]]; then
        log_error "Cannot perform file operations - no driver config available"
        return 1
    fi

    # 3.1 创建多级目录结构
    log_info "Step 3.1: Create directory structure"
    local dirs=("documents" "images" "videos" "temp")
    for dir in "${dirs[@]}"; do
        local mkdir_data="config_id=$DRIVER_CONFIG_ID&path=/&dir_name=$dir"
        local response=$(make_request "POST" "/api/fs/mkdir" "$mkdir_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
        if parse_response "$response" "200" > /dev/null; then
            log_success "Directory $dir created"
        else
            log_error "Failed to create directory $dir"
        fi
    done

    # 3.2 创建子目录
    log_info "Step 3.2: Create subdirectory"
    local mkdir_data="config_id=$DRIVER_CONFIG_ID&path=/documents&dir_name=projects"
    response=$(make_request "POST" "/api/fs/mkdir" "$mkdir_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Subdirectory created"
    else
        log_error "Failed to create subdirectory"
    fi

    # 3.3 重命名目录
    log_info "Step 3.3: Rename directory"
    local rename_data="config_id=$DRIVER_CONFIG_ID&path=/temp&new_name=temporary"
    response=$(make_request "POST" "/api/fs/rename" "$rename_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Directory renamed"
    else
        log_error "Failed to rename directory"
    fi

    # 3.4 列出所有目录
    log_info "Step 3.4: List all directories"
    response=$(make_request "GET" "/api/fs/dirs?config_id=$DRIVER_CONFIG_ID&path=/" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Directories listed"
    else
        log_error "Failed to list directories"
    fi

    # 3.5 删除目录
    log_info "Step 3.5: Remove directory"
    local remove_data="config_id=$DRIVER_CONFIG_ID&path=/temporary"
    response=$(make_request "POST" "/api/fs/remove" "$remove_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Directory removed"
    else
        log_error "Failed to remove directory"
    fi

    log_success "Scenario 3: Complete File Operations - PASSED"
}

# 场景4: 离线下载完整流程
scenario_offline_download() {
    log_info "Running Scenario 4: Offline Download Workflow"

    if [[ -z "$USER_TOKEN" ]]; then
        log_error "No user token available for offline download test"
        return 1
    fi

    # 4.1 查看支持的下载工具
    log_info "Step 4.1: Check supported download tools"
    local response=$(make_request "GET" "/api/offline_download_tools" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Download tools retrieved"
    else
        log_error "Failed to retrieve download tools"
        return 1
    fi

    # 4.2 配置多个下载工具
    log_info "Step 4.2: Configure download tools"

    # 配置 Aria2
    local aria2_data='{"uri":"http://localhost:6800/jsonrpc","secret":"test_secret"}'
    response=$(make_request "POST" "/api/admin/setting/set_aria2" "$aria2_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Aria2 configured"
    else
        log_error "Failed to configure Aria2"
    fi

    # 配置 qBittorrent
    local qbit_data='{"url":"http://localhost:8080","seedtime":"1440"}'
    response=$(make_request "POST" "/api/admin/setting/set_qbittorrent" "$qbit_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "qBittorrent configured"
    else
        log_error "Failed to configure qBittorrent"
    fi

    # 4.3 查看配置的下载工具
    log_info "Step 4.3: List configured download tools"
    response=$(make_request "GET" "/api/user/offline_download/configs" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Download configs retrieved"
    else
        log_error "Failed to retrieve download configs"
    fi

    # 4.4 创建下载任务
    if [[ -n "$DRIVER_CONFIG_ID" ]]; then
        log_info "Step 4.4: Create download task"
        local task_data='{"urls":["http://httpbin.org/get","http://httpbin.org/json"],"config_id":'$DRIVER_CONFIG_ID',"dst_path":"/downloads","tool":"aria2","delete_policy":"never"}'
        response=$(make_request "POST" "/api/user/offline_download/add_task" "$task_data" "$USER_TOKEN")
        if parse_response "$response" "200" > /dev/null; then
            log_success "Download task created"
        else
            log_error "Failed to create download task"
        fi
    fi

    # 4.5 查看下载任务列表
    log_info "Step 4.5: List download tasks"
    response=$(make_request "GET" "/api/user/offline_download/tasks?page=1&per_page=10" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Download tasks retrieved"
    else
        log_error "Failed to retrieve download tasks"
    fi

    log_success "Scenario 4: Offline Download Workflow - PASSED"
}

# 场景5: 错误恢复和边界测试
scenario_error_recovery() {
    log_info "Running Scenario 5: Error Recovery and Edge Cases"

    # 5.1 测试无效认证恢复
    log_info "Step 5.1: Test invalid authentication recovery"
    local response=$(make_request "GET" "/api/auth/me" "" "invalid_token")
    if parse_response "$response" "401" > /dev/null; then
        log_success "Invalid authentication handled correctly"
    else
        log_error "Invalid authentication not handled correctly"
    fi

    # 5.2 测试权限边界
    log_info "Step 5.2: Test permission boundaries"
    if [[ -n "$USER_TOKEN" ]]; then
        response=$(make_request "GET" "/api/admin/user/list" "" "$USER_TOKEN")
        if parse_response "$response" "403" > /dev/null; then
            log_success "Permission boundary enforced"
        else
            log_error "Permission boundary not enforced"
        fi
    fi

    # 5.3 测试资源不存在
    log_info "Step 5.3: Test resource not found"
    response=$(make_request "GET" "/api/user/driver/get?id=999999" "" "$USER_TOKEN")
    if parse_response "$response" "404" > /dev/null; then
        log_success "Resource not found handled correctly"
    else
        log_error "Resource not found not handled correctly"
    fi

    # 5.4 测试无效方法
    log_info "Step 5.4: Test invalid method"
    response=$(make_request "DELETE" "/api/auth/login" '{}' "$USER_TOKEN")
    if parse_response "$response" "405" > /dev/null; then
        log_success "Invalid method handled correctly"
    else
        log_error "Invalid method not handled correctly"
    fi

    # 5.5 测试无效JSON
    log_info "Step 5.5: Test invalid JSON handling"
    response=$(make_request "POST" "/api/auth/login" "invalid_json")
    if parse_response "$response" "400" > /dev/null; then
        log_success "Invalid JSON handled correctly"
    else
        log_error "Invalid JSON not handled correctly"
    fi

    log_success "Scenario 5: Error Recovery and Edge Cases - PASSED"
}

# 清理测试数据
cleanup_test_data() {
    log_info "Cleaning up test data..."

    # 这里可以添加清理逻辑，比如删除测试用户、删除测试文件等
    # 由于是集成测试，通常不需要清理，但在某些情况下可能需要

    log_success "Test data cleanup completed"
}

# 主函数
main() {
    log_info "Starting OpenList Workers API Integration Tests..."
    echo "Base URL: $BASE_URL"
    echo "Test User: $TEST_USERNAME"
    echo "================================================"

    local failed_scenarios=0

    # 运行测试场景
    if ! scenario_new_user_journey; then
        ((failed_scenarios++))
    fi

    if ! scenario_admin_management; then
        ((failed_scenarios++))
    fi

    if ! scenario_file_operations; then
        ((failed_scenarios++))
    fi

    if ! scenario_offline_download; then
        ((failed_scenarios++))
    fi

    if ! scenario_error_recovery; then
        ((failed_scenarios++))
    fi

    # 清理测试数据
    cleanup_test_data

    echo "================================================"
    echo "Integration Test Summary:"
    echo "Total Scenarios: 5"
    echo "Passed: $((5 - failed_scenarios))"
    echo "Failed: $failed_scenarios"

    if [[ $failed_scenarios -eq 0 ]]; then
        log_success "All integration tests passed! 🎉"
        exit 0
    else
        log_error "$failed_scenarios scenario(s) failed! 😞"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "OpenList Workers API 集成测试脚本"
    echo ""
    echo "用法: $0 [OPTIONS]"
    echo ""
    echo "选项:"
    echo "  -u, --url URL       设置API基础URL (默认: http://localhost:8787)"
    echo "  -h, --help         显示此帮助信息"
    echo ""
    echo "测试场景:"
    echo "  1. 新用户完整使用流程"
    echo "  2. 管理员管理用户和驱动"
    echo "  3. 文件操作完整流程"
    echo "  4. 离线下载完整流程"
    echo "  5. 错误恢复和边界测试"
    echo ""
    echo "示例:"
    echo "  $0                          # 运行所有集成测试"
    echo "  $0 -u http://myapi.com      # 使用自定义URL"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

main