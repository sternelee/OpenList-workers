#!/bin/bash

# OpenList Workers API 性能测试脚本
# 测试API响应时间、并发性能和负载能力

set -e

# 配置
BASE_URL="http://localhost:8787"
CONCURRENT_USERS=10
REQUESTS_PER_USER=5
TIMEOUT=30

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果存储
declare -a RESPONSE_TIMES
declare -a STATUS_CODES
TOTAL_REQUESTS=0
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0

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

# 获取当前时间戳（毫秒）
get_timestamp_ms() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        python3 -c "import time; print(int(time.time() * 1000))"
    else
        # Linux
        date +%s%3N
    fi
}

# 测试单个API端点
test_endpoint() {
    local method="$1"
    local url="$2"
    local data="$3"
    local token="$4"
    local expected_status="$5"

    local start_time=$(get_timestamp_ms)

    local headers=(-H "Content-Type: application/json")
    if [[ -n "$token" ]]; then
        headers+=(-H "Authorization: Bearer $token")
    fi

    local response
    if [[ "$method" == "GET" ]]; then
        response=$(curl -s -w "%{http_code}:%{time_total}" --max-time $TIMEOUT "${headers[@]}" "$BASE_URL$url" 2>/dev/null || echo "000:$TIMEOUT")
    else
        response=$(curl -s -w "%{http_code}:%{time_total}" --max-time $TIMEOUT -X "$method" "${headers[@]}" -d "$data" "$BASE_URL$url" 2>/dev/null || echo "000:$TIMEOUT")
    fi

    local end_time=$(get_timestamp_ms)
    local total_time=$((end_time - start_time))

    local status_code=$(echo "$response" | tail -c 20 | cut -d: -f1)
    local curl_time=$(echo "$response" | tail -c 20 | cut -d: -f2)

    RESPONSE_TIMES+=($total_time)
    STATUS_CODES+=($status_code)
    TOTAL_REQUESTS=$((TOTAL_REQUESTS + 1))

    if [[ "$status_code" == "$expected_status" ]]; then
        SUCCESSFUL_REQUESTS=$((SUCCESSFUL_REQUESTS + 1))
        echo "${total_time}ms:${status_code}:SUCCESS"
    else
        FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
        echo "${total_time}ms:${status_code}:FAILED"
    fi
}

# 获取认证Token
get_auth_token() {
    local response=$(curl -s -X POST -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' \
        "$BASE_URL/api/auth/login" 2>/dev/null)

    if [[ -n "$response" ]]; then
        echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
    else
        echo ""
    fi
}

# 并发用户测试
concurrent_user_test() {
    local endpoint="$1"
    local method="$2"
    local data="$3"
    local token="$4"
    local expected_status="$5"

    log_info "Running concurrent test for $method $endpoint"
    log_info "Concurrent users: $CONCURRENT_USERS, Requests per user: $REQUESTS_PER_USER"

    local pids=()
    local temp_files=()

    for ((i=1; i<=CONCURRENT_USERS; i++)); do
        local temp_file=$(mktemp)
        temp_files+=("$temp_file")

        {
            for ((j=1; j<=REQUESTS_PER_USER; j++)); do
                test_endpoint "$method" "$endpoint" "$data" "$token" "$expected_status"
            done
        } > "$temp_file" &

        pids+=($!)
    done

    # 等待所有进程完成
    for pid in "${pids[@]}"; do
        wait $pid
    done

    # 收集结果
    for temp_file in "${temp_files[@]}"; do
        while IFS= read -r line; do
            echo "$line"
        done < "$temp_file"
        rm -f "$temp_file"
    done
}

# 计算统计信息
calculate_stats() {
    if [[ ${#RESPONSE_TIMES[@]} -eq 0 ]]; then
        echo "No response times recorded"
        return
    fi

    # 排序响应时间
    local sorted_times=($(printf '%s\n' "${RESPONSE_TIMES[@]}" | sort -n))

    # 计算基本统计
    local total=0
    local min=${sorted_times[0]}
    local max=${sorted_times[-1]}

    for time in "${sorted_times[@]}"; do
        total=$((total + time))
    done

    local avg=$((total / ${#sorted_times[@]}))

    # 计算百分位数
    local count=${#sorted_times[@]}
    local p50_index=$((count * 50 / 100))
    local p90_index=$((count * 90 / 100))
    local p95_index=$((count * 95 / 100))
    local p99_index=$((count * 99 / 100))

    local p50=${sorted_times[$p50_index]}
    local p90=${sorted_times[$p90_index]}
    local p95=${sorted_times[$p95_index]}
    local p99=${sorted_times[$p99_index]}

    # 计算请求/秒
    local total_time_seconds=$((max / 1000))
    if [[ $total_time_seconds -eq 0 ]]; then
        total_time_seconds=1
    fi
    local rps=$((TOTAL_REQUESTS / total_time_seconds))

    echo "================================================"
    echo "Performance Test Results:"
    echo "================================================"
    echo "Total Requests: $TOTAL_REQUESTS"
    echo "Successful: $SUCCESSFUL_REQUESTS"
    echo "Failed: $FAILED_REQUESTS"
    echo "Success Rate: $(((SUCCESSFUL_REQUESTS * 100) / TOTAL_REQUESTS))%"
    echo ""
    echo "Response Times (ms):"
    echo "  Min: $min"
    echo "  Max: $max"
    echo "  Average: $avg"
    echo "  P50: $p50"
    echo "  P90: $p90"
    echo "  P95: $p95"
    echo "  P99: $p99"
    echo ""
    echo "Throughput: ~$rps req/sec"
    echo "================================================"
}

# 清理测试数据
cleanup() {
    RESPONSE_TIMES=()
    STATUS_CODES=()
    TOTAL_REQUESTS=0
    SUCCESSFUL_REQUESTS=0
    FAILED_REQUESTS=0
}

# 基础性能测试
test_basic_performance() {
    log_info "Running basic performance tests..."

    # 测试健康检查
    log_info "Testing health endpoint..."
    cleanup
    concurrent_user_test "/health" "GET" "" "" "200"
    calculate_stats

    # 测试系统初始化
    log_info "Testing init endpoint..."
    cleanup
    concurrent_user_test "/init" "GET" "" "" "200"
    calculate_stats
}

# 认证性能测试
test_auth_performance() {
    log_info "Running authentication performance tests..."

    # 测试登录
    log_info "Testing login endpoint..."
    cleanup
    local login_data='{"username":"admin","password":"admin123"}'
    concurrent_user_test "/api/auth/login" "POST" "$login_data" "" "200"
    calculate_stats
}

# API性能测试（需要认证）
test_api_performance() {
    log_info "Running authenticated API performance tests..."

    # 获取认证token
    local token=$(get_auth_token)
    if [[ -z "$token" ]]; then
        log_error "Failed to get authentication token"
        return 1
    fi

    # 测试用户信息
    log_info "Testing user info endpoint..."
    cleanup
    concurrent_user_test "/api/auth/me" "GET" "" "$token" "200"
    calculate_stats

    # 测试驱动列表
    log_info "Testing drivers list endpoint..."
    cleanup
    concurrent_user_test "/api/drivers" "GET" "" "$token" "200"
    calculate_stats

    # 测试用户驱动配置
    log_info "Testing user driver configs endpoint..."
    cleanup
    concurrent_user_test "/api/user/driver/list" "GET" "" "$token" "200"
    calculate_stats
}

# 压力测试
test_stress() {
    log_info "Running stress tests..."

    local original_concurrent=$CONCURRENT_USERS
    local original_requests=$REQUESTS_PER_USER

    # 增加并发数
    CONCURRENT_USERS=20
    REQUESTS_PER_USER=10

    log_info "Stress testing with $CONCURRENT_USERS concurrent users, $REQUESTS_PER_USER requests each"

    # 测试健康检查压力
    cleanup
    concurrent_user_test "/health" "GET" "" "" "200"
    calculate_stats

    # 恢复原始设置
    CONCURRENT_USERS=$original_concurrent
    REQUESTS_PER_USER=$original_requests
}

# 错误处理性能测试
test_error_performance() {
    log_info "Running error handling performance tests..."

    # 测试404错误
    log_info "Testing 404 error handling..."
    cleanup
    concurrent_user_test "/nonexistent" "GET" "" "" "404"
    calculate_stats

    # 测试401错误
    log_info "Testing 401 error handling..."
    cleanup
    concurrent_user_test "/api/auth/me" "GET" "" "invalid_token" "401"
    calculate_stats
}

# 显示帮助信息
show_help() {
    echo "OpenList Workers API 性能测试脚本"
    echo ""
    echo "用法: $0 [OPTIONS] [TEST_TYPE]"
    echo ""
    echo "选项:"
    echo "  -u, --url URL           设置API基础URL (默认: http://localhost:8787)"
    echo "  -c, --concurrent N      并发用户数 (默认: 10)"
    echo "  -r, --requests N        每用户请求数 (默认: 5)"
    echo "  -t, --timeout N         请求超时时间（秒） (默认: 30)"
    echo "  -h, --help             显示此帮助信息"
    echo ""
    echo "测试类型:"
    echo "  basic                  基础性能测试"
    echo "  auth                   认证性能测试"
    echo "  api                    API性能测试"
    echo "  stress                 压力测试"
    echo "  error                  错误处理性能测试"
    echo "  all                    运行所有测试 (默认)"
    echo ""
    echo "示例:"
    echo "  $0                                  # 运行所有性能测试"
    echo "  $0 -c 20 -r 10 stress              # 运行压力测试"
    echo "  $0 -u http://myapi.com basic        # 测试基础性能"
}

# 主函数
main() {
    local test_type="all"

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--url)
                BASE_URL="$2"
                shift 2
                ;;
            -c|--concurrent)
                CONCURRENT_USERS="$2"
                shift 2
                ;;
            -r|--requests)
                REQUESTS_PER_USER="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            basic|auth|api|stress|error|all)
                test_type="$1"
                shift
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    log_info "Starting OpenList Workers API Performance Tests..."
    echo "Base URL: $BASE_URL"
    echo "Concurrent Users: $CONCURRENT_USERS"
    echo "Requests per User: $REQUESTS_PER_USER"
    echo "Timeout: ${TIMEOUT}s"
    echo "================================================"

    case $test_type in
        basic)
            test_basic_performance
            ;;
        auth)
            test_auth_performance
            ;;
        api)
            test_api_performance
            ;;
        stress)
            test_stress
            ;;
        error)
            test_error_performance
            ;;
        all)
            test_basic_performance
            test_auth_performance
            test_api_performance
            test_stress
            test_error_performance
            ;;
    esac

    log_success "Performance tests completed!"
}

main "$@"