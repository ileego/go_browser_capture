#!/bin/bash

# Go Chrome Project Build Script

echo "🚀 Go Chrome Project Build Script"
echo "=================================="

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 函数：打印成功消息
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 函数：打印警告消息
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 函数：打印错误消息
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# 检查Go是否安装
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
    fi
    print_success "Go is installed: $(go version)"
}

# 初始化后端项目
init_backend() {
    echo ""
    echo "📦 Initializing backend project..."
    cd backend
    
    # 初始化Go模块
    if [ ! -f "go.sum" ]; then
        print_warning "Go modules not initialized. Running 'go mod tidy'..."
        go mod tidy
        if [ $? -eq 0 ]; then
            print_success "Go modules initialized successfully"
        else
            print_error "Failed to initialize Go modules"
            exit 1
        fi
    else
        print_success "Go modules already initialized"
    fi
    
    # 生成Swagger文档
    print_warning "Generating Swagger documentation..."
    if command -v swag &> /dev/null; then
        swag init -g main.go
        if [ $? -eq 0 ]; then
            print_success "Swagger documentation generated"
        else
            print_warning "Failed to generate Swagger documentation"
        fi
    else
        print_warning "Swagger CLI not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"
    fi
    
    cd ..
}

# 构建后端项目
build_backend() {
    echo ""
    echo "🔨 Building backend project..."
    cd backend
    
    # 创建构建目录
    mkdir -p ../dist
    
    # 构建可执行文件
    print_warning "Building backend executable..."
    go build -o ../dist/go_chrome_server main.go
    
    if [ $? -eq 0 ]; then
        print_success "Backend built successfully: dist/go_chrome_server"
    else
        print_error "Failed to build backend"
        exit 1
    fi
    
    cd ..
}

# 运行后端项目
run_backend() {
    echo ""
    echo "▶️  Starting backend server..."
    cd backend
    go run main.go
}

# 构建Chrome插件
build_extension() {
    echo ""
    echo "🔨 Building Chrome extension..."
    
    # 创建构建目录
    mkdir -p dist
    
    # 复制插件文件到构建目录
    cp -r extension/* dist/extension/
    
    print_success "Chrome extension built successfully: dist/extension/"
}

# 显示帮助信息
show_help() {
    echo ""
    echo "Usage: ./build.sh [command]"
    echo ""
    echo "Commands:"
    echo "  init       Initialize the project (install dependencies)"
    echo "  build      Build both backend and extension"
    echo "  backend    Build only the backend"
    echo "  extension  Build only the extension"
    echo "  run        Run the backend server"
    echo "  help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./build.sh init     # Initialize project"
    echo "  ./build.sh build    # Build everything"
    echo "  ./build.sh run      # Run backend server"
}

# 主逻辑
main() {
    check_go
    
    case "${1:-help}" in
        init)
            init_backend
            ;;
        build)
            build_backend
            build_extension
            ;;
        backend)
            build_backend
            ;;
        extension)
            build_extension
            ;;
        run)
            run_backend
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
    
    echo ""
    print_success "Build completed successfully!"
}

# 运行主函数
main "$@"