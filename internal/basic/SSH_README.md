# Kraken Basic SSH 执行设计说明

> 本文件记录 basic 模块的 SSH 执行模式、认证策略与配置约定，供后续功能（如 cluster）统一遵循。修改 SSH 相关行为时，请同步更新本文档。

## 1. 两种执行模式

basic 提供两种 SSH 执行模式，通过入口参数区分：

| 模式 | 触发方式 | 归属 |
|---|---|---|
| **单机模式** | 命令行 flag `--ssh-host=<IP>` | 一次性、临时、bootstrap |
| **批量模式** | `--config <file>` + 配置 `host_inventory: true` | 正式、多节点 |

### 1.1 单机模式（保持不变）

- 目标主机、登录用户、认证方式全部来自命令行 flag，不进配置文件。
- 由 `basic.NewSingleExecutor()` 构建。

```go
--ssh-host=<IP>     必须，目标主机
--ssh-user=<user>   默认 root
--ssh-port=<port>   默认 22
--ssh-password=...  密码认证（临时场景可用）
```

### 1.2 批量模式（config 驱动）

- 从 `kraken.yaml` 的 `basic` 段读取 SSH 凭据与并发数。
- 由 `basic.NewBatchExecutor()` 构建，内部调用 `runner.BuildFromConfig()` + `runner.Run()`。

## 2. 认证策略（唯一来源，仅私钥）

批量 / config 模式 **只使用私钥认证**，不使用密码。

```
认证来源：basic.private_key_path     （唯一）
登录用户：basic.user                 （唯一）
并发数：  basic.concurrency          （唯一，cluster 复用）
目标主机：basic.iplist 或 cluster.nodes
```

### 2.1 理由

- 单一身份来源，无密码分支，逻辑简单。
- 私钥认证失败时，错误信息天然能定位"哪台主机、哪个 user、凭据无效"。
- 密码安全：密钥路径落盘在配置文件中，但密码不进配置、不进代码。
- bootstrap（首次装机）本就需要人工介入：先单机 SSH `adduser` 并把公钥写入 `authorized_keys`，随后才能进入 config 批量模式。

### 2.2 认证兜底原则

私钥不可用时的兜底流程（**不属于本模块，属人工流程**）：

```
config 批量模式
   └─ 私钥认证失败
        └─ 单机模式 ssh 手动 bootstrap：adduser + 写入公钥
             └─ 成功后回到 config 批量模式
```

## 3. 认证预检（TestSSH）

在执行任何批量命令前，先对全部目标主机做一次连接 + 鉴权预检，避免中途批量失败后才暴露凭据错误。

```go
// basic/executor/verify.go
// TestSSH 对全部 hosts 做连接+鉴权预检，任一失败返回聚合错误
func TestSSH(ctx context.Context, hosts []Host, opts Options) error {
    results := NewExecutor(opts).Run(ctx, hosts, func(ctx context.Context, h Host, c *ssh.Client) Result {
        return RunShell(ctx, h, c, "echo kraken-ssh-ok")
    })
    return nil  // 存在 error 时返回："host x: ssh authentication failed: ..."
}
```

调用位置：批量执行命令（bc、cluster init/join）在 `RunOnHosts` / `Runner.Run` 之前调用一次。

## 4. 三种目标的职责划分

避免"优先级覆盖"心智负担，三类信息各自独立、互不覆盖：

| 类别 | 来源 | 说明 |
|---|---|---|
| **认证凭证** | `basic.user` + `basic.private_key_path` | 唯一来源，无覆盖 |
| **连接目标** | `cluster.nodes[i].address/port` | 每节点独立的元数据，非认证 |
| **并发数** | `basic.concurrency` | 唯一来源，bc 与 cluster 共用 |

## 5. 配置示例（kraken.yaml）

```yaml
# ============ 全局 SSH 凭据（唯一来源，仅私钥认证）============
basic:
  user: root                                # 登录用户
  private_key_path: ~/.ssh/id_rsa           # 唯一认证方式
  concurrency: 5                            # 批量并发数（cluster 复用）
  host_inventory: true                      # 启用批量模式

# ============ Cluster 模块 ============
cluster:
  name: my-cluster
  kubernetes_version: 1.30.0

  pod_cidr: 10.244.0.0/16
  service_cidr: 10.96.0.0/12

  control_plane:
    endpoint: 10.0.0.1:6443

  container_runtime:
    type: containerd
    socket: /run/containerd/containerd.sock

  nodes:
    - address: 10.0.0.1
      role: control-plane
    - address: 10.0.0.2
      role: worker
```

### 5.1 Cluster 复用 basic 凭据

cluster 命令（init/join）不新增 SSH 配置段，直接从 `basic` 读取：

```go
opts := executor.Options{
    Concurrency: cfg.Basic.Concurrency,
    SSH: executor.SSHOptions{
        User:       cfg.Basic.User,
        PrivateKey: cfg.Basic.PrivateKeyPath,   // 唯一认证
        Timeout:    5 * time.Second,
    },
}
```

## 6. 关键文件

| 文件 | 职责 |
|---|---|
| `internal/basic/executor/host.go` | `Host` 定义（address/user/port） |
| `internal/basic/executor/executor.go` | `Executor.Run(hosts, operation)` 并发执行 |
| `internal/basic/executor/shell.go` | `RunShell(ctx, host, client, cmdStr)` |
| `internal/basic/executor/ssh.go` | `SSHOptions` + 认证（Password 优先 / PrivateKey 兜底） |
| `internal/basic/executor/cmd.go` | `RunOnHosts` 批量执行同一命令 |
| `internal/basic/executor/verify.go` | 【新增】`TestSSH` 认证预检 |
| `internal/basic/runner/runner.go` | 从 config 构建 hosts 并执行 |
| `cmd/basic/cmd.go` | 单机/批量模式入口分发 |
| `cmd/basic/flags.go` | 单机模式 SSH flag |

## 7. 变更记录

| 日期 | 变更 |
|---|---|
| 2026-09-29 | 确认批量/config 模式仅私钥认证；单一认证来源（basic 段）；新增 TestSSH 认证预检；cluster 复用 basic 凭据；单机模式保持不变；密码场景从 config 体系移除，兜底走人工 adduser |