# ============================================================================
# 系统初始化检查工具 - Makefile
# ============================================================================

.PHONY: help build build-verbose build-debug build-full build-no-clean clean clean-logs clean-all run run-debug test test-unit test-integration test-coverage test-version test-logger test-cmd bench install fmt lint mod dev info version-info

# ============================================================================
# 变量定义
# ============================================================================
BINARY_NAME=sysint
BINARY_DIR=bin
BUILD_LOG_DIR=build_logs
BUILD_LOG=$(BUILD_LOG_DIR)/build_$(shell date +%Y%m%d_%H%M%S).log

# 版本信息
VERSION?=dev
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null)
ifeq ($(GIT_COMMIT),)
    GIT_COMMIT=unknown
endif

GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
ifeq ($(GIT_BRANCH),)
    GIT_BRANCH=unknown
endif

# 清理分支名称中的空格和特殊字符
GIT_BRANCH_CLEAN=$(shell echo $(GIT_BRANCH) | sed 's/[[:space:]/_-]/_/g' | sed 's/__*/_/g')
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# 构建 ldflags（注意包路径）
LDFLAGS := -X 'github.com/sysint/internal/version.Version=$(VERSION)'
LDFLAGS += -X 'github.com/sysint/internal/version.GitCommit=$(GIT_COMMIT)'
LDFLAGS += -X 'github.com/sysint/internal/version.GitBranch=$(GIT_BRANCH_CLEAN)'
LDFLAGS += -X 'github.com/sysint/internal/version.BuildTime=$(BUILD_TIME)'

# 颜色输出
GREEN  := $(shell printf "\033[32m")
YELLOW := $(shell printf "\033[33m")
RED    := $(shell printf "\033[31m")
BLUE   := $(shell printf "\033[34m")
CYAN   := $(shell printf "\033[36m")
MAGENTA:= $(shell printf "\033[35m")
WHITE  := $(shell printf "\033[37m")
BOLD   := $(shell printf "\033[1m")
RESET  := $(shell printf "\033[0m")

# ============================================================================
# 帮助信息
# ============================================================================
help: ## 显示帮助信息
	@printf "$(GREEN)系统初始化检查工具 - Makefile命令列表$(RESET)\n"
	@printf "$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@printf "$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"

# ============================================================================
# 编译信息显示
# ============================================================================
info: ## 显示项目信息
	@printf "$(GREEN)项目信息:$(RESET)\n"
	@printf "$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"
	@printf "  项目根目录: $(PWD)\n"
	@printf "  Git Commit: $(GIT_COMMIT)\n"
	@printf "  Git Branch: $(GIT_BRANCH)\n"
	@printf "  Go版本:     $$(go version)\n"
	@printf "  依赖数量:   $$(go list -m all 2>/dev/null | wc -l | tr -d ' ') 个\n"
	@printf "$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"

version-info: ## 显示版本信息
	@printf "$(GREEN)版本信息:$(RESET)\n"
	@printf "  Version:    $(VERSION)\n"
	@printf "  GitCommit:  $(GIT_COMMIT)\n"
	@printf "  GitBranch:  $(GIT_BRANCH_CLEAN)\n"
	@printf "  BuildTime:  $(BUILD_TIME)\n"

# ============================================================================
# 编译相关
# ============================================================================
build: clean ## 编译项目（自动清理）
	@printf "$(GREEN)编译 $(BINARY_NAME)...$(RESET)\n"
	@mkdir -p $(BINARY_DIR)
	@printf "$(YELLOW)版本信息:$(RESET)\n"
	@printf "  Version:    $(VERSION)\n"
	@printf "  GitCommit:  $(GIT_COMMIT)\n"
	@printf "  GitBranch:  $(GIT_BRANCH_CLEAN)\n"
	@printf "  BuildTime:  $(BUILD_TIME)\n"
	@printf "\n"
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/sysint
	@if [ $$? -eq 0 ]; then \
		printf "\n$(GREEN)✓ 编译成功$(RESET)\n"; \
		printf "$(YELLOW)文件大小:$(RESET) $$(du -h $(BINARY_DIR)/$(BINARY_NAME) | cut -f1)\n"; \
		printf "\n$(YELLOW)版本验证:$(RESET)\n"; \
		./$(BINARY_DIR)/$(BINARY_NAME) version; \
	else \
		printf "\n$(RED)✗ 编译失败$(RESET)\n"; \
		exit 1; \
	fi

build-verbose: clean ## 详细编译（显示编译的包）
	@printf "$(GREEN)详细编译 $(BINARY_NAME)...$(RESET)\n"
	@mkdir -p $(BINARY_DIR)
	@printf "\n$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BLUE)编译的包列表:$(RESET)\n"
	@printf "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@go build -v -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/sysint
	@printf "\n$(GREEN)✓ 编译完成$(RESET)\n"
	@printf "$(YELLOW)文件大小:$(RESET) $$(du -h $(BINARY_DIR)/$(BINARY_NAME) | cut -f1)\n"

build-debug: clean ## 调试编译（显示完整调用日志）
	@printf "$(GREEN)调试模式编译 $(BINARY_NAME)...$(RESET)\n"
	@mkdir -p $(BINARY_DIR) $(BUILD_LOG_DIR)
	@printf "\n$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BLUE)详细调用日志:$(RESET)\n"
	@printf "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@echo ""
	@go build -v -x -work -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/sysint 2>&1 | tee $(BUILD_LOG)
	@printf "\n$(GREEN)✓ 编译完成$(RESET)\n"
	@printf "$(YELLOW)文件大小:$(RESET) $$(du -h $(BINARY_DIR)/$(BINARY_NAME) | cut -f1)\n"
	@printf "$(YELLOW)编译日志:$(RESET) $(BUILD_LOG)\n"

build-full: clean ## 完整编译（显示所有信息）
	@printf "$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"
	@printf "$(BOLD)$(WHITE)                    Go 完整编译信息$(RESET)\n"
	@printf "$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"
	@printf "\n$(YELLOW)【1. 系统环境】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "  操作系统:   $$(uname -a)\n"
	@printf "  主机名:     $$(hostname)\n"
	@printf "  当前用户:   $$(whoami)\n"
	@printf "\n$(YELLOW)【2. Go 环境】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "  Go版本:     $$(go version)\n"
	@printf "  GOOS:       $$(go env GOOS)\n"
	@printf "  GOARCH:     $$(go env GOARCH)\n"
	@printf "  CGO_ENABLED:$$(go env CGO_ENABLED)\n"
	@printf "  GOCACHE:    $$(go env GOCACHE)\n"
	@printf "\n$(YELLOW)【3. 版本信息】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "  Version:    $(VERSION)\n"
	@printf "  GitCommit:  $(GIT_COMMIT)\n"
	@printf "  GitBranch:  $(GIT_BRANCH_CLEAN)\n"
	@printf "  BuildTime:  $(BUILD_TIME)\n"
	@printf "\n$(YELLOW)【4. 编译命令】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "  go build -v -x -work -ldflags \"$(LDFLAGS)\" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/sysint\n"
	@printf "\n$(YELLOW)【5. 编译过程】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@mkdir -p $(BINARY_DIR) $(BUILD_LOG_DIR)
	@echo ""
	@time go build -v -x -work -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/sysint 2>&1 | tee $(BUILD_LOG)
	@printf "\n$(YELLOW)【6. 编译结果】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@if [ -f "$(BINARY_DIR)/$(BINARY_NAME)" ]; then \
		printf "$(GREEN)✓ 编译成功$(RESET)\n"; \
		printf "  文件路径: $(BINARY_DIR)/$(BINARY_NAME)\n"; \
		printf "  文件大小: $$(du -h $(BINARY_DIR)/$(BINARY_NAME) | cut -f1)\n"; \
		printf "  文件类型: $$(file $(BINARY_DIR)/$(BINARY_NAME) | cut -d: -f2)\n"; \
		printf "\n$(YELLOW)【7. 版本验证】$(RESET)\n"; \
		printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"; \
		./$(BINARY_DIR)/$(BINARY_NAME) version; \
	else \
		printf "$(RED)✗ 编译失败$(RESET)\n"; \
		exit 1; \
	fi
	@printf "\n$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"

build-no-clean: ## 增量编译（不清理）
	@printf "$(GREEN)增量编译 $(BINARY_NAME)...$(RESET)\n"
	@mkdir -p $(BINARY_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/sysint
	@printf "$(GREEN)✓ 编译完成: $(BINARY_DIR)/$(BINARY_NAME)$(RESET)\n"

# ============================================================================
# 清理相关
# ============================================================================
clean: ## 清理上次编译产物
	@printf "$(GREEN)清理上次编译产物...$(RESET)\n"
	@if [ -d "$(BINARY_DIR)" ]; then \
		rm -rf $(BINARY_DIR); \
		printf "  $(GREEN)✓$(RESET) 删除目录: $(BINARY_DIR)\n"; \
	fi
	@if [ -f "coverage.out" ]; then \
		rm -f coverage.out; \
		printf "  $(GREEN)✓$(RESET) 删除文件: coverage.out\n"; \
	fi
	@if [ -f "coverage.html" ]; then \
		rm -f coverage.html; \
		printf "  $(GREEN)✓$(RESET) 删除文件: coverage.html\n"; \
	fi
	@printf "$(GREEN)✓ 清理完成$(RESET)\n"

clean-logs: ## 清理编译日志
	@printf "$(GREEN)清理编译日志...$(RESET)\n"
	@if [ -d "$(BUILD_LOG_DIR)" ]; then \
		rm -rf $(BUILD_LOG_DIR); \
		printf "  $(GREEN)✓$(RESET) 删除目录: $(BUILD_LOG_DIR)\n"; \
	fi
	@printf "$(GREEN)✓ 日志清理完成$(RESET)\n"

clean-all: clean clean-logs ## 深度清理（包括缓存）
	@printf "$(GREEN)深度清理...$(RESET)\n"
	@go clean -cache
	@printf "$(GREEN)✓ Go缓存已清理$(RESET)\n"

# ============================================================================
# 运行相关
# ============================================================================
run: build ## 编译并运行
	@printf "$(GREEN)运行 $(BINARY_NAME)...$(RESET)\n\n"
	@./$(BINARY_DIR)/$(BINARY_NAME) version

run-debug: build ## 调试模式运行
	@printf "$(GREEN)调试模式运行...$(RESET)\n\n"
	@./$(BINARY_DIR)/$(BINARY_NAME) version --log-level debug

# ============================================================================
# 测试相关
# ============================================================================
test: ## 运行所有测试
	@printf "$(GREEN)运行所有测试...$(RESET)\n"
	@go test ./... -v -race -timeout 30s

test-unit: ## 运行单元测试
	@printf "$(GREEN)运行单元测试...$(RESET)\n"
	@go test ./internal/... -v -short -race

test-integration: ## 运行集成测试
	@printf "$(GREEN)运行集成测试...$(RESET)\n"
	@go test ./... -v -tags=integration

test-coverage: ## 运行测试并生成覆盖率报告
	@printf "$(GREEN)生成测试覆盖率报告...$(RESET)\n"
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@printf "$(GREEN)✓ 覆盖率报告已生成: coverage.html$(RESET)\n"

test-version: ## 测试version包
	@printf "$(GREEN)测试 version 包...$(RESET)\n"
	@go test ./internal/version -v

test-logger: ## 测试logger包
	@printf "$(GREEN)测试 logger 包...$(RESET)\n"
	@go test ./internal/logger -v

test-cmd: ## 测试cmd包
	@printf "$(GREEN)测试 cmd 包...$(RESET)\n"
	@go test ./internal/cmd -v

bench: ## 运行基准测试
	@printf "$(GREEN)运行基准测试...$(RESET)\n"
	@go test ./... -bench=. -benchmem -run=^$$

# ============================================================================
# 安装相关
# ============================================================================
install: build ## 安装到 GOPATH/bin
	@printf "$(GREEN)安装 $(BINARY_NAME)...$(RESET)\n"
	@go install -ldflags "$(LDFLAGS)" ./cmd/sysint
	@printf "$(GREEN)✓ 安装完成$(RESET)\n"

# ============================================================================
# 代码质量
# ============================================================================
fmt: ## 格式化代码
	@printf "$(GREEN)格式化代码...$(RESET)\n"
	@go fmt ./...
	@printf "$(GREEN)✓ 格式化完成$(RESET)\n"

lint: ## 运行代码检查
	@printf "$(GREEN)运行代码检查...$(RESET)\n"
	@command -v golangci-lint >/dev/null 2>&1 || { \
		printf "$(RED)golangci-lint 未安装，请运行:$(RESET)\n"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		exit 1; \
	}
	@golangci-lint run ./...
	@printf "$(GREEN)✓ 代码检查完成$(RESET)\n"

mod: ## 更新依赖
	@printf "$(GREEN)更新依赖...$(RESET)\n"
	@go mod tidy
	@go mod verify
	@printf "$(GREEN)✓ 依赖更新完成$(RESET)\n"

# ============================================================================
# 开发流程
# ============================================================================
dev: fmt lint test build ## 开发流程：格式化、检查、测试、编译
	@printf "$(GREEN)✓ 开发流程完成$(RESET)\n"

dev-full: fmt lint test-coverage build-full ## 完整开发流程
	@printf "$(GREEN)✓ 完整开发流程完成$(RESET)\n"