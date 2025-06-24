#!/bin/bash

# OpenList Workers 完整测试套件运行器
# 运行所有类型的API测试

set -e

# 配置
BASE_URL="http://localhost:8787"
REPORT_DIR="test_reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 创建报告目录
create_report_dir() {
    mkdir -p "$REPORT_DIR"
    log_info "Created report directory: $REPORT_DIR"
}

# 检查服务器是否运行
check_server() {
    log_info "Checking if server is running at $BASE_URL..."

    local response
    if response=$(curl -s -w "%{http_code}" "$BASE_URL/health" --max-time 10 2>/dev/null); then
        local status_code=$(echo "$response" | tail -c 4)
        if [[ "$status_code" == "200" ]]; then
            log_success "Server is running and healthy"
            return 0
        else
            log_error "Server returned status code: $status_code"
            return 1
        fi
    else
        log_error "Cannot connect to server at $BASE_URL"
        return 1
    fi
}

# 运行测试脚本
run_test() {
    local test_name="$1"
    local test_script="$2"
    local test_args="$3"
    local report_file="$REPORT_DIR/${test_name}_${TIMESTAMP}.log"

    log_info "Running $test_name..."

    if [[ ! -f "$test_script" ]]; then
        log_error "Test script not found: $test_script"
        return 1
    fi

    if [[ ! -x "$test_script" ]]; then
        chmod +x "$test_script"
        log_info "Made $test_script executable"
    fi

    # 运行测试并记录输出
    if "./$test_script" $test_args > "$report_file" 2>&1; then
        log_success "$test_name passed"
        return 0
    else
        log_error "$test_name failed"
        log_error "See report: $report_file"
        return 1
    fi
}

# 运行基础API测试
run_basic_tests() {
    log_info "=== Running Basic API Tests ==="

    if run_test "comprehensive_api_tests" "test_api_comprehensive.sh" "-u $BASE_URL"; then
        return 0
    else
        return 1
    fi
}

# 运行性能测试
run_performance_tests() {
    log_info "=== Running Performance Tests ==="

    if run_test "performance_tests" "test_performance.sh" "-u $BASE_URL -c 5 -r 3"; then
        return 0
    else
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    log_info "=== Running Integration Tests ==="

    if run_test "integration_tests" "test_integration.sh" "-u $BASE_URL"; then
        return 0
    else
        return 1
    fi
}

# 运行认证测试
run_auth_tests() {
    log_info "=== Running Authentication Tests ==="

    if [[ -f "test_auth_api.sh" ]]; then
        if run_test "auth_tests" "test_auth_api.sh" "$BASE_URL"; then
            return 0
        else
            return 1
        fi
    else
        log_warning "Authentication test script not found (test_auth_api.sh)"
        return 0
    fi
}

# 生成测试报告汇总
generate_summary_report() {
    local summary_file="$REPORT_DIR/test_summary_${TIMESTAMP}.html"

    log_info "Generating test summary report..."

    cat > "$summary_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>OpenList Workers API Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .test-result { margin: 10px 0; padding: 10px; border-radius: 5px; }
        .passed { background-color: #d4edda; color: #155724; }
        .failed { background-color: #f8d7da; color: #721c24; }
        .warning { background-color: #fff3cd; color: #856404; }
        .details { margin-top: 20px; }
        pre { background-color: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>OpenList Workers API Test Report</h1>
        <p><strong>Generated:</strong> $(date)</p>
        <p><strong>Base URL:</strong> $BASE_URL</p>
        <p><strong>Test Run ID:</strong> $TIMESTAMP</p>
    </div>

    <div class="details">
        <h2>Test Results</h2>
EOF

    # 添加各个测试的结果
    for log_file in "$REPORT_DIR"/*_${TIMESTAMP}.log; do
        if [[ -f "$log_file" ]]; then
            local test_name=$(basename "$log_file" | sed "s/_${TIMESTAMP}.log//")
            local last_line=$(tail -n 1 "$log_file" 2>/dev/null || echo "")

            if echo "$last_line" | grep -q "passed\|SUCCESS"; then
                echo "        <div class=\"test-result passed\">" >> "$summary_file"
                echo "            <h3>✅ $test_name</h3>" >> "$summary_file"
                echo "            <p>Status: PASSED</p>" >> "$summary_file"
            else
                echo "        <div class=\"test-result failed\">" >> "$summary_file"
                echo "            <h3>❌ $test_name</h3>" >> "$summary_file"
                echo "            <p>Status: FAILED</p>" >> "$summary_file"
            fi

            echo "            <details>" >> "$summary_file"
            echo "                <summary>View Details</summary>" >> "$summary_file"
            echo "                <pre>$(cat "$log_file" | tail -n 50)</pre>" >> "$summary_file"
            echo "            </details>" >> "$summary_file"
            echo "        </div>" >> "$summary_file"
        fi
    done

    cat >> "$summary_file" << EOF
    </div>

    <div class="details">
        <h2>System Information</h2>
        <pre>
OS: $(uname -s) $(uname -r)
Date: $(date)
User: $(whoami)
Working Directory: $(pwd)
        </pre>
    </div>
</body>
</html>
EOF

    log_success "Test summary report generated: $summary_file"
}

# 显示使用帮助
show_help() {
    echo "OpenList Workers API 测试套件运行器"
    echo ""
    echo "用法: $0 [OPTIONS] [TEST_TYPE]"
    echo ""
    echo "选项:"
    echo "  -u, --url URL          设置API基础URL (默认: http://localhost:8787)"
    echo "  -r, --report-dir DIR   设置报告目录 (默认: test_reports)"
    echo "  -h, --help            显示此帮助信息"
    echo ""
    echo "测试类型:"
    echo "  all                   运行所有测试 (默认)"
    echo "  basic                 基础API功能测试"
    echo "  performance           性能测试"
    echo "  integration           集成测试"
    echo "  auth                  认证测试"
    echo ""
    echo "示例:"
    echo "  $0                           # 运行所有测试"
    echo "  $0 basic                     # 只运行基础测试"
    echo "  $0 -u http://api.example.com # 使用自定义URL"
    echo "  $0 -r /tmp/reports           # 使用自定义报告目录"
}

# 主函数
main() {
    local test_type="all"
    local total_tests=0
    local passed_tests=0
    local failed_tests=0

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--url)
                BASE_URL="$2"
                shift 2
                ;;
            -r|--report-dir)
                REPORT_DIR="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            all|basic|performance|integration|auth)
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

    echo "=========================================="
    echo "OpenList Workers API Test Suite"
    echo "=========================================="
    echo "Base URL: $BASE_URL"
    echo "Report Directory: $REPORT_DIR"
    echo "Test Type: $test_type"
    echo "Timestamp: $TIMESTAMP"
    echo "=========================================="

    # 创建报告目录
    create_report_dir

    # 检查服务器状态
    if ! check_server; then
        log_error "Server check failed. Please ensure the OpenList Workers server is running."
        exit 1
    fi

    # 根据测试类型运行相应测试
    case $test_type in
        all)
            log_info "Running all test suites..."

            ((total_tests++))
            if run_basic_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi

            ((total_tests++))
            if run_auth_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi

            ((total_tests++))
            if run_integration_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi

            ((total_tests++))
            if run_performance_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
        basic)
            ((total_tests++))
            if run_basic_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
        auth)
            ((total_tests++))
            if run_auth_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
        integration)
            ((total_tests++))
            if run_integration_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
        performance)
            ((total_tests++))
            if run_performance_tests; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
    esac

    # 生成汇总报告
    generate_summary_report

    # 输出最终结果
    echo "=========================================="
    echo "Test Suite Summary"
    echo "=========================================="
    echo "Total Test Suites: $total_tests"
    echo "Passed: $passed_tests"
    echo "Failed: $failed_tests"
    echo "Success Rate: $(( (passed_tests * 100) / total_tests ))%"
    echo "Report Directory: $REPORT_DIR"
    echo "=========================================="

    if [[ $failed_tests -eq 0 ]]; then
        log_success "All test suites passed! 🎉"
        exit 0
    else
        log_error "$failed_tests test suite(s) failed!"
        exit 1
    fi
}

main "$@"