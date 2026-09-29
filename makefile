# ============================================================================
# Kraken - Makefile
#
# 项目：Kraken 系统初始化 / 部署工具
#
# 日常开发：
#
#   make fmt
#   make check
#   make test
#   make build
#
# 构建：
#
#   make build
#   make build-arm64
#   make build-amd64
#   make build-full
#
# 一键流水线：
#
#   make pipeline
#
# 发布：
#
#   make release VERSION=v1.0.0
#
# 上传：
#
#   make 139_scp ARGS="bin/kraken-linux-arm64"
#   make 138_scp ARGS="bin/kraken-linux-amd64"
#
# Git 快捷操作：
#
#   make git-status
#   make git-log
#   make git-diff
#   make git-pull
#   make git-push
#   make git-add
#   make git-commit MSG="xxx"
#
# 清理：
#
#   make clean
#   make clean-all
#
# 帮助：
#
#   make help
#
# ============================================================================


# ============================================================================
# 基础配置
# ============================================================================

BINARY_NAME := kraken
BINARY_DIR  := bin


# ============================================================================
# 版本信息
# ============================================================================

# 默认版本
#
# 示例：
#
#   make build VERSION=v1.0.0
#
VERSION ?= dev


# Git Commit
GIT_COMMIT := $(shell \
	git rev-parse --short HEAD 2>/dev/null || echo unknown \
)


# Git Branch
GIT_BRANCH := $(shell \
	git rev-parse --abbrev-ref HEAD 2>/dev/null \
	| sed 's/[[:space:]/_-]/_/g' \
	| sed 's/__*/_/g' \
	|| echo unknown \
)


# 构建时间
#
# 使用 UTC，避免不同机器时区导致构建时间不一致。
#
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)


# ============================================================================
# Go 配置
# ============================================================================

GO    := go
GOFMT := gofmt


# 当前平台
#
# 默认使用当前机器的 GOOS / GOARCH。
#
# 可以通过命令行覆盖：
#
#   make build GOOS=linux GOARCH=arm64
#
GOOS   ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)


# 构建产物按平台区分命名：
#
#   bin/kraken-linux-arm64
#   bin/kraken-linux-amd64
#   bin/kraken-darwin-arm64
#
# 注意：BINARY_PATH 里用到了 GOOS / GOARCH，
# 所以必须放在这两个变量之后定义。
#
BINARY_PATH := $(BINARY_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)


# ============================================================================
# Version Package
# ============================================================================

# version.go 所在 package
#
# version.go 中需要存在类似：
#
#   var (
#       version   = "dev"
#       gitCommit = "unknown"
#       gitBranch = "unknown"
#       buildTime = "unknown"
#   )
#
#PKG_VERSION := github.com/kraken-pedestal/internal/deploy/executor
PKG_VERSION := github.com/kraken-pedestal/pkg/cli


# ============================================================================
# Linker Flags
# ============================================================================

# 将版本信息写入二进制文件。
#
# 使用：
#
#   go build -ldflags "$(LDFLAGS)"
#
LDFLAGS := \
	-X '$(PKG_VERSION).version=$(VERSION)' \
	-X '$(PKG_VERSION).gitCommit=$(GIT_COMMIT)' \
	-X '$(PKG_VERSION).gitBranch=$(GIT_BRANCH)' \
	-X '$(PKG_VERSION).buildTime=$(BUILD_TIME)'


# ============================================================================
# SCP 配置
# ============================================================================

# 10.32.9.139
SCP_HOST_139 := sketc@10.32.9.139
SCP_TARGET_139 := $(SCP_HOST_139):/home/sketc/


# 10.32.9.138
SCP_HOST_138 := sketc@10.32.9.138
SCP_TARGET_138 := $(SCP_HOST_138):/home/sketc/


# ============================================================================
# 颜色
# ============================================================================

COLOR_RESET  := \033[0m
COLOR_GREEN  := \033[32m
COLOR_YELLOW := \033[33m
COLOR_RED    := \033[31m
COLOR_BLUE   := \033[34m
COLOR_CYAN   := \033[36m


# ============================================================================
# 默认目标
# ============================================================================

.DEFAULT_GOAL := help


# ============================================================================
# Phony
# ============================================================================

.PHONY: \
	fmt \
	fmt-check \
	vet \
	check \
	test \
	coverage \
	build \
	build-debug \
	build-arm64 \
	build-amd64 \
	build-full \
	release \
	pipeline \
	139_scp \
	138_scp \
	clean \
	clean-all \
	git-status \
	git-log \
	git-diff \
	git-pull \
	git-push \
	git-add \
	git-commit \
	help


# ============================================================================
# 格式化
# ============================================================================

fmt: ## 格式化 Go 源代码
	@printf "$(COLOR_GREEN)格式化 Go 源代码...$(COLOR_RESET)\n"
	@$(GOFMT) -w .
	@printf "$(COLOR_GREEN)✓ 格式化完成$(COLOR_RESET)\n"


# ============================================================================
# 格式检查
# ============================================================================

fmt-check: ## 检查 Go 源代码格式
	@printf "$(COLOR_GREEN)检查 Go 源代码格式...$(COLOR_RESET)\n"
	@files="$$(find . -type f -name '*.go' \
		-not -path './vendor/*' \
		-not -path './.git/*' \
		-exec $(GOFMT) -l {} \;)"; \
	if [ -n "$$files" ]; then \
		printf "$(COLOR_RED)✗ 以下文件未通过 gofmt 检查:$(COLOR_RESET)\n"; \
		printf '%s\n' "$$files"; \
		printf "\n$(COLOR_YELLOW)请执行：make fmt$(COLOR_RESET)\n"; \
		exit 1; \
	fi
	@printf "$(COLOR_GREEN)✓ 格式检查通过$(COLOR_RESET)\n"


# ============================================================================
# Go Vet
# ============================================================================

vet: ## 执行 go vet
	@printf "$(COLOR_GREEN)执行 go vet...$(COLOR_RESET)\n"
	@sleep 10
	@if $(GO) vet ./...; then \
		printf "$(COLOR_GREEN)✓ vet 检查通过$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_RED)✗ vet 检查失败$(COLOR_RESET)\n"; \
		exit 1; \
	fi


# ============================================================================
# 综合检查
# ============================================================================

check: fmt-check vet ## 执行格式检查和静态检查
	@sleep 10
	@printf "\n$(COLOR_GREEN)✓ 代码检查全部通过$(COLOR_RESET)\n"


# ============================================================================
# 单元测试
# ============================================================================

test: ## 执行单元测试
	@printf "$(COLOR_GREEN)执行单元测试...$(COLOR_RESET)\n"
	@printf "$(COLOR_CYAN)  命令: $(GO) test -v -count=1 ./...$(COLOR_RESET)\n"
	@printf "$(COLOR_CYAN)  参数: -v (详细输出) | -count=1 (禁用测试缓存)$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)───────────────────────────────────────────────────────────────$(COLOR_RESET)\n"
	@set -o pipefail; \
	if $(GO) test -v -count=1 ./... 2>&1 | tee /tmp/kraken_test_output.log; then \
		printf "$(COLOR_BLUE)───────────────────────────────────────────────────────────────$(COLOR_RESET)\n"; \
		PASS_COUNT=$$(grep -c -- '--- PASS' /tmp/kraken_test_output.log); \
		FAIL_COUNT=$$(grep -c -- '--- FAIL' /tmp/kraken_test_output.log); \
		SKIP_COUNT=$$(grep -c -- '--- SKIP' /tmp/kraken_test_output.log); \
		PKG_COUNT=$$(grep -E '^(ok|FAIL)\s' /tmp/kraken_test_output.log | wc -l | tr -d ' '); \
		printf "$(COLOR_YELLOW)  测试包数: $$PKG_COUNT$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)  通过 (PASS): $$PASS_COUNT$(COLOR_RESET)\n"; \
		if [ "$$FAIL_COUNT" -gt 0 ]; then \
			printf "$(COLOR_RED)  失败 (FAIL): $$FAIL_COUNT$(COLOR_RESET)\n"; \
		fi; \
		if [ "$$SKIP_COUNT" -gt 0 ]; then \
			printf "$(COLOR_YELLOW)  跳过 (SKIP): $$SKIP_COUNT$(COLOR_RESET)\n"; \
		fi; \
		printf "$(COLOR_GREEN)✓ 单元测试通过$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_BLUE)───────────────────────────────────────────────────────────────$(COLOR_RESET)\n"; \
		FAIL_COUNT=$$(grep -c -- '--- FAIL' /tmp/kraken_test_output.log); \
		printf "$(COLOR_RED)  失败 (FAIL): $$FAIL_COUNT$(COLOR_RESET)\n"; \
		printf "$(COLOR_RED)  失败的测试详情：$(COLOR_RESET)\n"; \
		grep -- '--- FAIL' /tmp/kraken_test_output.log | sed 's/^/    /'; \
		printf "$(COLOR_RED)✗ 单元测试失败$(COLOR_RESET)\n"; \
		printf "$(COLOR_YELLOW)  完整日志已保存至: /tmp/kraken_test_output.log$(COLOR_RESET)\n"; \
		exit 1; \
	fi
	@rm -f /tmp/kraken_test_output.log


# ============================================================================
# 测试覆盖率
# ============================================================================

coverage: ## 执行测试并生成覆盖率
	@printf "$(COLOR_GREEN)执行测试并生成覆盖率...$(COLOR_RESET)\n"
	@if $(GO) test ./... -coverprofile=coverage.out; then \
		printf "$(COLOR_GREEN)✓ 测试通过$(COLOR_RESET)\n"; \
		printf "\n$(COLOR_YELLOW)覆盖率统计：$(COLOR_RESET)\n"; \
		$(GO) tool cover -func=coverage.out; \
	else \
		printf "$(COLOR_RED)✗ 测试失败$(COLOR_RESET)\n"; \
		exit 1; \
	fi


# ============================================================================
# 当前平台构建
# ============================================================================
#
# make build
#
# 使用当前机器的 GOOS / GOARCH。
#
# 最终输出：
#
#   bin/kraken-<goos>-<goarch>
#
# ============================================================================

build: ## 构建当前平台二进制文件
	@$(MAKE) build-debug \
		GOOS=$(GOOS) \
		GOARCH=$(GOARCH) \
		VERSION="$(VERSION)"


# ============================================================================
# Linux ARM64
# ============================================================================
#
# make build-arm64
#
# 构建：
#
#   linux/arm64
#
# 最终输出：
#
#   bin/kraken-linux-arm64
#
# ============================================================================

build-arm64: ## 构建 Linux ARM64（完整编译链）
	@sleep 10
	@$(MAKE) build-debug \
		GOOS=linux \
		GOARCH=arm64 \
		VERSION="$(VERSION)"


# ============================================================================
# Linux AMD64
# ============================================================================
#
# make build-amd64
#
# 构建：
#
#   linux/amd64
#
# 最终输出：
#
#   bin/kraken-linux-amd64
#
# ============================================================================

build-amd64: ## 构建 Linux AMD64（完整编译链）
	@$(MAKE) build-debug \
		GOOS=linux \
		GOARCH=amd64 \
		VERSION="$(VERSION)"


# ============================================================================
# 底层构建
# ============================================================================
#
# 所有构建最终都会进入这里。
#
# 编译参数：
#
#   -v
#       显示正在编译的 package。
#
#   -x
#       显示 Go 实际执行的底层命令。
#
#   -work
#       保留 Go 编译临时工作目录。
#
# ============================================================================

build-debug: ## 执行完整 Go 编译链
	@printf "\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)                    构建 $(BINARY_NAME)$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"

	# ------------------------------------------------------------------------
	# 构建信息
	# ------------------------------------------------------------------------

	@printf "\n$(COLOR_YELLOW)【构建信息】$(COLOR_RESET)\n"
	@printf "  Version:     $(VERSION)\n"
	@printf "  GitCommit:   $(GIT_COMMIT)\n"
	@printf "  GitBranch:   $(GIT_BRANCH)\n"
	@printf "  BuildTime:   $(BUILD_TIME)\n"
	@printf "  Platform:    $(GOOS)/$(GOARCH)\n"
	@printf "  Output:      $(BINARY_PATH)\n"

	# ------------------------------------------------------------------------
	# Go 环境
	# ------------------------------------------------------------------------

	@printf "\n$(COLOR_YELLOW)【Go 环境】$(COLOR_RESET)\n"
	@printf "  Go版本:      $$($(GO) version)\n"
	@printf "  GOROOT:      $$($(GO) env GOROOT)\n"
	@printf "  GOPATH:      $$($(GO) env GOPATH)\n"
	@printf "  GOOS:        $(GOOS)\n"
	@printf "  GOARCH:      $(GOARCH)\n"
	@printf "  CGO_ENABLED: $$($(GO) env CGO_ENABLED)\n"

	# ------------------------------------------------------------------------
	# 创建输出目录
	# ------------------------------------------------------------------------

	@mkdir -p "$(BINARY_DIR)"

	# ------------------------------------------------------------------------
	# 删除旧二进制
	# ------------------------------------------------------------------------

	@if [ -f "$(BINARY_PATH)" ]; then \
		printf "\n$(COLOR_YELLOW)删除旧的二进制文件: $(BINARY_PATH)$(COLOR_RESET)\n"; \
		rm -f "$(BINARY_PATH)"; \
	fi

	# ------------------------------------------------------------------------
	# 完整编译链
	# ------------------------------------------------------------------------

	@printf "\n$(COLOR_YELLOW)【完整编译链】$(COLOR_RESET)\n"

	# 注意：
	#
	# 这里必须使用：
	#
	#   printf '%s\n'
	#
	# 而不能：
	#
	#   printf "-ldflags ..."
	#
	# 因为 macOS 的 printf 会把 "-ldflags" 的 "-l"
	# 识别成 printf 自身的 option。
	#
	@printf '%s\n' "$(COLOR_CYAN)GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -v -x -work -ldflags \"$(LDFLAGS)\" -o $(BINARY_PATH) .$(COLOR_RESET)"

	# ------------------------------------------------------------------------
	# 执行 Go Build
	# ------------------------------------------------------------------------

	@if GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build \
		-v \
		-x \
		-work \
		-ldflags "$(LDFLAGS)" \
		-o "$(BINARY_PATH)" .; then \
		printf "\n"; \
		printf "$(COLOR_GREEN)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)✓ 构建成功$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)  Version:  $(VERSION)$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)  Platform: $(GOOS)/$(GOARCH)$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)  File:     $(BINARY_PATH)$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)  Size:     $$(du -h "$(BINARY_PATH)" | cut -f1)$(COLOR_RESET)\n"; \
		printf "$(COLOR_GREEN)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"; \
	else \
		printf "\n"; \
		printf "$(COLOR_RED)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"; \
		printf "$(COLOR_RED)✗ 构建失败$(COLOR_RESET)\n"; \
		printf "$(COLOR_RED)  Version:  $(VERSION)$(COLOR_RESET)\n"; \
		printf "$(COLOR_RED)  Platform: $(GOOS)/$(GOARCH)$(COLOR_RESET)\n"; \
		printf "$(COLOR_RED)  Output:   $(BINARY_PATH)$(COLOR_RESET)\n"; \
		printf "$(COLOR_RED)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"; \
		exit 1; \
	fi


# ============================================================================
# 完整构建
# ============================================================================
#
# make build-full
#
# build-full 默认构建当前机器平台。
#
# 与 make build 的区别：
#
#   build:
#       直接执行构建。
#
#   build-full:
#       构建之前额外显示系统环境和完整 Go 环境。
#
# 最终仍然进入 build-debug。
#
# ============================================================================

build-full: ## 显示完整系统环境并执行构建
	@printf "\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)                     Go 完整构建信息$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"

	# ------------------------------------------------------------------------
	# 系统环境
	# ------------------------------------------------------------------------

	@printf "\n$(COLOR_YELLOW)【系统环境】$(COLOR_RESET)\n"
	@printf "  操作系统:    $$(uname -a)\n"
	@printf "  主机名:      $$(hostname)\n"
	@printf "  当前用户:    $$(whoami)\n"

	# ------------------------------------------------------------------------
	# Go 环境
	# ------------------------------------------------------------------------

	@printf "\n$(COLOR_YELLOW)【Go 环境】$(COLOR_RESET)\n"
	@printf "  Go版本:      $$($(GO) version)\n"
	@printf "  GOOS:        $(GOOS)\n"
	@printf "  GOARCH:      $(GOARCH)\n"
	@printf "  CGO_ENABLED: $$($(GO) env CGO_ENABLED)\n"
	@printf "  GOPATH:      $$($(GO) env GOPATH)\n"
	@printf "  GOROOT:      $$($(GO) env GOROOT)\n"

	# ------------------------------------------------------------------------
	# 版本信息
	# ------------------------------------------------------------------------

	@printf "\n$(COLOR_YELLOW)【版本信息】$(COLOR_RESET)\n"
	@printf "  Version:     $(VERSION)\n"
	@printf "  GitCommit:   $(GIT_COMMIT)\n"
	@printf "  GitBranch:   $(GIT_BRANCH)\n"
	@printf "  BuildTime:   $(BUILD_TIME)\n"
	@printf "  Platform:    $(GOOS)/$(GOARCH)\n"
	@printf "  Output:      $(BINARY_PATH)\n"

	# ------------------------------------------------------------------------
	# 进入统一构建流程
	# ------------------------------------------------------------------------

	@$(MAKE) build-debug \
		GOOS=$(GOOS) \
		GOARCH=$(GOARCH) \
		VERSION="$(VERSION)"


# ============================================================================
# Release
# ============================================================================
#
# Release 执行：
#
#   1. fmt-check
#   2. go vet
#   3. 单元测试
#   4. 构建 arm64 + amd64
#
# 任意一步失败，Release 都失败。
#
# ============================================================================

release: ## 执行完整检查、测试并构建（arm64 + amd64）
	@printf "\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)                     Release $(BINARY_NAME)$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"

	@printf "\n$(COLOR_YELLOW)[1/4] 代码检查$(COLOR_RESET)\n"
	@$(MAKE) check

	@printf "\n$(COLOR_YELLOW)[2/4] 单元测试$(COLOR_RESET)\n"
	@$(MAKE) test

	@printf "\n$(COLOR_YELLOW)[3/4] 构建（linux/arm64 + linux/amd64）$(COLOR_RESET)\n"
	@$(MAKE) build-arm64 VERSION="$(VERSION)"
	@$(MAKE) build-amd64 VERSION="$(VERSION)"

	@printf "\n$(COLOR_YELLOW)[4/4] Release 完成$(COLOR_RESET)\n"
	@printf "$(COLOR_GREEN)✓ Release 成功$(COLOR_RESET)\n"
	@printf "  Version: $(VERSION)\n"
	@printf "  File:    $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64\n"
	@printf "  File:    $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64\n"

	@printf "\n$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"


# ============================================================================
# 一键流水线
# ============================================================================
#
# 依次执行：
#
#   1. make clean-all
#   2. make fmt
#   3. make vet
#   4. make build-arm64
#   5. make build-amd64
#
# 任意一步失败，pipeline 立即终止。
#
# 用法：
#
#   make pipeline
#   make pipeline VERSION=v1.0.0
#
# ============================================================================

pipeline: ## 一键执行 clean-all → fmt → vet → build-arm64 → build-amd64
	@printf "\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)                One-Key Pipeline（一键流水线）$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"

	@printf "\n$(COLOR_YELLOW)[1/5] make clean-all$(COLOR_RESET)\n"
	@$(MAKE) clean-all
	@sleep 1

	@printf "\n$(COLOR_YELLOW)[2/5] make fmt$(COLOR_RESET)\n"
	@$(MAKE) fmt
	@sleep 1

	@printf "\n$(COLOR_YELLOW)[3/5] make vet$(COLOR_RESET)\n"
	@$(MAKE) vet
	@sleep 1

	@printf "\n$(COLOR_YELLOW)[4/5] make build-arm64$(COLOR_RESET)\n"
	@$(MAKE) build-arm64 VERSION="$(VERSION)"
	@sleep 1

	@printf "\n$(COLOR_YELLOW)[5/5] make build-amd64$(COLOR_RESET)\n"
	@$(MAKE) build-amd64 VERSION="$(VERSION)"

	@printf "\n$(COLOR_GREEN)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "$(COLOR_GREEN)✓ One-Key Pipeline 完成$(COLOR_RESET)\n"
	@printf "$(COLOR_GREEN)  - linux/arm64: $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64$(COLOR_RESET)\n"
	@printf "$(COLOR_GREEN)  - linux/amd64: $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64$(COLOR_RESET)\n"
	@printf "$(COLOR_GREEN)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"


# ============================================================================
# Git 快捷操作
# ============================================================================

git-status: ## 查看 Git 状态
	@printf "$(COLOR_GREEN)git status$(COLOR_RESET)\n"
	@git status

git-log: ## 查看最近 20 条提交（图形化）
	@printf "$(COLOR_GREEN)git log --oneline --graph --decorate -20$(COLOR_RESET)\n"
	@git log --oneline --graph --decorate -20

git-diff: ## 查看未暂存的改动
	@printf "$(COLOR_GREEN)git diff$(COLOR_RESET)\n"
	@git diff

git-pull: ## 从远端拉取代码
	@printf "$(COLOR_GREEN)git pull$(COLOR_RESET)\n"
	@git pull

git-push: ## 推送到远端
	@printf "$(COLOR_GREEN)git push$(COLOR_RESET)\n"
	@git push

git-add: ## 将所有改动添加到暂存区
	@printf "$(COLOR_GREEN)git add -A$(COLOR_RESET)\n"
	@git add -A

git-commit: ## 提交改动（需要 MSG 参数）
	@if [ -z "$(strip $(MSG))" ]; then \
		printf "$(COLOR_RED)✗ 未指定提交信息$(COLOR_RESET)\n"; \
		printf "\n$(COLOR_YELLOW)用法：$(COLOR_RESET)\n"; \
		printf "  make git-commit MSG=\"提交信息\"\n"; \
		printf "\n$(COLOR_YELLOW)示例：$(COLOR_RESET)\n"; \
		printf "  make git-commit MSG=\"fix: 修复 vet 缩进\"\n"; \
		exit 1; \
	fi
	@printf "$(COLOR_GREEN)git add -A && git commit -m \"$(MSG)\"$(COLOR_RESET)\n"
	@git add -A
	@git commit -m "$(MSG)"


# ============================================================================
# SCP - 10.32.9.139
# ============================================================================
#
# 上传：
#
#   make 139_scp ARGS="bin/kraken-linux-arm64"
#
# 多文件：
#
#   make 139_scp ARGS="bin/kraken-linux-arm64 config.yaml"
#
# ============================================================================

139_scp: ## 上传文件到 10.32.9.139
	@if [ -z "$(strip $(ARGS))" ]; then \
		printf "$(COLOR_RED)✗ 未指定上传文件$(COLOR_RESET)\n"; \
		printf "\n$(COLOR_YELLOW)用法：$(COLOR_RESET)\n"; \
		printf "  make 139_scp ARGS=\"文件路径\"\n"; \
		printf "\n$(COLOR_YELLOW)示例：$(COLOR_RESET)\n"; \
		printf "  make 139_scp ARGS=\"bin/kraken-linux-arm64\"\n"; \
		printf "  make 139_scp ARGS=\"bin/kraken-linux-amd64\"\n"; \
		exit 1; \
	fi

	@printf "\n$(COLOR_GREEN)SCP 上传$(COLOR_RESET)\n"
	@printf "  文件: $(ARGS)\n"
	@printf "  目标: $(SCP_TARGET_139)\n"

	@if scp $(ARGS) "$(SCP_TARGET_139)"; then \
		printf "$(COLOR_GREEN)✓ SCP 上传成功$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_RED)✗ SCP 上传失败$(COLOR_RESET)\n"; \
		exit 1; \
	fi


# ============================================================================
# SCP - 10.32.9.138
# ============================================================================

138_scp: ## 上传文件到 10.32.9.138
	@if [ -z "$(strip $(ARGS))" ]; then \
		printf "$(COLOR_RED)✗ 未指定上传文件$(COLOR_RESET)\n"; \
		printf "\n$(COLOR_YELLOW)用法：$(COLOR_RESET)\n"; \
		printf "  make 138_scp ARGS=\"文件路径\"\n"; \
		printf "\n$(COLOR_YELLOW)示例：$(COLOR_RESET)\n"; \
		printf "  make 138_scp ARGS=\"bin/kraken-linux-arm64\"\n"; \
		printf "  make 138_scp ARGS=\"bin/kraken-linux-amd64\"\n"; \
		exit 1; \
	fi

	@printf "\n$(COLOR_GREEN)SCP 上传$(COLOR_RESET)\n"
	@printf "  文件: $(ARGS)\n"
	@printf "  目标: $(SCP_TARGET_138)\n"

	@if scp $(ARGS) "$(SCP_TARGET_138)"; then \
		printf "$(COLOR_GREEN)✓ SCP 上传成功$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_RED)✗ SCP 上传失败$(COLOR_RESET)\n"; \
		exit 1; \
	fi


# ============================================================================
# 清理
# ============================================================================

clean: ## 清理构建产物
	@printf "$(COLOR_GREEN)清理构建产物...$(COLOR_RESET)\n"

	@if [ -d "$(BINARY_DIR)" ]; then \
		rm -rf "$(BINARY_DIR)"; \
		printf "$(COLOR_GREEN)✓ 已删除: $(BINARY_DIR)$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_YELLOW)目录不存在，跳过: $(BINARY_DIR)$(COLOR_RESET)\n"; \
	fi


# ============================================================================
# 深度清理
# ============================================================================

clean-all: ## 深度清理（包括 Go 缓存）
	@printf "$(COLOR_GREEN)执行深度清理...$(COLOR_RESET)\n"

	@$(MAKE) clean

	@if [ -f coverage.out ]; then \
		rm -f coverage.out; \
		printf "$(COLOR_GREEN)✓ 已删除: coverage.out$(COLOR_RESET)\n"; \
	fi

	@if [ -f coverage.html ]; then \
		rm -f coverage.html; \
		printf "$(COLOR_GREEN)✓ 已删除: coverage.html$(COLOR_RESET)\n"; \
	fi

	@printf "$(COLOR_YELLOW)清理 Go 编译缓存...$(COLOR_RESET)\n"
	@$(GO) clean -cache -testcache

	@printf "$(COLOR_GREEN)✓ 深度清理完成$(COLOR_RESET)\n"


# ============================================================================
# Help
# ============================================================================
#
# 自动读取：
#
#   target: ## description
#
# 新增 target 时只需要：
#
#   xxx: ## xxx 功能
#
# 即可自动显示。
#
# ============================================================================

help: ## 显示可用目标
	@printf "\n"
	@printf "$(COLOR_BLUE)Kraken Makefile$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)═══════════════════════════════════════════════════════════════$(COLOR_RESET)\n"
	@printf "\n"

	@awk 'BEGIN { \
		FS = ":.*## "; \
		printf "$(COLOR_GREEN)可用命令:$(COLOR_RESET)\n\n" \
	} \
	/^[a-zA-Z0-9_.-]+:.*## / { \
		printf "  $(COLOR_CYAN)%-18s$(COLOR_RESET) %s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)

	@printf "\n"