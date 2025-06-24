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
    log_info "=== Scenario 1: New User Complete Journey ==="

    # 系统初始化
    log_info "Initializing system..."
    local response=$(make_request "GET" "/init")
    if parse_response "$response" "200" > /dev/null; then
        log_success "System initialized"
    else
        log_error "System initialization failed"
        return 1
    fi

    # 用户注册
    log_info "Registering new user..."
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

    # 配置驱动
    log_info "Configuring storage driver..."
    local driver_data='{"name":"Local","display_name":"测试本地存储","description":"集成测试用本地存储","config":"{\"root_folder_path\": \"/tmp/test\"}","icon":"folder","enabled":true,"order":1}'
    response=$(make_request "POST" "/api/user/driver/create" "$driver_data" "$USER_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "Storage driver configured"
    else
        log_error "Failed to configure storage driver"
        return 1
    fi

    log_success "Scenario 1 completed successfully"
}

# 场景2: 管理员操作
scenario_admin_operations() {
    log_info "=== Scenario 2: Admin Operations ==="

    # 管理员登录
    log_info "Admin login..."
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

    # 查看用户列表
    log_info "Listing users..."
    response=$(make_request "GET" "/api/admin/user/list" "" "$ADMIN_TOKEN")
    if parse_response "$response" "200" > /dev/null; then
        log_success "User list retrieved"
    else
        log_error "Failed to retrieve user list"
        return 1
    fi

    log_success "Scenario 2 completed successfully"
}

# 主函数
main() {
    log_info "Starting OpenList Workers Integration Tests..."
    echo "Base URL: $BASE_URL"
    echo "================================================"

    local failed_scenarios=0

    # 运行测试场景
    if ! scenario_new_user_journey; then
        ((failed_scenarios++))
    fi

    if ! scenario_admin_operations; then
        ((failed_scenarios++))
    fi

    echo "================================================"
    echo "Integration Test Summary:"
    echo "Total Scenarios: 2"
    echo "Passed: $((2 - failed_scenarios))"
    echo "Failed: $failed_scenarios"

    if [[ $failed_scenarios -eq 0 ]]; then
        log_success "All integration tests passed! 🎉"
        exit 0
    else
        log_error "$failed_scenarios scenario(s) failed!"
        exit 1
    fi
}

# 显示帮助
show_help() {
    echo "OpenList Workers API 集成测试脚本"
    echo ""
    echo "用法: $0 [OPTIONS]"
    echo ""
    echo "选项:"
    echo "  -u, --url URL       设置API基础URL (默认: http://localhost:8787)"
    echo "  -h, --help         显示此帮助信息"
}

# 解析参数
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