# ============================================================================
# 系统初始化检查工具 - Makefile（精简版，带详细清理日志）
# ============================================================================

# ============================================================================
# 变量定义
# ============================================================================
BINARY_NAME := kraken
BINARY_DIR  := bin

# ---- 版本信息 ----
VERSION     ?= dev
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
GIT_BRANCH  := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null | sed 's/[[:space:]/_-]/_/g' | sed 's/__*/_/g' || echo unknown)
BUILD_TIME  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# ---- 包路径（与 version.go 所在包一致） ----
PKG_VERSION := github.com/kraken-pedestal/internal/deploy/executor

LDFLAGS := -X '$(PKG_VERSION).version=$(VERSION)' \
           -X '$(PKG_VERSION).gitCommit=$(GIT_COMMIT)' \
           -X '$(PKG_VERSION).gitBranch=$(GIT_BRANCH)' \
           -X '$(PKG_VERSION).buildTime=$(BUILD_TIME)'

# ---- 本机构建默认目标 ----
GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BINARY_PATH := $(BINARY_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)

# 如果是本机构建，去掉后缀
ifeq ($(GOOS),$(shell go env GOOS))
ifeq ($(GOARCH),$(shell go env GOARCH))
    BINARY_PATH := $(BINARY_DIR)/$(BINARY_NAME)
endif
endif

# ---- 颜色输出（终端支持） ----
COLOR_RESET  := \033[0m
COLOR_GREEN  := \033[32m
COLOR_YELLOW := \033[33m
COLOR_RED    := \033[31m
COLOR_BLUE   := \033[34m

# ============================================================================
# 构建目标
# ============================================================================
build-debug: ## 调试构建（显示完整调用日志）
	@printf "$(COLOR_GREEN)调试构建 $(BINARY_NAME)...$(COLOR_RESET)\n"
	@mkdir -p $(BINARY_DIR)
	@if [ -f "$(BINARY_PATH)" ]; then \
		printf "$(COLOR_YELLOW)删除旧的二进制文件: $(BINARY_PATH)$(COLOR_RESET)\n"; \
		rm -f $(BINARY_PATH); \
	fi
	@printf "  版本: $(VERSION) | 平台: $(GOOS)/$(GOARCH)\n"
	@printf "\n$(COLOR_YELLOW)详细调用日志:$(COLOR_RESET)\n"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -x -work -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) .
	@if [ $$? -eq 0 ]; then \
		printf "\n$(COLOR_GREEN)✓ 构建成功$(COLOR_RESET)\n"; \
		printf "  文件: $(BINARY_PATH) ($$(du -h $(BINARY_PATH) | cut -f1))\n"; \
	else \
		printf "\n$(COLOR_RED)✗ 构建失败$(COLOR_RESET)\n"; \
		exit 1; \
	fi

build-full: ## 完整构建（显示系统、Go 环境及详细编译信息）
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)                     Go 完整构建信息$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "\n$(COLOR_YELLOW)【系统环境】$(COLOR_RESET)\n"
	@printf "  操作系统:   $$(uname -a)\n"
	@printf "  主机名:     $$(hostname)\n"
	@printf "  当前用户:   $$(whoami)\n"
	@printf "\n$(COLOR_YELLOW)【Go 环境】$(COLOR_RESET)\n"
	@printf "  Go版本:     $$(go version)\n"
	@printf "  GOOS:       $$(go env GOOS)\n"
	@printf "  GOARCH:     $$(go env GOARCH)\n"
	@printf "  CGO_ENABLED:$$(go env CGO_ENABLED)\n"
	@printf "\n$(COLOR_YELLOW)【版本信息】$(COLOR_RESET)\n"
	@printf "  Version:    $(VERSION)\n"
	@printf "  GitCommit:  $(GIT_COMMIT)\n"
	@printf "  GitBranch:  $(GIT_BRANCH)\n"
	@printf "  BuildTime:  $(BUILD_TIME)\n"
	@printf "  目标平台:   $(GOOS)/$(GOARCH)\n"
	@printf "\n$(COLOR_YELLOW)【编译命令】$(COLOR_RESET)\n"
	@printf "  GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -x -work -ldflags \"$(LDFLAGS)\" -o $(BINARY_PATH) .\n"
	@printf "\n$(COLOR_YELLOW)【编译过程】$(COLOR_RESET)\n"
	@mkdir -p $(BINARY_DIR)
	@if [ -f "$(BINARY_PATH)" ]; then \
		printf "$(COLOR_YELLOW)删除旧的二进制文件: $(BINARY_PATH)$(COLOR_RESET)\n"; \
		rm -f $(BINARY_PATH); \
	fi
	@time GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -x -work -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) .
	@if [ $$? -eq 0 ]; then \
		printf "\n$(COLOR_YELLOW)【编译结果】$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)✓ 构建成功$(COLOR_RESET)\n"; \
		printf "  文件路径: $(BINARY_PATH)\n"; \
		printf "  文件大小: $$(du -h $(BINARY_PATH) | cut -f1)\n"; \
		printf "  文件类型: $$(file $(BINARY_PATH) | cut -d: -f2)\n"; \
	else \
		printf "\n$(COLOR_RED)✗ 构建失败$(COLOR_RESET)\n"; \
		exit 1; \
	fi
	@printf "\n$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"

build-linux-arm: ## 构建 Linux arm64（构建前自动删除旧文件）
	@$(MAKE) build-debug GOOS=linux GOARCH=arm64

clean-all: ## 深度清理（包括 Go 缓存）
	@printf "$(COLOR_GREEN)深度清理...$(COLOR_RESET)\n"
	@if [ -d "$(BINARY_DIR)" ]; then \
		printf "$(COLOR_YELLOW)删除目录: $(BINARY_DIR)$(COLOR_RESET)\n"; \
		rm -rf $(BINARY_DIR); \
	else \
		printf "$(COLOR_YELLOW)目录 $(BINARY_DIR) 不存在，跳过$(COLOR_RESET)\n"; \
	fi
	@if [ -f "coverage.out" ]; then \
		printf "$(COLOR_YELLOW)删除文件: coverage.out$(COLOR_RESET)\n"; \
		rm -f coverage.out; \
	fi
	@if [ -f "coverage.html" ]; then \
		printf "$(COLOR_YELLOW)删除文件: coverage.html$(COLOR_RESET)\n"; \
		rm -f coverage.html; \
	fi
	@printf "$(COLOR_YELLOW)清理 Go 缓存...$(COLOR_RESET)\n"
	@go clean -cache -testcache
	@printf "$(COLOR_GREEN)✓ 清理完成（含缓存）$(COLOR_RESET)\n"

# ============================================================================
# 默认目标（显示可用目标）
# ============================================================================
.DEFAULT_GOAL := help
help:
	@printf "$(COLOR_GREEN)可用目标:$(COLOR_RESET)\n"
	@printf "  build-debug      调试构建（显示完整调用日志）\n"
	@printf "  build-full       完整构建（显示系统、Go 环境及详细编译信息）\n"
	@printf "  build-linux-arm  构建 Linux arm64（自动删除旧文件）\n"
	@printf "  clean-all        深度清理（包括 Go 缓存）\n"