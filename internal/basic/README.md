# Kraken Basic 模块设计说明

> Basic 是 Kraken 的"主机级批处理"功能域，提供对远程主机的批量操作能力。
> 本文件作为 basic 模块的统一说明入口，涵盖整体链路、SSH 设计、check 框架规划。
> 新增/调整 basic 相关行为时，请同步更新本文件。

## 目录

1. [模块定位](#1-模块定位)
2. [整体执行链路](#2-整体执行链路)
3. [两种执行模式](#3-两种执行模式)
4. [命令面（bc 子命令）](#4-命令面bc-子命令)
5. [SSH 设计与认证策略](#5-ssh-设计与认证策略)
6. [check 检查框架（规划）](#6-check-检查框架规划)
7. [关键文件](#7-关键文件)
8. [扩展指南](#8-扩展指南)
9. [变更记录](#9-变更记录)

---

## 1. 模块定位

Basic 承担主机级、批量化、可编排的运维操作，区别于 cluster 的"集群生命周期"。两者可通过 `basic` 段共享 SSH 凭据。

```
Kraken
 ├── carrot（命令框架）→ 分组：basic / cluster
 ├── basic（本模块）   → 主机级批处理：cmd / script / adduser / check
 └── cluster          → 集群生命周期（kubeadm），复用 basic 段 SSH 凭据
```

---

## 2. 整体执行链路

```
bc <command>
   │
   ├─ 单机模式 --ssh-host=<IP>           批量模式 --config <file>+host_inventory
   │                                           │
   ▼                                           ▼
NewSingleExecutor()                    runner.BuildFromConfig(cfg)
   │                                           │
   └───────────────┬───────────────────────────┘
                   ▼
       executor.NewExecutor(Options)
       带 SSH 凭据（user / private_key / password）
                   │
                   ▼
       exec.Run(ctx, hosts, operation)
                   │
                   ▼
      utils.PrintResultAndCheck(results)
```

命令层（`cmd/basic/`）只负责"模式分发 + operation 闭包"；执行细节全部下沉到 `internal/basic/executor` 与 `internal/basic/runner`。

---

## 3. 两种执行模式

| 模式 | 触发方式 | SSH 参数来源 | 适用场景 |
|---|---|---|---|
| **单机** | `--ssh-host=<IP>` flag | flags：`--ssh-user` / `--ssh-port` / `--ssh-password` | 一次性、临时、bootstrap |
| **批量** | `--config <file>` + `host_inventory: true` | config：`basic.user` / `basic.private_key_path` / `basic.iplist` / `basic.concurrency` | 正式、多节点 |

模式分发由各子命令 `RunE` 内判断 `cmd.Flags().Changed("ssh-host")` 完成（见 cmd.go）。

---

## 4. 命令面（bc 子命令）

| 子命令 | 状态 | 说明 |
|---|---|---|
| `bc cmd -c <cmd>` | 已实现 | 远程执行单条 shell 命令 |
| `bc script <file>` | 已实现 | 将本地脚本 base64 编码经 SSH 管道远端执行 |
| `bc adduser` | 已实现 | 用 bootstrap 凭据创建管理用户并部署公钥 |
| `bc check` | 规划中 | 清单驱动的检查框架（见第 6 章） |
| `bc ping / fetch / copy` | 预留 | 后续扩展（basic.go 已有注释位） |

命令注册在 basic.go 的 `cmd.AddCommand(...)`。

---

## 5. SSH 设计与认证策略

### 5.1 认证核心（executor/ssh.go）

`Connect(host)` 的认证策略：

```
authMethods()：Password 非空 → ssh.Password(密码)
              Password 为空 + PrivateKey 非空 → loadPrivateKey() → ssh.PublicKeys(私钥)
              两者皆空 → 报错 "ssh auth methods is empty"
```

字段优先级（节点级覆盖全局）：
- User：`host.User` 优先，空则 `options.User`
- Port：`host.Port` 优先，无效则 `options.Port`，再兜底 22
- 认证：密码优先，私钥兜底

### 5.2 政策（与 cluster 一致）

批量/config 模式**仅私钥认证**，来源单一；单机模式维持 flag 密码。

```
认证来源：basic.private_key_path   （唯一）
登录用户：basic.user               （唯一）
并发数：  basic.concurrency        （唯一，cluster 复用）
```

私钥不可用时的兜底流程（人工）：单机 SSH → `adduser` + 写入公钥 → 回批量模式。

### 5.3 PingSSH（executor/verify.go）

```go
// PingSSH 对全部 hosts 做连接+鉴权校验，任一失败返回聚合错误。
// 供需要 SSH 操作的功能在执行前调用，作为连接级预防。
func PingSSH(ctx context.Context, hosts []Host, opts Options) error {
    results := NewExecutor(opts).Run(ctx, hosts, func(ctx, h, c) Result {
        return RunShell(ctx, h, c, "echo kraken-ssh-ok")
    })
    // 出现 error 返回："host x: ssh ping failed: ..."
}
```

角色定位（连接级预防，非 check 预检层）：

- **PingSSH = 连接级预防**：需要 SSH 操作的功能（如 cluster init/join）执行前，先调用确认"连不连得上"，快速暴露连接/凭据错误。
- **check 框架不依赖它作预检层** —— SSH 访问是每种 check 的隐含首判，`ssh` 类型靠连接会话直接收集并判定（见 §6）。
- **两者互补独立**：PingSSH 只管"能不能 SSH 干活"；check 只管"连上后配置/资源达不达标"。

### 5.4 配置示例（kraken.yaml）

```yaml
basic:
  user: root
  private_key_path: ~/.ssh/id_rsa
  concurrency: 5
  host_inventory: true
```

---

## 6. check 检查框架（规划）

> 状态：规划中，尚未实现。设计目标见下。

### 6.1 核心思想

把"检查什么"清单化外置，"怎么检查"类型化内置，逐项执行后聚合输出。

```
.check.yaml（声明检查项）
   │
   ▼ 加载 + 注册
Check Registry（按 type 查找工厂）
   │
   ▼ 逐项执行（复用 runner）
每 host × 每 check → 执行 Command → Validate 判定
   │
   ▼
聚合输出（host × check 矩阵）
```

### 6.2 清单文件（.check.yaml）

```yaml
checks:
  - name: disk-root
    type: disk
    params:
      path: "/"
      min_free_gb: 20

  - name: memory-total
    type: memory
    params:
      min_gb: 8

  - name: network-apiserver
    type: network
    params:
      host: 10.0.0.1
      port: 6443

  - name: kernel-info
    type: kernel
    # 信息采集型：输出内核版本 + swap 状态，供人工判断，无 params
```

### 6.2.1 加载来源（本地清单文件，懒生成）

检查项清单是**单个本地文件**（默认 `.check.yaml`），程序在首次运行时**自动生成**，不存在"内嵌服务端默认 + 外层覆盖"的两层模型。

```
bc check 启动
   │
   ▼ 定位清单路径
默认：工作目录 .check.yaml
指定：--check-file <path>
   │
   ├─ 文件不存在？
   │     ├─ 默认路径 → 用内置默认内容 生成 .check.yaml，并提示：
   │     │            "已生成默认 .check.yaml，直接编辑即可自定义检查项"
   │     └─ --check-file 指定 → 报错终止（显式指定，不自动生成）
   │
   ▼
读取该文件作为最终检查项（单源，顺序执行）
```

**默认内容模板**（内置在代码中，仅作首次生成的初始内容）

```yaml
# .check.yaml — 首次生成，可自由编辑
# 删除某检查项 = 不执行；修改 params = 调整阈值
checks:
  - name: ssh                       # sshd 服务优化检查（见 §6.3 ssh 类型）
    type: ssh
    params:
      hostkey_algo: ed25519
      password_auth: "no"
      permit_root_login: prohibit-password
  - name: disk-root
    type: disk
    params:
      path: "/"
      min_free_gb: 5
  - name: memory-total
    type: memory
    params:
      min_gb: 2
```

**连接 = 隐含首判（无独立预检层）**

SSH 访问是每种 check 的**天然第一道判定**，不单独做预检：

- 目标主机**连不上 / 鉴权失败** → 该 host 所有检查项直接 FAIL（`connection/authentication failed`）。
- 连接凭据**统一来自 basic 段**（`user` + `private_key_path`），内部读取，用户感知上"就是 basic 里配的身份"。
- 只有**连得上**才继续收集输出、执行逐条 Validate。

因此 `ssh` 也是一个普通 check type：先靠基础连接会话收集 sshd 信息，再判定——**连接检查并入该项，而不是分开跑一次预检**。

**设计特性**

- **生成本地文件**：默认内容只在首次生成时写入 `.check.yaml`，运行逻辑不依赖代码常量内存态。
- **修改即自定义**：用户编辑该文件（增删检查项、改阈值）即生效。
- **不修改也没事**：文件若被删除/不存在，下次运行会自动重新生成，始终可用。
- **单源无合并**：一份文件就是最终执行清单，语义简单、可预测。
- **连接凭据不变**：basic 配置全局唯一，check 内部复用，用户无感知。

**路径规则**

| 场景 | 行为 |
|---|---|
| 未指定 `--check-file`，`.check.yaml` 存在 | 读取该文件 |
| 未指定 `--check-file`，`.check.yaml` 不存在 | 自动生成本地默认文件，再读取 |
| 指定 `--check-file`，文件存在 | 读取指定文件 |
| 指定 `--check-file`，文件不存在 | 报错终止（不自动生成） |

### 6.3 Check 接口 + 注册表

```go
type Check interface {
    Name() string
    Command() string                                 // 目标主机执行的 shell 命令
    Validate(stdout string) (ok bool, detail string) // 解析输出判定
}

type Registry struct{ /* map[type] -> factory */ }
func RegisterCheck(t string, factory func(params map[string]any) Check)
```

内置类型：

| type | Command | Validate |
|---|---|---|
| `ssh` | 靠基础连接会话收集 sshd -T 信息 | 逐 param 判定（hostkey_algo / password_auth / permit_root_login） |
| `disk` | `df -P <path>` | 剩余 >= min_free_gb |
| `memory` | `cat /proc/meminfo` | 总内存 >= min_gb |
| `network` | 探测 host:port 连通 | dial 成功 = PASS |
| `kernel` | `uname -r` + `cat /proc/swaps` | 信息采集型：不做判定，输出内核版本 + swap 状态供人工判断 |

**ssh 类型（sshd 服务优化检查）**

针对 sshd 服务的配置检查，判定连接本身也是该项的一部分：

```yaml
- name: ssh-config
  type: ssh
  params:
    hostkey_algo: ed25519          # 期望认证密钥类型
    password_auth: "no"            # 是否应关闭密码认证
    permit_root_login: prohibit-password
    # verify_key_user: appuser     # 可选：验证指定用户能免密
```

参数到检查项的映射（内部实现）：

| param | 收集方式 | 判定 |
|---|---|---|
| `hostkey_algo` | 会话内查询 sshd 支持的 hostkey 算法 | 支持该算法 = PASS |
| `password_auth` | `sshd -T` 的 `PasswordAuthentication` | 等于期望值 |
| `permit_root_login` | `sshd -T` 的 `PermitRootLogin` | 等于期望值 |
| `verify_key_user` | 对指定用户做免密探测（`sudo -n`/读 authorized_keys） | 可免密 = PASS |

> 连接身份用 basic 段凭据；params 里的 user（如 `verify_key_user`）是**检查目标**，非连接身份。二者内部区分，用户无感知。

### 6.4 执行与聚合

```go
// 逐项
result := runner.Run(ctx, cfg, func(ctx, host, client) Result {
    return executor.RunShell(ctx, host, client, check.Command())
})
result.Stdout -> check.Validate() -> Result{Name,Type,OK,Detail}
```

连接失败时：`RunShell` 的 Error 非空 → 该 host 该项直接 FAIL（`connection/authentication failed`）。

聚合输出：

```
Host      | Check         | Status | Detail
10.0.0.1  | ssh-config    | PASS   | hostkey=ed25519, PasswordAuth=no
10.0.0.1  | disk-root     | PASS   | free=120GB
10.0.0.1  | memory-total  | FAIL   | 4GB < 8GB
```

### 6.5 落地结构（规划）

```
cmd/basic/check.go            <- bc check 命令（加载清单、遍历、聚合）
internal/basic/check/
    ├── registry.go           <- Check 接口 + 注册表
    ├── manifest.go           <- 清单懒生成 + 解析
    └── builtins/
        ├── ssh.go / disk.go / memory.go / network.go / kernel.go
```

> `ssh.go` 的 `Command()` 依托 runner 已建立的连接会话收集 sshd 信息并判定；连接身份取 basic 段，params 中 `verify_key_user` 等为检查目标，内部区分。

**manifest.go（懒生成 + 解析）**

```go
const defaultManifestYAML = `...`  // 内置默认清单内容（仅作首次生成初始值）

// Manifest 清单结构（.check.yaml 顶层）
type Manifest struct {
    Checks []CheckItem `yaml:"checks"`
}

type CheckItem struct {
    Name   string         `yaml:"name"`
    Type   string         `yaml:"type"`
    Params map[string]any `yaml:"params"`
}

// LoadManifest 加载检查项清单。
//   - path 为空 → 默认 .check.yaml
//   - 默认路径不存在 → 用 defaultManifestYAML 生成文件，再解析
//   - 显式 path 不存在 → 返回错误（不自动生成）
func LoadManifest(path string) (*Manifest, error)

// ensureDefaultFile 兜底生成：默认路径缺失时写入默认内容并提示
func ensureDefaultFile(path string) error
```

**registry.go（Check 接口 + 注册表）**

```go
type Check interface {
    Name() string
    Command() string                                  // 目标主机执行的 shell 命令
    Validate(stdout string) (ok bool, detail string)  // 解析输出判定
}

type Factory func(params map[string]any) Check

type Registry struct{ factories map[string]Factory }
func RegisterCheck(t string, f Factory)
func (r *Registry) Build(item CheckItem) (Check, error)  // 按 type 查工厂，未知 type 报错
```

**cmd/basic/check.go（命令编排）**

```go
// 流程：LoadManifest → 逐项 Registry.Build → 复用 runner 执行 → Validate → 聚合
func NewCheckCmd() *cobra.Command {
    // --check-file <path>：默认空 → 走懒生成
    // 单机 --ssh-host / 批量 config 复用 basic 模式分发
}
```

**builtins/（内置检查型，每文件一个 type）**

```go
// disk.go
type DiskCheck struct{ path string; min int }    // df -P <path> → 剩余 >= min(GB)
// memory.go
type MemoryCheck struct{ min int }               // /proc/meminfo → 总内存 >= min(GB)
// network.go
type NetworkCheck struct{ host string; port int }// 探测 host:port 连通
// kernel.go 信息采集型：不做判定，输出 (内核版本 + swap 状态) 供人工判断
type KernelCheck struct{} // uname -r + /proc/swaps → detail 输出原始采集信息，恒 ok

// 各文件 v 通过 init() 调用 RegisterCheck 注册工厂
func init() { RegisterCheck("disk", func(p map[string]any) Check { return &DiskCheck{...} }) }
```

---

## 7. 关键文件

| 文件 | 职责 |
|---|---|
| cmd/basic/basic.go | bc 根命令，注册子命令与全局 flag |
| cmd/basic/cmd.go | cmd 子命令，单机/批量分发 |
| cmd/basic/script.go | script 子命令，base64 脚本远端执行 |
| cmd/basic/adduser.go | adduser 子命令，bootstrap 创建用户 |
| cmd/basic/flags.go | SSH flags + NewSingleExecutor / NewBatchExecutor |
| cmd/basic/check.go | check 子命令（清单加载、遍历、聚合） |
| internal/basic/check/manifest.go | 清单懒生成 + 解析（CheckItem / LoadManifest / ensureDefaultFile） |
| internal/basic/check/registry.go | Check 接口 + 注册表（RegisterCheck / Build） |
| internal/basic/check/builtins/*.go | 内置检查型：ssh / disk / memory / network / kernel |
| internal/basic/executor/executor.go | Executor.Run(hosts, op) 并发执行 |
| internal/basic/executor/ssh.go | SSHOptions + 认证策略 |
| internal/basic/executor/shell.go | RunShell |
| internal/basic/executor/host.go | Host 定义 |
| internal/basic/executor/cmd.go | RunOnHosts 批量同命令 |
| internal/basic/executor/verify.go | PingSSH 连接级预防 |
| internal/basic/runner/runner.go | 从 config 构建 hosts + 执行 |
| utils/print_result.go | 结果输出与聚合 |
| internal/basic/SSH_README.md | SSH 专项文档（本条为概览） |

---

## 8. 扩展指南

### 新增一个 bc 子命令

1. 在 cmd/basic/xxx.go 写 NewXxxCmd()，内部复用模式分发逻辑。
2. 在 basic.go 的 cmd.AddCommand(...) 注册（同步撤销对应预留注释）。
3. 若涉及远程操作，复用 runner.BuildFromConfig + executor.Run。

### 新增一个 check 检查项

1. 在 internal/basic/check/builtins/ 实现一个 Check 工厂。
2. 在注册表 init() 内 RegisterCheck("new-type", factory)。
3. 在 .check.yaml 中声明 type: new-type 引用。

### 调整 SSH 认证政策

同步修改 ssh.go、本文件第 5 章、SSH_README.md 三处，保持一致。

---

## 9. 变更记录

| 日期 | 变更 |
|---|---|
| 2026-09-29 | 建立 basic 模块整体说明；确认批量/config 仅私钥认证、单一来源；TestSSH 更名 PingSSH（连接级预防，非 check 预检层）；新增 check 框架设计（规划态）：**合并模型**（连接=隐含首判，ssh 为 check 类型，无独立预检层，basic 连接凭据不变）；cluster 复用 basic 凭据 |