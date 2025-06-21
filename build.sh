#!/bin/bash

# OpenList Workers 构建脚本
# 使用方法: ./build.sh [command]
# 命令: build, deploy, dev, test

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PROJECT_NAME="openlist-workers"
BUILD_DIR="dist"
MAIN_FILE="main.go"

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}Checking dependencies...${NC}"

    # 检查 Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi

    # 检查 Wrangler
    if ! command -v wrangler &> /dev/null; then
        echo -e "${RED}Error: Wrangler CLI is not installed${NC}"
        echo "Install with: npm install -g wrangler"
        exit 1
    fi

    echo -e "${GREEN}✓ All dependencies are installed${NC}"
}

# 构建项目
build() {
    echo -e "${BLUE}Building $PROJECT_NAME...${NC}"

    # 创建构建目录
    mkdir -p $BUILD_DIR

    # 下载依赖
    echo "Downloading dependencies..."
    go mod download

    # 构建
    echo "Compiling..."
    go build -o $BUILD_DIR/worker $MAIN_FILE

    echo -e "${GREEN}✓ Build completed successfully${NC}"
    echo "Output: $BUILD_DIR/worker"
}

# 部署到 Cloudflare Workers
deploy() {
    echo -e "${BLUE}Deploying to Cloudflare Workers...${NC}"

    # 检查 wrangler.toml 是否存在
    if [ ! -f "wrangler.toml" ]; then
        echo -e "${RED}Error: wrangler.toml not found${NC}"
        exit 1
    fi

    # 部署
    wrangler deploy

    echo -e "${GREEN}✓ Deployment completed successfully${NC}"
}

# 本地开发
dev() {
    echo -e "${BLUE}Starting local development server...${NC}"

    # 检查 wrangler.toml 是否存在
    if [ ! -f "wrangler.toml" ]; then
        echo -e "${RED}Error: wrangler.toml not found${NC}"
        exit 1
    fi

    # 启动开发服务器
    wrangler dev
}

# 运行测试
test() {
    echo -e "${BLUE}Running tests...${NC}"

    # 运行 Go 测试
    go test ./...

    # 运行 API 测试
    if [ -f "test_api.sh" ]; then
        echo -e "${BLUE}Running API tests...${NC}"
        chmod +x test_api.sh
        ./test_api.sh
    fi

    echo -e "${GREEN}✓ Tests completed${NC}"
}

# 清理构建文件
clean() {
    echo -e "${BLUE}Cleaning build files...${NC}"

    if [ -d "$BUILD_DIR" ]; then
        rm -rf $BUILD_DIR
        echo -e "${GREEN}✓ Cleaned $BUILD_DIR${NC}"
    fi

    # 清理 Go 缓存
    go clean -cache
    echo -e "${GREEN}✓ Cleaned Go cache${NC}"
}

# 显示帮助信息
show_help() {
    echo -e "${YELLOW}OpenList Workers Build Script${NC}"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build   - Build the project"
    echo "  deploy  - Deploy to Cloudflare Workers"
    echo "  dev     - Start local development server"
    echo "  test    - Run tests"
    echo "  clean   - Clean build files"
    echo "  help    - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build"
    echo "  $0 deploy"
    echo "  $0 dev"
    echo "  $0 test"
}

# 主函数
main() {
    case "${1:-help}" in
        "build")
            check_dependencies
            build
            ;;
        "deploy")
            check_dependencies
            build
            deploy
            ;;
        "dev")
            check_dependencies
            dev
            ;;
        "test")
            check_dependencies
            test
            ;;
        "clean")
            clean
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# 运行主函数
main "$@"
