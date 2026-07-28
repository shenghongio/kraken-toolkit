# ============================================================================
# 系统初始化检查工具 - Makefile
# ============================================================================

.PHONY: help build build-verbose build-debug build-full build-no-clean clean clean-all run run-debug install fmt lint mod dev info version-info
.PHONY: build-linux build-linux-arm   # 保留 Linux 快捷目标，可按需删除

# ============================================================================
# 变量定义
# ============================================================================
BINARY_NAME=kraken
BINARY_DIR=bin

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
LDFLAGS := -X 'github.com/kraken-pedestal/internal/executor.Version=$(VERSION)'
LDFLAGS += -X 'github.com/kraken-pedestal/internal/executor.GitCommit=$(GIT_COMMIT)'
LDFLAGS += -X 'github.com/kraken-pedestal/internal/executor.GitBranch=$(GIT_BRANCH_CLEAN)'
LDFLAGS += -X 'github.com/kraken-pedestal/internal/executor.BuildTime=$(BUILD_TIME)'

# 交叉编译设置
GOOS ?=
GOARCH ?=
ifeq ($(GOOS),)
    OUTPUT_SUFFIX :=
else
    ifeq ($(GOARCH),)
        GOARCH := amd64
    endif
    OUTPUT_SUFFIX := -$(GOOS)-$(GOARCH)
endif
BINARY_PATH := $(BINARY_DIR)/$(BINARY_NAME)$(OUTPUT_SUFFIX)

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
	@printf "$(GREEN)System Initialization Check Tool - Makefile Command List$(RESET)\n"
	@printf "$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@printf "\n$(YELLOW)交叉编译示例:$(RESET)\n"
	@printf "  make build GOOS=linux GOARCH=amd64     # 编译 Linux amd64\n"
	@printf "  make build-linux                       # 快捷编译 Linux amd64\n"
	@printf "  make build-linux-arm                   # 快捷编译 Linux arm64\n"
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
	@if [ -n "$(GOOS)" ]; then \
		printf "  目标平台:  $(GOOS)/$(GOARCH)\n"; \
	fi
	@printf "\n"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/
	@if [ $$? -eq 0 ]; then \
		printf "\n$(GREEN)✓ 编译成功$(RESET)\n"; \
		printf "$(YELLOW)文件大小:$(RESET) $$(du -h $(BINARY_PATH) | cut -f1)\n"; \
		if [ -z "$(GOOS)" ]; then \
			printf "\n$(YELLOW)版本验证:$(RESET)\n"; \
			./$(BINARY_PATH) version; \
		else \
			printf "\n$(YELLOW)交叉编译产物，跳过本地运行验证$(RESET)\n"; \
		fi \
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
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/
	@printf "\n$(GREEN)✓ 编译完成$(RESET)\n"
	@printf "$(YELLOW)文件大小:$(RESET) $$(du -h $(BINARY_PATH) | cut -f1)\n"

build-debug: clean ## 调试编译（显示完整调用日志，不保存文件）
	@printf "$(GREEN)调试模式编译 $(BINARY_NAME)...$(RESET)\n"
	@mkdir -p $(BINARY_DIR)
	@printf "\n$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "$(BLUE)详细调用日志:$(RESET)\n"
	@printf "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@echo ""
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -x -work -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/
	@printf "\n$(GREEN)✓ 编译完成$(RESET)\n"
	@printf "$(YELLOW)文件大小:$(RESET) $$(du -h $(BINARY_PATH) | cut -f1)\n"

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
	@if [ -n "$(GOOS)" ]; then \
		printf "  目标平台:  $(GOOS)/$(GOARCH)\n"; \
	fi
	@printf "\n$(YELLOW)【4. 编译命令】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@printf "  GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -x -work -ldflags \"$(LDFLAGS)\" -o $(BINARY_PATH) ./cmd/\n"
	@printf "\n$(YELLOW)【5. 编译过程】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@echo ""
	@time GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -x -work -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/
	@printf "\n$(YELLOW)【6. 编译结果】$(RESET)\n"
	@printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"
	@if [ -f "$(BINARY_PATH)" ]; then \
		printf "$(GREEN)✓ 编译成功$(RESET)\n"; \
		printf "  文件路径: $(BINARY_PATH)\n"; \
		printf "  文件大小: $$(du -h $(BINARY_PATH) | cut -f1)\n"; \
		printf "  文件类型: $$(file $(BINARY_PATH) | cut -d: -f2)\n"; \
		if [ -z "$(GOOS)" ]; then \
			printf "\n$(YELLOW)【7. 版本验证】$(RESET)\n"; \
			printf "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)\n"; \
			./$(BINARY_PATH) version; \
		else \
			printf "\n$(YELLOW)交叉编译产物，跳过本地运行验证$(RESET)\n"; \
		fi \
	else \
		printf "$(RED)✗ 编译失败$(RESET)\n"; \
		exit 1; \
	fi
	@printf "\n$(BLUE)═══════════════════════════════════════════════════════════════$(RESET)\n"

build-no-clean: ## 增量编译（不清理）
	@printf "$(GREEN)增量编译 $(BINARY_NAME)...$(RESET)\n"
	@mkdir -p $(BINARY_DIR)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/
	@printf "$(GREEN)✓ 编译完成: $(BINARY_PATH)$(RESET)\n"

# ============================================================================
# 平台快捷编译目标（仅保留 Linux 示例）
# ============================================================================
build-linux: ## 编译 Linux (amd64) 版本
	$(MAKE) build GOOS=linux GOARCH=amd64

build-linux-arm: ## 编译 Linux (arm64) 版本
	$(MAKE) build GOOS=linux GOARCH=arm64

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
	@printf "$(GREEN)✓ 清理完成$(RESET)\n"

clean-all: clean ## 深度清理（包括缓存）
	@printf "$(GREEN)深度清理...$(RESET)\n"
	@go clean -cache
	@printf "$(GREEN)✓ Go缓存已清理$(RESET)\n"

# ============================================================================
# 运行相关
# ============================================================================
run: build ## 编译并运行（仅当编译本机版本时有效）
	@printf "$(GREEN)运行 $(BINARY_NAME)...$(RESET)\n\n"
	@./$(BINARY_PATH) version

run-debug: build ## 调试模式运行
	@printf "$(GREEN)调试模式运行...$(RESET)\n\n"
	@./$(BINARY_PATH) version --log-level debug

# ============================================================================
# 安装相关
# ============================================================================
install: build ## 安装到 GOPATH/bin（仅本机）
	@printf "$(GREEN)安装 $(BINARY_NAME)...$(RESET)\n"
	@go install -ldflags "$(LDFLAGS)" ./cmd/
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
dev: fmt lint build ## 开发流程：格式化、检查、编译
	@printf "$(GREEN)✓ 开发流程完成$(RESET)\n"

dev-full: fmt lint build-full ## 完整开发流程（含详细编译信息）
	@printf "$(GREEN)✓ 完整开发流程完成$(RESET)\n"