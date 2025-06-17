#!/bin/bash

# OpenList 数据库迁移脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否安装了 wrangler
check_wrangler() {
    if ! command -v wrangler &> /dev/null; then
        print_error "wrangler CLI 未安装。请运行: npm install -g wrangler"
        exit 1
    fi
}

# 检查数据库 ID 是否配置
check_database_config() {
    if grep -q "your-database-id-here" wrangler.toml; then
        print_warning "请先在 wrangler.toml 中配置正确的 database_id"
        print_warning "运行以下命令创建数据库:"
        print_warning "  npx wrangler d1 create openlist-db"
        exit 1
    fi
}

# 应用迁移
apply_migrations() {
    local env=$1
    
    print_message "开始应用数据库迁移..."
    
    if [ "$env" = "local" ]; then
        print_message "应用到本地开发环境..."
        npx wrangler d1 migrations apply openlist-db --local
    else
        print_message "应用到生产环境..."
        npx wrangler d1 migrations apply openlist-db
    fi
    
    if [ $? -eq 0 ]; then
        print_message "数据库迁移完成!"
    else
        print_error "数据库迁移失败!"
        exit 1
    fi
}

# 显示数据库信息
show_database_info() {
    local env=$1
    
    print_message "数据库信息:"
    
    if [ "$env" = "local" ]; then
        npx wrangler d1 info openlist-db --local
    else
        npx wrangler d1 info openlist-db
    fi
}

# 执行 SQL 查询（用于测试）
execute_query() {
    local env=$1
    local query=$2
    
    if [ "$env" = "local" ]; then
        npx wrangler d1 execute openlist-db --local --command="$query"
    else
        npx wrangler d1 execute openlist-db --command="$query"
    fi
}

# 显示帮助信息
show_help() {
    echo "OpenList 数据库迁移脚本"
    echo ""
    echo "用法:"
    echo "  $0 [选项]"
    echo ""
    echo "选项:"
    echo "  migrate [local]    应用数据库迁移 (默认生产环境，添加 local 为本地环境)"
    echo "  info [local]       显示数据库信息"
    echo "  test [local]       测试数据库连接"
    echo "  help              显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 migrate         # 应用迁移到生产环境"
    echo "  $0 migrate local   # 应用迁移到本地环境"
    echo "  $0 info local      # 显示本地数据库信息"
    echo "  $0 test local      # 测试本地数据库连接"
}

# 测试数据库连接
test_connection() {
    local env=$1
    
    print_message "测试数据库连接..."
    
    # 执行简单查询测试连接
    if execute_query "$env" "SELECT 1;"; then
        print_message "数据库连接成功!"
        
        # 检查用户表是否存在
        print_message "检查数据表..."
        execute_query "$env" "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;"
        
        # 显示用户数量
        print_message "用户数量:"
        execute_query "$env" "SELECT COUNT(*) as user_count FROM users;"
        
    else
        print_error "数据库连接失败!"
        exit 1
    fi
}

# 主函数
main() {
    local command=${1:-help}
    local env=${2:-production}
    
    # 转到项目根目录
    cd "$(dirname "$0")/.."
    
    case $command in
        "migrate")
            check_wrangler
            if [ "$env" != "local" ]; then
                check_database_config
            fi
            apply_migrations "$env"
            ;;
        "info")
            check_wrangler
            show_database_info "$env"
            ;;
        "test")
            check_wrangler
            test_connection "$env"
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# 运行主函数
main "$@" 