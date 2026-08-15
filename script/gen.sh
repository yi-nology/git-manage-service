#!/bin/bash
# script/gen.sh - 代码生成脚本
# 生成 Kitex RPC 代码和 Hz HTTP 代码

set -e

PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$PROJECT_ROOT"

echo "=========================================="
echo "  Git Manage Service - Code Generator"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PROTOC_GEN_GO_VERSION="${PROTOC_GEN_GO_VERSION:-v1.28.0}"
HZ_VERSION="${HZ_VERSION:-v0.9.7}"
KITEX_VERSION="${KITEX_VERSION:-v0.15.4}"

# 打印带颜色的消息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查工具是否安装
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        error "$1 is not installed. Please install it first."
        return 1
    fi
    return 0
}

# 1. 检查依赖工具
info "Checking required tools..."
MISSING_TOOLS=()

if ! check_tool "protoc"; then
    MISSING_TOOLS+=("protoc")
fi

if [ "$REQUIRE_CODEGEN_TOOLS" = "1" ] && ! check_tool "protoc-gen-go"; then
    MISSING_TOOLS+=("protoc-gen-go")
fi

if ! check_tool "kitex"; then
    if [ "$REQUIRE_CODEGEN_TOOLS" = "1" ]; then
        MISSING_TOOLS+=("kitex")
    else
        warn "kitex not found. Skipping Kitex code generation."
        warn "Install with: go install github.com/cloudwego/kitex/tool/cmd/kitex@$KITEX_VERSION"
        SKIP_KITEX=true
    fi
fi

if ! check_tool "hz"; then
    if [ "$REQUIRE_CODEGEN_TOOLS" = "1" ]; then
        MISSING_TOOLS+=("hz")
    else
        warn "hz not found. Skipping Hz code generation."
        warn "Install with: go install github.com/cloudwego/hertz/cmd/hz@$HZ_VERSION"
        SKIP_HZ=true
    fi
fi

if [ ${#MISSING_TOOLS[@]} -ne 0 ]; then
    error "Missing required tools: ${MISSING_TOOLS[*]}"
    exit 1
fi

# 2. 生成 Kitex RPC 代码
if [ "$SKIP_KITEX" != "true" ]; then
    info "Generating Kitex RPC code..."
    cd biz
    # NOTE: no `-service` flag — that generates the scaffold main.go/handler.go
    # under biz/, which this project does not use (the RPC server lives in
    # biz/rpc_handler + cmd/server). We only need the kitex_gen/ package.
    kitex -module github.com/yi-nology/git-manage-service \
          -I ../idl \
          ../idl/git.proto
    cd ..
    info "Kitex code generation completed"
else
    warn "Skipped Kitex code generation"
fi

# 3. 生成 Hz HTTP 代码（如果有 biz proto 文件）
if [ "$SKIP_HZ" != "true" ]; then
    if ls idl/biz/*.proto 1> /dev/null 2>&1; then
        info "Generating Hz HTTP code..."

        # 检查是否已初始化 Hz（当前项目使用 .hz 记录生成目录）
        if [ ! -f ".hz" ]; then
            info "Initializing Hz project..."
            hz new -idl idl/biz/repo.proto \
                -I idl \
                -module github.com/yi-nology/git-manage-service
        fi

        # sync 路由被手动从 v1 stub 改指向 v2 handler（git-sync-service）。
        # hz update 会把它改回 v1 并生成空 stub —— 备份、生成后恢复。
        SYNC_ROUTER_BAK="$(mktemp -d)"
        cp -r biz/router/sync "$SYNC_ROUTER_BAK/router_sync" 2>/dev/null || true
        rm -f biz/handler/sync/sync_service.go 2>/dev/null || true

        # 更新生成代码
        for proto in idl/biz/*.proto; do
            info "Processing $proto..."
            hz update -idl "$proto" \
                -I idl \
                --snake_tag || {
                if [ "$REQUIRE_CODEGEN_TOOLS" = "1" ]; then
                    error "Failed to process $proto"
                    exit 1
                fi
                warn "Failed to process $proto"
            }
        done

        # 恢复 v2 sync 路由，删除 hz 生成的 v1 stub。
        if [ -d "$SYNC_ROUTER_BAK/router_sync" ]; then
            rm -rf biz/router/sync
            cp -r "$SYNC_ROUTER_BAK/router_sync" biz/router/sync
        fi
        rm -f biz/handler/sync/sync_service.go 2>/dev/null || true
        rm -rf "$SYNC_ROUTER_BAK"

        info "Hz code generation completed"
    else
        warn "No proto files found in idl/biz/. Skipping Hz code generation."
    fi
else
    warn "Skipped Hz code generation"
fi

# 4. 整理 Go 依赖
info "Tidying Go modules..."
go mod tidy

# 5. 清理 hz update 生成的问题
info "Cleaning up generated code..."

# (5a removed: the swagger-generate consts import no longer exists in any
#  handler — and the old `sed -i ''` was BSD-only, breaking Linux CI.)

# 5b. 删除 hz 生成的重复 stub 函数（已有真实实现的文件）
if grep -q 'func ListLLMPresets' biz/handler/settings/settings_service.go 2>/dev/null; then
    awk '/^\/\/ ListLLMPresets \./{skip=1} skip && /^}/{skip=0; next} !skip' biz/handler/settings/settings_service.go > biz/handler/settings/settings_service.go.tmp && mv biz/handler/settings/settings_service.go.tmp biz/handler/settings/settings_service.go
    info "Removed duplicate ListLLMPresets from settings_service.go"
fi

# 5c. 清理 unused imports（删除函数后可能产生）
go install golang.org/x/tools/cmd/goimports@latest 2>/dev/null || true
goimports -w biz/handler/ 2>/dev/null || true

# 6. 格式化代码
info "Formatting code..."
go fmt ./...

echo ""
echo "=========================================="
info "Code generation completed successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Review generated code in biz/handler and biz/router"
echo "  2. Implement business logic in generated handlers"
echo "  3. Run 'make build' to compile the project"
echo "  4. Run 'make run' to start the service"
