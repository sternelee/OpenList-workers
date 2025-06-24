#!/bin/bash

# OpenList Workers 完整API自动化测试脚本
# 测试所有功能模块的API接口

set -e

# 配置
BASE_URL="http://localhost:8787"
TEMP_DIR="/tmp/openlist_test"
TEST_FILE="test_upload.txt"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 全局变量
ADMIN_TOKEN=""
USER_TOKEN=""
USER_ID=""
DRIVER_CONFIG_ID=""
OFFLINE_TASK_ID=""

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 计数函数
count_test() {
    ((TOTAL_TESTS++))
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

# 创建测试环境
setup_test_env() {
    log_info "Setting up test environment..."

    # 创建临时目录
    mkdir -p "$TEMP_DIR"

    # 创建测试文件
    echo "This is a test file for upload testing." > "$TEMP_DIR/$TEST_FILE"

    log_success "Test environment setup completed"
}

# 清理测试环境
cleanup_test_env() {
    log_info "Cleaning up test environment..."
    rm -rf "$TEMP_DIR"
    log_success "Test environment cleanup completed"
}

# 1. 测试系统初始化和健康检查
test_system_health() {
    log_info "Testing system health and initialization..."

    # 测试健康检查
    count_test
    local response=$(make_request "GET" "/health")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Health check passed"
    else
        log_error "Health check failed"
        return 1
    fi

    # 测试系统初始化
    count_test
    response=$(make_request "GET" "/init")
    if parse_response "$response" "200" > /dev/null; then
        log_success "System initialization passed"
    else
        log_error "System initialization failed"
        return 1
    fi
}

# 2. 测试用户认证
test_user_auth() {
    log_info "Testing user authentication..."

    # 2.1 用户注册
    count_test
    local register_data='{"username":"testuser","password":"password123","base_path":"/"}'
    local response=$(make_request "POST" "/api/auth/register" "$register_data")
    local body=$(parse_response "$response" "200")
    if [[ $? -eq 0 ]]; then
        USER_TOKEN=$(extract_json_field "$body" "token")
        if [[ -n "$USER_TOKEN" ]]; then
            log_success "User registration passed"
        else
            log_error "User registration failed - no token returned"
            return 1
        fi
    else
        log_error "User registration failed"
        return 1
    fi

    # 2.2 管理员登录（使用默认管理员账户）
    count_test
    local login_data='{"username":"admin","password":"admin123"}'
    response=$(make_request "POST" "/api/auth/login" "$login_data")
    body=$(parse_response "$response" "200")
    if [[ $? -eq 0 ]]; then
        ADMIN_TOKEN=$(extract_json_field "$body" "token")
        if [[ -n "$ADMIN_TOKEN" ]]; then
            log_success "Admin login passed"
        else
            log_error "Admin login failed - no token returned"
            return 1
        fi
    else
        log_error "Admin login failed"
        return 1
    fi

    # 2.3 获取当前用户信息
    count_test
    response=$(make_request "GET" "/api/auth/me" "" "$USER_TOKEN")
    body=$(parse_response "$response" "200")
    if [[ $? -eq 0 ]]; then
        USER_ID=$(extract_json_field "$body" "id")
        log_success "Get current user info passed (User ID: $USER_ID)"
    else
        log_error "Get current user info failed"
        return 1
    fi

    # 2.4 用户登出
    count_test
    response=$(make_request "POST" "/api/auth/logout" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "User logout passed"
    else
        log_error "User logout failed"
    fi
}

# 3. 测试用户管理（管理员权限）
test_user_management() {
    log_info "Testing user management (admin)..."

    # 3.1 获取用户列表
    count_test
    local response=$(make_request "GET" "/api/admin/user/list?page=1&per_page=10" "" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Get users list passed"
    else
        log_error "Get users list failed"
    fi

    # 3.2 获取单个用户
    count_test
    response=$(make_request "GET" "/api/admin/user/get?id=1" "" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Get single user passed"
    else
        log_error "Get single user failed"
    fi

    # 3.3 创建用户
    count_test
    local user_data='{"username":"apitest","password":"test123","base_path":"/","role":2}'
    response=$(make_request "POST" "/api/admin/user/create" "$user_data" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Create user passed"
    else
        log_error "Create user failed"
    fi

    # 3.4 更新用户
    count_test
    local update_data='{"id":2,"username":"apitest_updated","password":"","base_path":"/updated","role":2}'
    response=$(make_request "POST" "/api/admin/user/update" "$update_data" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Update user passed"
    else
        log_error "Update user failed"
    fi
}

# 4. 测试驱动配置管理
test_driver_management() {
    log_info "Testing driver configuration management..."

    # 4.1 获取驱动列表（兼容API）
    count_test
    local response=$(make_request "GET" "/api/drivers" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Get drivers list passed"
    else
        log_error "Get drivers list failed"
    fi

    # 4.2 获取用户驱动配置列表
    count_test
    response=$(make_request "GET" "/api/user/driver/list?page=1&per_page=10" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Get user driver configs passed"
    else
        log_error "Get user driver configs failed"
    fi

    # 4.3 创建驱动配置
    count_test
    local driver_data='{"name":"TestLocal","display_name":"测试本地存储","description":"测试用本地存储","config":"{\"root_folder_path\": \"/tmp/test\"}","icon":"folder","enabled":true,"order":100}'
    response=$(make_request "POST" "/api/user/driver/create" "$driver_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Create driver config passed"
    else
        log_error "Create driver config failed"
    fi

    # 4.4 获取单个驱动配置
    count_test
    response=$(make_request "GET" "/api/user/driver/get?name=Local" "" "$USER_TOKEN")
    local body=$(parse_response "$response" "200")
    if [[ $? -eq 0 ]]; then
        DRIVER_CONFIG_ID=$(extract_json_field "$body" "id")
        log_success "Get single driver config passed (ID: $DRIVER_CONFIG_ID)"
    else
        log_error "Get single driver config failed"
    fi

    # 4.5 更新驱动配置
    count_test
    local update_data='{"name":"Local","display_name":"本地存储(更新)","description":"更新后的本地存储","config":"{\"root_folder_path\": \"/data/updated\"}","icon":"folder","enabled":true,"order":1}'
    response=$(make_request "POST" "/api/user/driver/update" "$update_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Update driver config passed"
    else
        log_error "Update driver config failed"
    fi

    # 4.6 禁用驱动配置
    if [[ -n "$DRIVER_CONFIG_ID" ]]; then
        count_test
        response=$(make_request "POST" "/api/user/driver/disable?id=$DRIVER_CONFIG_ID" "" "$USER_TOKEN")
        if parse_response "$response" "200" > /dev/null; then
            log_success "Disable driver config passed"
        else
            log_error "Disable driver config failed"
        fi

        # 4.7 启用驱动配置
        count_test
        response=$(make_request "POST" "/api/user/driver/enable?id=$DRIVER_CONFIG_ID" "" "$USER_TOKEN")
        if parse_response "$response" "200" > /dev/null; then
            log_success "Enable driver config passed"
        else
            log_error "Enable driver config failed"
        fi
    fi
}

# 5. 测试文件系统操作
test_filesystem_operations() {
    log_info "Testing filesystem operations..."

    if [[ -z "$DRIVER_CONFIG_ID" ]]; then
        log_warning "Skipping filesystem tests - no driver config ID available"
        return
    fi

    # 5.1 列出文件
    count_test
    local response=$(make_request "GET" "/api/fs/list?config_id=$DRIVER_CONFIG_ID&path=/" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "List files passed"
    else
        log_error "List files failed"
    fi

    # 5.2 获取目录列表
    count_test
    response=$(make_request "GET" "/api/fs/dirs?config_id=$DRIVER_CONFIG_ID&path=/" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "List directories passed"
    else
        log_error "List directories failed"
    fi

    # 5.3 创建目录
    count_test
    local mkdir_data="config_id=$DRIVER_CONFIG_ID&path=/&dir_name=test_dir"
    response=$(make_request "POST" "/api/fs/mkdir" "$mkdir_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Create directory passed"
    else
        log_error "Create directory failed"
    fi

    # 5.4 重命名目录
    count_test
    local rename_data="config_id=$DRIVER_CONFIG_ID&path=/test_dir&new_name=test_dir_renamed"
    response=$(make_request "POST" "/api/fs/rename" "$rename_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Rename directory passed"
    else
        log_error "Rename directory failed"
    fi

    # 5.5 删除目录
    count_test
    local remove_data="config_id=$DRIVER_CONFIG_ID&path=/test_dir_renamed"
    response=$(make_request "POST" "/api/fs/remove" "$remove_data" "$USER_TOKEN" "application/x-www-form-urlencoded")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Remove directory passed"
    else
        log_error "Remove directory failed"
    fi
}

# 6. 测试离线下载功能
test_offline_download() {
    log_info "Testing offline download functionality..."

    # 6.1 获取支持的下载工具
    count_test
    local response=$(make_request "GET" "/api/offline_download_tools" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Get offline download tools passed"
    else
        log_error "Get offline download tools failed"
    fi

    # 6.2 获取用户离线下载配置
    count_test
    response=$(make_request "GET" "/api/user/offline_download/configs" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Get offline download configs passed"
    else
        log_error "Get offline download configs failed"
    fi

    # 6.3 配置 Aria2
    count_test
    local aria2_data='{"uri":"http://localhost:6800/jsonrpc","secret":"test_secret"}'
    response=$(make_request "POST" "/api/admin/setting/set_aria2" "$aria2_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Configure Aria2 passed"
    else
        log_error "Configure Aria2 failed"
    fi

    # 6.4 获取用户离线下载任务
    count_test
    response=$(make_request "GET" "/api/user/offline_download/tasks?page=1&per_page=10" "" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Get offline download tasks passed"
    else
        log_error "Get offline download tasks failed"
    fi

    # 6.5 添加离线下载任务
    if [[ -n "$DRIVER_CONFIG_ID" ]]; then
        count_test
        local task_data='{"urls":["http://example.com/test.zip"],"config_id":'$DRIVER_CONFIG_ID',"dst_path":"/downloads","tool":"aria2","delete_policy":"never"}'
        response=$(make_request "POST" "/api/user/offline_download/add_task" "$task_data" "$USER_TOKEN")
        local body=$(parse_response "$response" "200")
        if [[ $? -eq 0 ]]; then
            log_success "Add offline download task passed"
        else
            log_error "Add offline download task failed"
        fi
    fi
}

# 7. 测试错误处理
test_error_handling() {
    log_info "Testing error handling..."

    # 7.1 未认证访问
    count_test
    local response=$(make_request "GET" "/api/auth/me")
    if parse_response "$response" "401" > /dev/null; then
        log_success "Unauthorized access handling passed"
    else
        log_error "Unauthorized access handling failed"
    fi

    # 7.2 无效Token
    count_test
    response=$(make_request "GET" "/api/auth/me" "" "invalid_token")
    if parse_response "$response" "401" > /dev/null; then
        log_success "Invalid token handling passed"
    else
        log_error "Invalid token handling failed"
    fi

    # 7.3 权限不足（普通用户访问管理员API）
    count_test
    response=$(make_request "GET" "/api/admin/user/list" "" "$USER_TOKEN")
    if parse_response "$response" "403" > /dev/null; then
        log_success "Insufficient permission handling passed"
    else
        log_error "Insufficient permission handling failed"
    fi

    # 7.4 无效参数
    count_test
    response=$(make_request "GET" "/api/user/driver/get?id=999999" "" "$USER_TOKEN")
    if parse_response "$response" "404" > /dev/null; then
        log_success "Invalid parameter handling passed"
    else
        log_error "Invalid parameter handling failed"
    fi

    # 7.5 方法不允许
    count_test
    response=$(make_request "PUT" "/api/auth/login" '{}' "$USER_TOKEN")
    if parse_response "$response" "405" > /dev/null; then
        log_success "Method not allowed handling passed"
    else
        log_error "Method not allowed handling failed"
    fi
}

# 8. 测试数据一致性
test_data_consistency() {
    log_info "Testing data consistency..."

    # 8.1 创建然后获取驱动配置
    count_test
    local driver_name="ConsistencyTest"
    local create_data='{"name":"'$driver_name'","display_name":"一致性测试","description":"测试数据一致性","config":"{}","icon":"test","enabled":true,"order":999}'
    local response=$(make_request "POST" "/api/user/driver/create" "$create_data" "$USER_TOKEN")

    if parse_response "$response" "200" > /dev/null; then
        # 获取刚创建的配置
        response=$(make_request "GET" "/api/user/driver/get?name=$driver_name" "" "$USER_TOKEN")
        local body=$(parse_response "$response" "200")
        if [[ $? -eq 0 ]]; then
            local retrieved_name=$(extract_json_field "$body" "name")
            if [[ "$retrieved_name" == "$driver_name" ]]; then
                log_success "Data consistency test passed"
            else
                log_error "Data consistency test failed - name mismatch"
            fi
        else
            log_error "Data consistency test failed - cannot retrieve created config"
        fi
    else
        log_error "Data consistency test failed - cannot create config"
    fi
}

# 主测试函数
run_all_tests() {
    log_info "Starting OpenList Workers API Comprehensive Tests..."
    echo "Base URL: $BASE_URL"
    echo "================================================"

    # 设置测试环境
    setup_test_env

    # 运行测试
    test_system_health
    test_user_auth
    test_user_management
    test_driver_management
    test_filesystem_operations
    test_offline_download
    test_error_handling
    test_data_consistency

    # 清理测试环境
    cleanup_test_env

    # 输出测试结果
    echo "================================================"
    echo -e "Test Summary:"
    echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "${GREEN}All tests passed! 🎉${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed! 😞${NC}"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "OpenList Workers API 自动化测试脚本"
    echo ""
    echo "用法: $0 [OPTIONS]"
    echo ""
    echo "选项:"
    echo "  -u, --url URL       设置API基础URL (默认: http://localhost:8787)"
    echo "  -h, --help         显示此帮助信息"
    echo ""
    echo "测试模块:"
    echo "  system             系统健康检查和初始化"
    echo "  auth               用户认证"
    echo "  user               用户管理"
    echo "  driver             驱动配置管理"
    echo "  filesystem         文件系统操作"
    echo "  offline            离线下载"
    echo "  error              错误处理"
    echo "  consistency        数据一致性"
    echo ""
    echo "示例:"
    echo "  $0                          # 运行所有测试"
    echo "  $0 -u http://myapi.com      # 使用自定义URL"
    echo "  $0 auth                     # 只运行认证测试"
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
        system)
            setup_test_env
            test_system_health
            cleanup_test_env
            exit 0
            ;;
        auth)
            setup_test_env
            test_user_auth
            cleanup_test_env
            exit 0
            ;;
        user)
            setup_test_env
            test_user_auth  # 需要token
            test_user_management
            cleanup_test_env
            exit 0
            ;;
        driver)
            setup_test_env
            test_user_auth  # 需要token
            test_driver_management
            cleanup_test_env
            exit 0
            ;;
        filesystem)
            setup_test_env
            test_user_auth  # 需要token
            test_driver_management  # 需要driver config
            test_filesystem_operations
            cleanup_test_env
            exit 0
            ;;
        offline)
            setup_test_env
            test_user_auth  # 需要token
            test_offline_download
            cleanup_test_env
            exit 0
            ;;
        error)
            setup_test_env
            test_user_auth  # 需要token
            test_error_handling
            cleanup_test_env
            exit 0
            ;;
        consistency)
            setup_test_env
            test_user_auth  # 需要token
            test_data_consistency
            cleanup_test_env
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# 运行所有测试
run_all_tests