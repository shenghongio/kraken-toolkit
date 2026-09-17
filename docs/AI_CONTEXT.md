# Kraken Pedestal — 现状接管报告

> 生成日期：2026-09-15
> 用途：AI 助手与开发者协作时的共享上下文，维护项目"地图"和关键约定。

---

## 1. 项目地图

### 1.1 目录结构

```
.
├── main.go                     # 入口：panic recovery → logger → Cobra.Execute
├── go.mod / go.sum             # Go 1.26.0, 依赖: cobra, slog-console, x/crypto, yaml.v3
├── makefile                    # 构建/测试/发布/S Sh打包/SCP 上传
├── README.md                   # 项目说明（功能、架构、路线图）
│
├── cmd/                        # Cobra 命令层（薄层，仅做编排）
│   ├── root.go                 # NewRootCmd(): 全局 flags, 初始化日志, command groups
│   ├── version.go              # kraken version [--json]
│   ├── initconfig.go           # TODO: kraken initconfig
│   └── basic/
│       ├── basic.go            # kraken basic (目前空壳，subcommands 全部被注释)
│       ├── flags.go            # Basic 模块的持久化 flags
│       └── README.md           # Basic 模块流程说明
│
├── internal/
│   ├── basic/                  # Basic 模块核心实现
│   │   ├── executor/           # 并发 SSH 执行引擎 ★ 核心
│   │   │   ├── executor.go     # Result, Options, Executor.Run (semaphore 并发控制)
│   │   │   ├── executor_test.go# 单元测试（mockSSHConnector）
│   │   │   ├── shell.go        # runShell (单机) + Shell (批量)
│   │   │   └── ssh.go          # SSHConnection 接口 + SSHClient 实现
│   │   └── inventory/          # 主机清单
│   │       ├── host.go         # Host 结构体, ParseFile, parseHostLine
│   │       ├── inventory.go    # Inventory 结构体, Load, Resolve, Deduplicate
│   │       └── inventory_test.go
│   ├── config/                 # 全局配置结构 + YAML 加载
│   │   ├── config.go           # Config (Log, Basic, Cluster, Deploy, Middleware)
│   │   ├── load.go             # Load(path) + Validate
│   │   ├── basic.go            # BasicConfig / BasicSSHConfig / BasicConfigHistory
│   │   ├── cluster.go          # ClusterConfig (TODO)
│   │   ├── deploy.go           # DeployConfig (TODO)
│   │   └── middleware.go       # MiddlewareConfig (TODO)
│   └── deploy/                 # 部署模块 (README 仅一行)
│       └── README.md
│
├── pkg/                        # 可复用公共包
│   ├── cli/
│   │   ├── groups.go           # 命令分组常量 (basic, settings, other...)
│   │   └── version.go          # 版本信息 (ldflags 注入)
│   └── logger/
│       ├── logger.go           # slog 初始化 + console-slog handler
│       └── kerror.go           # 自定义错误 Code/Error/Wrap
│
├── utils/
│   └── printroot_help.go       # Cobra help 自定义渲染
│
├── doc/
│   └── ssh-gen-key.md          # SSH 密钥生成命令文档
│
├── .trae/
│   └── skills/code/SKILL.md    # Trae AI 协作 Skill
│
└── bin/kraken                  # 编译产物（已存在）
```

### 1.2 入口与启动流程

```
main()
  ├── logger.Init()             # 兜底日志（PersistentPreRunE 之前的错误）
  ├── defer recover()           # panic → slog.Error + exit(1)
  ├── cmd.NewRootCmd()          # 构造 Cobra 命令树
  │     ├── PersistentPreRunE:  # 初始化日志（log-level, log-source 参数）
  │     ├── AddCommand: basic / version / completion
  │     └── SetHelpFunc: 自定义 help 输出
  └── rootCmd.Execute()         # Cobra 执行
        └── handleError(err)     # 统一错误入口
```

### 1.3 核心模块依赖关系

```
basic 命令
  └── internal/basic/executor
        ├── Options → SSHOptions
        ├── Executor → SSHConnection (接口)
        ├── NewExecutor → SSHClient (实现)
        ├── Run() → semaphore 并发控制
        │     └── operation(ctx, host, *ssh.Client) Result
        └── Shell() → runShell()

internal/basic/inventory
  └── Host{Address, User, Port, Passwd, Group}
  └── Inventory → Load(path), Resolve(targets), Deduplicate(hosts)

internal/config
  └── Config → BasicConfig / ClusterConfig / DeployConfig / MiddlewareConfig
  └── Load(path) → yaml.Unmarshal → Validate

pkg/logger → slog + console-slog
pkg/cli → version (ldflags) + groups
```

### 1.4 外部依赖

| 依赖 | 用途 | 版本 |
|------|------|------|
| github.com/spf13/cobra | CLI 框架 | v1.10.2 |
| github.com/phsym/console-slog | 彩色 slog handler | v0.3.1 |
| golang.org/x/crypto | SSH 客户端 | v0.57.0 |
| gopkg.in/yaml.v3 | YAML 解析 | v3.0.1 |

---

## 2. 运行方式

### 2.1 构建

```bash
make build              # 当前平台
make build-arm64        # Linux ARM64
make build-amd64        # Linux AMD64
make build-full         # 显示完整系统环境 + 构建
go build ./...          # 直接 Go 编译
```

编译产物：`bin/kraken`

### 2.2 测试

```bash
make test               # go test -v -count=1 ./... + 统计输出
make coverage           # 测试 + 覆盖率报告
go test ./...           # 直接 Go 测试
go test -v ./...        # 详细输出
```

### 2.3 代码检查

```bash
make fmt                # gofmt -w .
make fmt-check          # 检查格式
make vet                # go vet ./...
make check              # fmt-check + vet
```

### 2.4 发布

```bash
make release VERSION=v1.0.0    # check → test → build
```

### 2.5 SCP 上传

```bash
make 139_scp ARGS="bin/kraken"    # → 10.32.9.139
make 138_scp ARGS="bin/kraken"    # → 10.32.9.138
```

### 2.6 当前运行

```bash
./bin/kraken
# 输出：帮助信息（目前仅有 version / completion 可用）
```

---

## 3. 架构现状

### 3.1 分层设计

```
┌──────────────┐
│    CLI 层    │  cmd/ — Cobra 命令树，薄层编排
├──────────────┤
│   业务逻辑层  │  internal/basic/executor — 并发引擎+SSH操作
├──────────────┤
│  公共运行时层 │  pkg/logger, internal/config, pkg/cli
├──────────────┤
│  基础设施层   │  golang.org/x/crypto (SSH), yaml.v3, cobra
└──────────────┘
```

符合 README 中描述的 "CLI Should Stay Thin" 和 "Separate Operation From Presentation" 原则。

### 3.2 核心数据流（Basic Shell）

```
CLI Flags (concurrency, timeout, limit...)
  → cmd/basic 解析
    → internal/basic/executor.Shell(ctx, hosts, cmd)
      → Executor.Run(ctx, hosts, operation)
        ┌─────────────────────────────────────────┐
        │  for each host (semaphore 控制并发度)    │
        │    go func():                            │
        │      1. 检查 ctx.Done()                  │
        │      2. 获取 semaphore 槽位               │
        │      3. e.ssh.Connect(host) → *ssh.Client│
        │      4. operation(ctx, host, client)     │
        │      5. 释放 semaphore 槽位               │
        └─────────────────────────────────────────┘
        → []Result {Host, Success, ExitCode, Stdout, Stderr, Error, Duration}
```

### 3.3 关键设计决策

| 决策 | 理由 |
|------|------|
| `Executor` 依赖 `SSHConnection` 接口 | 可注入 mock，单元测试无需真实 SSH |
| semaphore = `chan struct{}` 控制并发 | 简单可靠，支持 ctx 取消 |
| results 按下标写入，无锁 | 每个 goroutine 只写自己的 `results[index]` |
| 操作与输出分离 | Result 结构体承载数据，CMD 层决定渲染 |
| 自定义 logger 包装 slog | 统一日志格式、颜色、时间戳 |
| 自定义 Error 类型 | 区分错误来源（SSH/网络/配置/权限） |

---

## 4. 已有功能清单

### 4.1 已完成 ✅

| 功能 | 位置 | 说明 |
|------|------|------|
| 基础 CLI 框架 | `cmd/root.go` | Cobra 命令树，分组，自定义 help |
| 版本查看 | `cmd/version.go` | `kraken version [--json]` |
| shell completion | `cmd/initconfig.go` | `kraken completion [bash\|zsh]` |
| 日志系统 | `pkg/logger/` | slog + console-slog，彩色输出，级别控制 |
| 错误模型 | `pkg/logger/kerror.go` | Code + Error + Wrap |
| 配置加载 | `internal/config/` | YAML 加载 + 校验（Basic 模块完备，其余占位） |
| SSH 客户端 | `internal/basic/executor/ssh.go` | 密码/私钥认证，超时控制，host key 校验 |
| 并发执行引擎 | `internal/basic/executor/executor.go` | semaphore 并发控制，ctx 取消，nil 防御 |
| Shell 命令执行 | `internal/basic/executor/shell.go` | `Shell()` 批量 + `runShell()` 单机 |
| 主机清单解析 | `internal/basic/inventory/host.go` | 类 Ansible inventory 文件解析 |
| 主机清单查询 | `internal/basic/inventory/inventory.go` | Load/Resolve(分组/地址/all)/Deduplicate |
| 单元测试（inventory） | `internal/basic/inventory/inventory_test.go` | 9 个测试用例，覆盖 ParseFile/Resolve/Deduplicate |
| 单元测试（executor） | `internal/basic/executor/executor_test.go` | 7 个测试用例，mockSSHConnector |
| 构建系统 | `makefile` | 格式化/检查/测试/构建/发布/SCP 全流程 |
| SSH 密钥生成文档 | `doc/ssh-gen-key.md` | 功能说明 + 命令示例 |

### 4.2 半成品/不可用 🔶

| 功能 | 位置 | 问题 |
|------|------|------|
| `kraken basic` | `cmd/basic/basic.go` | subcommands 全部被注释（NewShellCmd/NewCopyCmd/...） |
| `kraken basic shell` | — | 从未实现 |
| `kraken gen-key` | — | 文档存在但代码不存在（`doc/ssh-gen-key.md`） |
| Config 全局加载 | `cmd/root.go:37-39` | 被注释掉 |
| `--config` 全局 flag | `cmd/root.go:42-46` | 被注释掉 |
| `kraken initconfig` | `cmd/initconfig.go` | 空命令 |
| internal/deploy | — | 仅有 README 一行 |
| `loadPrivateKey` | `ssh.go:158-187` | 找到密钥但解析失败时直接返回，未尝试下一个 |
| StrictHostKey=true | `ssh.go:217` | 返回 "not implemented yet" |

### 4.3 TODO / 计划中 📋

| 功能 | 位置 | 备注 |
|------|------|------|
| Copy (SCP 上传) | 计划在 shell.go 同级 | `runCopy` + `Copy()` |
| Fetch (SCP 下载) | 计划在 shell.go 同级 | `runFetch` + `Fetch()` |
| Script (脚本执行) | 计划在 shell.go 同级 | `runScript` + `Script()` |
| Ping (连通性检测) | 预期 | — |
| Check (系统巡检) | 预期 | — |
| Cluster 模块 | `internal/config/cluster.go` | 仅有空 Config struct |
| Deploy 模块 | `internal/config/deploy.go` | 仅有空 Config struct |
| Middleware 模块 | `internal/config/middleware.go` | 仅有空 Config struct |
| Ops Tools | README 规划 | system/network/storage/diagnostics |
| GPU (HAMi/DRA) | README 规划 | 远期路线图 |
| Known hosts 校验 | `ssh.go:216` | 已标 TODO |
| 配置加载集成 | `root.go` | 已注释 |

---

## 5. 编码约定

### 5.1 命名

- **包名**：全小写，单数（`logger`, `inventory`, `executor`）
- **文件**：`snake_case.go`（`printroot_help.go`, `ssh-gen-key.md`）
- **类型/接口**：PascalCase（`SSHConnection`, `SSHClient`, `BasicConfig`）
- **函数/方法**：PascalCase（公开）、camelCase（私有）（`NewExecutor`, `runWithConnector`）
- **变量**：camelCase（`connectorErr`, `hostMap`, `lineNumber`）
- **常量**：PascalCase（`CodeInvalidArgument`, `DefaultSSHPort`）
- **测试函数**：`TestXxx`（`TestNew`, `TestExecutorRunEmptyHosts`）

### 5.2 目录结构

- `cmd/` — 可执行命令入口，每个子命令一个目录
- `internal/` — 不导出的实现代码，按领域分包
- `pkg/` — 可复用的公共包
- `docs/` — 项目文档
- `doc/` — 功能文档（与 docs/ 并存，需统一）

### 5.3 错误处理

- 自定义错误类型 `logger.Error`，包含 `Code`, `Message`, `Err`
- `logger.Wrap(code, msg, err)` 包装底层错误
- `errors.Is()` / `errors.As()` 检查错误链
- `main.go` 中 `handleError()` 统一入口
- slog 记录错误时带 code 和 stack

### 5.4 日志

- 使用 `log/slog` 标准库 + `console-slog` handler
- 初始化在 CLI 的 PersistentPreRunE 中
- `slog.Debug/Info/Warn/Error` 结构化日志
- 时间格式 `[ 06-01-02/15:04:05 ]`

### 5.5 测试风格

- 使用 Go 标准 `testing` 包
- 使用 `t.TempDir()` 创建临时文件
- 表驱动测试（`tests []struct{...}` + `t.Run`)
- mock 通过接口注入（`mockSSHConnector` 实现 `SSHConnection`）
- 子测试清晰命名（`"default"`, `"custom"`）
- 断言使用 `t.Fatalf`（致命）和 `t.Errorf`（非致命）
- `t.Parallel()` 未使用（目前单线程执行即可）

### 5.6 防御性编程模式

- nil receiver 检查（`Options()`, `Connect()`, `Validate()`）
- nil ctx 回退到 `context.Background()`
- 空 hosts 直接返回 nil
- nil operation 返回错误结果
- 循环变量捕获 `index := index`

---

## 6. 技术债与风险

### 6.1 Bug 🐛

| 严重度 | 问题 | 位置 | 说明 |
|--------|------|------|------|
| 🔴 | `mockSSHConnector.Connect` 条件反转 | `executor_test.go:32` | `m.connect == nil` 写成了 `m.connect != nil`，导致所有依赖 mock 的测试失败 |
| 🔴 | `loadPrivateKey` 失败不兜底 | `ssh.go:181-185` | 如果 `id_ed25519` 文件损坏返回错误，不会尝试 `id_rsa` |
| 🟡 | `runShell` 未提取 ExitCode | `shell.go:29-31` | `ExitCode` 永远是 0 |
| 🟡 | `runShell` 未监听 ctx 取消 | `shell.go:25` | `session.Run` 阻塞，CTRL+C 可能无法立即退出 |
| 🟡 | `TestExecutorConcurrency` 断言 3 条件反了 | `executor_test.go` | `maxConcurrency.Load() >= 2` 应为 `> 2` |

### 6.2 代码坏味道

| 问题 | 位置 | 建议 |
|------|------|------|
| `doc/` 与 `docs/` 并存 | 根目录 | 统一文档目录 |
| `basic.go` subcommands 全部注释 | `cmd/basic/basic.go` | 或者删除，或者实现 |
| `internal/deploy/README.md` 仅一行 | `internal/deploy/` | 冗余文件 |
| `PKG_VERSION` 指向不存在路径 | makefile:73 | `github.com/kraken-pedestal/internal/deploy/executor` 不存在 |
| Config 全局加载被注释 | `root.go` | 配置模型已就绪但未接入 CLI |

### 6.3 缺失测试

| 模块 | 风险 |
|------|------|
| `executor.Run` — 连接成功 + operation 正常执行 | 无完整端到端测试 |
| `executor.Run` — operation 返回错误 | 无覆盖 |
| `ssh.go` — `authMethods`, `loadPrivateKey` | 依赖文件系统，需 mock |
| `shell.go` — `Shell` + `runShell` | 全无测试 |
| `config.go` — `Config.Load` + `Validate` | 全无测试 |
| `logger.go` — `ParseLevel`, `Init` | 全无测试 |

### 6.4 安全隐患

| 问题 | 位置 | 说明 |
|------|------|------|
| `StrictHostKey=false` 是默认行为 | `ssh.go:205-207` | MITM 风险，当前注释说"保持 fastdp 行为兼容" |
| `KnownHostsFile` 校验未实现 | `ssh.go:209-217` | 开启 StrictHostKey 会直接报错 |
| 密码在内存中以 string 存在 | `ssh.go:85-88` | 非 Go 特有，但值得注意 |

### 6.5 架构风险

| 问题 | 说明 |
|------|------|
| goroutine 数 = hosts 数 | 1 万台主机会创建 1 万个 goroutine，建议 worker pool |
| Connect 不支持 ctx | SSH 连接期间 ctx 取消不感知 |
| 无 recover 保护 operation | operation 内部 panic 会导致进程崩溃 |
| `client.Close()` 未判 nil | Connect 返回 `(nil, nil)` 时会 panic |

---

## 7. 后续开发建议

### 7.1 优先修复（P0）

1. **fix: mockSSHConnector.Connect 条件反转** — 一个字符改动，所有测试恢复
2. **fix: runShell 的 ctx 取消 + ExitCode** — 提升 shell 模块生产可用性
3. **fix: loadPrivateKey 失败兜底** — 密钥文件损坏时自动尝试下一个

### 7.2 增量扩展（P1）

按以下顺序，每次一个垂直切片：

```
1. shell_test.go          → 为 Shell + runShell 添加单元测试
2. script.go              → Script 方法（bash -s 传脚本）
3. copy.go                → Copy 方法（SCP 上传）
4. fetch.go               → Fetch 方法（SCP 下载）
5. basic shell 命令       → 接通 cmd/basic → executor.Shell
6. --limit / 主机清单集成  → CLI flag → inventory.Load → executor.Shell
```

### 7.3 可优化但不紧急（P2）

- worker pool 取代 per-host goroutine（大规模优化）
- `Connect(ctx)` 接口增加 context 参数
- 添加 `operation` recover
- 补齐 config/logger 的测试
- 统一 `doc/` 和 `docs/` 目录

### 7.4 不要碰的区域

| 区域 | 理由 |
|------|------|
| `pkg/logger/kerror.go` 的错误 Code 枚举 | 其他模块可能依赖，改动影响全局 |
| `cmd/root.go` 的命令分组 ID | 多个命令引用 `cli.Group*` 常量 |
| `main.go` 的 panic recovery 逻辑 | 全局兜底，改动需全量回归 |
| `makefile` 的 SCP 目标主机 IP | 可能是团队内部地址 |
| `README.md` 的路线图/规划 | 项目对外文档，需团队确认后方可修改 |

---

## 8. 需要补充的信息

为更好地协作，以下信息有助于降低沟通成本：

### 8.1 建议提供

| 信息 | 用途 |
|------|------|
| 项目 Roadmap 优先级 | 决定功能实现顺序 |
| 目标部署环境的 SSH 配置 | 辅助测试 SSH 连接场景 |
| 是否已有 CI/CD 流水线 | 决定测试/构建策略 |
| 团队约定的 Go 版本管理工具 | go env / devbox / asdf |
| 是否有 k8s 集群可测试 | 影响 Cluster 模块开发 |
| gen-key 功能的实现优先级 | 文档已有但代码缺失 |
| --limit / 主机清单功能的优先级 | 影响 cmd/basic 的集成顺序 |

### 8.2 已知假设

（以下是我分析中假设成立，如果不符合请纠正）

- 项目当前重点是 `basic` 模块（SSH 批量操作）
- `Kubernetes/Helm/GPU` 模块是远期规划，当前不阻塞
- `internal/deploy` 尚未开始开发
- 团队使用 macOS 开发，部署目标为 Linux 服务器
- SCP 目标（`10.32.9.139/138`）是测试/预发环境
- 测试目前仅在本地运行，无 CI 集成
- 当前 Go 版本 1.26 的 `slog` 特性可用（本项目已启用）

---

## 附录：常用命令速查

```bash
# 开发迭代
make fmt && make check && make test && make build

# 运行测试并查看详细输出
go test -v -count=1 ./internal/basic/...

# 运行单个测试
go test -v -run TestExecutorRunEmptyHosts ./internal/basic/executor/

# 查看当前可用命令
./bin/kraken

# 查看版本
./bin/kraken version
```