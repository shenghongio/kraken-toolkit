# Kraken Cluster Module Requirements

> 基于 kubeadm phase 的 Kubernetes 集群管理模块需求文档
>
> 状态：设计阶段 | 版本：v0.1 | 日期：2026-09-21

---

## 目录

1. [概述](#1-概述)
2. [术语定义](#2-术语定义)
3. [功能需求](#3-功能需求)
4. [架构设计](#4-架构设计)
5. [接口设计](#5-接口设计)
6. [配置设计](#6-配置设计)
7. [命令设计](#7-命令设计)
8. [执行流程](#8-执行流程)
9. [文件结构](#9-文件结构)
10. [开发计划](#10-开发计划)
11. [扩展性范例](#11-扩展性范例)
12. [设计决策记录](#12-设计决策记录)

---

## 1. 概述

### 1.1 项目背景

Kraken Pedestal 是一个 Go 语言编写的统一 CLI 工具，用于部署和运维 Kubernetes 集群及中间件。当前项目已实现 Basic 模块（主机批量操作），但 Cluster 模块尚未开发。

### 1.2 目标

将 `kubeadm phase` 机制引入 Kraken，实现 Kubernetes 集群生命周期的阶段化管理。核心目标：

1. **阶段化**：将 kubeadm init / join / reset / upgrade 拆分为独立可执行的阶段（Phase）
2. **可编排**：支持按依赖顺序执行全部阶段，或单独执行指定阶段
3. **可观测**：提供 dry-run 预览、阶段列表、依赖关系可视化
4. **可扩展**：新增集群操作只需实现 Phase 接口并注册，无需改动核心框架
5. **复用基础设施**：SSH 执行器、配置加载、日志、结果打印全部复用现有组件

### 1.3 覆盖范围

- kubeadm init（12 个阶段）：preflight → certs → ... → addon
- kubeadm join（5 个阶段）：支持 control-plane 和 worker 两种角色
- kubeadm reset（后续版本）
- kubeadm upgrade（后续版本）
- 阶段管理子命令：list（查看）/ run（执行指定阶段）

---

## 2. 术语定义

| 术语 | 定义 |
|---|---|
| **Phase** | kubeadm 操作的最小原子单元，如 preflight、certs、kubelet-start 等 |
| **PhaseRegistry** | 阶段注册器，管理阶段的注册、拓扑排序和批量执行 |
| **Kubeadm Wrapper** | 对 kubeadm 二进制的封装层，负责二进制发现、配置生成、命令构建 |
| **ExecuteFunc** | 执行函数类型，由调用方注入（本地执行或 SSH 远程执行） |
| **Action** | 集群操作的类别：init / join / reset / upgrade |
| **拓扑排序** | 基于阶段依赖关系（Dependencies）的排序算法，使用 Kahn 算法 |
| **Plane** | 多平面部署中的一组节点 + 它们要执行的 Phase 链。如 cp-1（第一个控制平面，执行 init）、workers（worker 节点组，执行 join） |
| **ClusterDeployer** | 多平面部署编排器，按平面依赖顺序串行执行各平面，平面内节点并发执行 |
| **RemoteExecutor** | SSH 远程执行器，桥接 Phase 层和 basic/executor，将 `Phase.Command()` 转为 SSH shell 命令 |
| **LocalExecutor** | 本地执行器，使用 `os/exec` 在本地执行 kubeadm 命令，适合单节点调试和 CI |

---

## 3. 功能需求

### 3.1 Phase 抽象层

#### FR-1.1 Phase 接口定义

定义统一的 Phase 接口，包含以下方法：

| 方法 | 返回值 | 说明 |
|---|---|---|
| `Name() string` | 阶段名称 | 对应 kubeadm 的阶段名，如 "preflight" |
| `Description() string` | 阶段描述 | 用于帮助信息和日志输出 |
| `Command(cfg PhaseConfig) []string` | 命令切片 | 构建实际要执行的 kubeadm 命令 |
| `Dependencies() []string` | 依赖列表 | 当前阶段依赖的前置阶段名列表 |

#### FR-1.2 PhaseRegistry 注册器

- 支持注册、覆盖、查询阶段
- 基于 Kahn 算法实现拓扑排序
- 检测循环依赖，存在循环时返回错误
- 支持 RunAll（全部执行）和 RunOne（单个+依赖执行）
- 执行过程中任一阶段失败则终止后续

#### FR-1.3 PhaseConfig 执行上下文

```go
type PhaseConfig struct {
    KubeadmConfigPath string            // kubeadm 配置文件路径
    KubeadmBinary     string            // kubeadm 二进制路径
    NodeAddress       string            // 目标节点地址
    DryRun            bool              // 是否仅预览命令
    ExtraArgs         map[string]string // 额外参数
}
```

### 3.2 kubeadm Wrapper 层

#### FR-2.1 二进制发现

- 自动查找 kubeadm 路径，优先级：
  1. 配置显式指定路径
  2. `$PATH` 中的 kubeadm
  3. `/usr/bin/kubeadm`
- 执行版本校验，确保兼容性

#### FR-2.2 配置生成

- 从 Kraken ClusterConfig 自动生成 kubeadm 所需的 YAML 配置
- 支持生成：InitConfiguration / JoinConfiguration / ClusterConfiguration / KubeletConfiguration

#### FR-2.3 命令构建

- 提供 `BuildPhaseCommand(action, phaseName, extraArgs...)` 方法
- 自动拼接 `--config` 参数
- 示例：`BuildPhaseCommand("init", "preflight")` → `["kubeadm", "init", "phase", "preflight", "--config=/tmp/kubeadm.yaml"]`

### 3.3 kubeadm init 阶段

共 12 个阶段，阶段间依赖关系如下：

```
preflight ──────┬──────────────────────────────────────────────┐
                │                                               │
                ▼                                               │
              certs ──────┬────────────┐                        │
                │         │            │                        │
                ▼         ▼            ▼                        │
          kubeconfig   etcd      control-plane                  │
                │         │            │                        │
                │         │            ▼                        │
                │         │    wait-control-plane               │
                │         │            │                        │
                │         │    ┌───────┘                        │
                │         │    ▼                                │
                │         └── kubelet-start                     │
                │              │                                │
                ├──────────────┤                                │
                │              │                                │
                ▼              ▼                                │
          upload-config  upload-certs                           │
                │              │                                │
                ▼              ▼                                │
          mark-control-plane   │                                │
                │              │                                │
                ▼              │                                │
          bootstrap-token      │                                │
                │              │                                │
                └──────┬───────┘                                │
                       ▼                                        │
                     addon                                      │
```

| 序号 | 阶段名 | 依赖 | kubeadm 命令 |
|---|---|---|---|
| 1 | preflight | 无 | `kubeadm init phase preflight` |
| 2 | certs | preflight | `kubeadm init phase certs all` |
| 3 | kubeconfig | certs | `kubeadm init phase kubeconfig all` |
| 4 | kubelet-start | preflight | `kubeadm init phase kubelet-start` |
| 5 | control-plane | certs | `kubeadm init phase control-plane all` |
| 6 | etcd | certs | `kubeadm init phase etcd local` |
| 7 | wait-control-plane | control-plane, etcd | `kubeadm init phase wait-control-plane` |
| 8 | upload-config | kubelet-start, control-plane | `kubeadm init phase upload-config all` |
| 9 | upload-certs | kubelet-start, control-plane | `kubeadm init phase upload-certs --upload-certs` |
| 10 | mark-control-plane | kubelet-start | `kubeadm init phase mark-control-plane` |
| 11 | bootstrap-token | kubelet-start | `kubeadm init phase bootstrap-token` |
| 12 | addon | upload-config, upload-certs | `kubeadm init phase addon all` |

### 3.4 kubeadm join 阶段

| 序号 | 阶段名 | 依赖 | kubeadm 命令 | 适用角色 |
|---|---|---|---|---|
| 1 | preflight | 无 | `kubeadm join phase preflight` | 全部 |
| 2 | control-plane-prepare | preflight | `kubeadm join phase control-plane-prepare all` | control-plane |
| 3 | kubelet-start | preflight | `kubeadm join phase kubelet-start` | 全部 |
| 4 | control-plane-join | control-plane-prepare, kubelet-start | `kubeadm join phase control-plane-join all` | control-plane |
| 5 | kubeconfig | control-plane-prepare | `kubeadm join phase kubeconfig` | control-plane |

### 3.5 执行策略

- **本地执行** vs **SSH 远程执行**：通过 `ExecuteFunc` 注入，单节点本地执行，多节点 SSH 执行
- **Dry-run 模式**：不实际执行命令，只打印将要执行的命令及依赖关系
- **单阶段执行**：自动解析并执行目标阶段的所有前置依赖
- **错误处理**：任一阶段失败则终止，保证集群状态一致性

### 3.6 SSH 执行层（FR-4）

#### FR-4.1 设计目标

桥接 Phase 层和已有的 SSH 基础设施，使 PhaseRegistry 无需感知执行方式。

#### FR-4.2 复用策略

| 现有组件 | 路径 | 复用方式 |
|---|---|---|
| `executor.Host` | `internal/basic/executor/host.go` | SSH 目标主机（Address/User/Port） |
| `executor.SSHOptions` | `internal/basic/executor/ssh.go` | 连接参数（私钥/密码/超时） |
| `executor.RunShell` | `internal/basic/executor/shell.go` | 远程执行单条命令 → stdout/stderr/exitCode |
| `executor.Executor.Run` | `internal/basic/executor/executor.go` | 并发编排，连接池 + 信号量控制 |
| `runner.Run` | `internal/basic/runner/runner.go` | 从 Config 构建批量执行 |

#### FR-4.3 LocalExecutor（本地执行）

用于单节点本地调试，直接在本地用 `os/exec` 执行 kubeadm 命令。

```go
type LocalExecutor struct{}

func (e *LocalExecutor) ExecuteFunc() kubeadm.ExecuteFunc
```

#### FR-4.4 RemoteExecutor（SSH 远程执行）

核心逻辑：将 `Phase.Command()` 返回的命令切片拼接为 shell 字符串，通过 runner.Run 批量推送到目标节点执行。

```go
type RemoteExecutor struct {
    runnerCfg runner.Config
}

// 从 ClusterConfig.Nodes 构建 SSH 目标列表
func NewRemoteExecutor(clusterCfg *config.ClusterConfig) (*RemoteExecutor, error)

// 返回 kubeadm.ExecuteFunc，适配 PhaseRegistry.RunAll/RunOne
func (e *RemoteExecutor) ExecuteFunc() kubeadm.ExecuteFunc
```

**执行流程**：

```
PhaseRegistry.RunAll(ctx, cfg, executor.ExecuteFunc())
    │
    ▼
RemoteExecutor.ExecuteFunc()
    │
    ├─ phase.Command(cfg)           ← 构建 kubeadm 命令切片
    ├─ strings.Join(cmd, " ")       ← 转成 shell 字符串
    │
    ▼
runner.Run(ctx, runnerCfg, operation)
    │
    ├─ executor.Run(ctx, hosts, operation)
    │      │
    │      ├─ SSH 连接池 + 并发控制
    │      ├─ 对每台 host: RunShell(ctx, host, client, cmdStr)
    │      │      │
    │      │      ├─ session.Run(cmd)
    │      │      └─ Result{Stdout, Stderr, ExitCode, Duration}
    │      │
    │      └─ []Result（每台主机一个）
    │
    └─ 聚合结果：遇 Error 则返回 phase 失败
```

#### FR-4.5 核心适配逻辑

```go
func (e *RemoteExecutor) ExecuteFunc() kubeadm.ExecuteFunc {
    return func(ctx context.Context, phase kubeadm.Phase, cfg kubeadm.PhaseConfig) error {
        // 1. 构建 shell 命令
        args := phase.Command(cfg)
        cmdStr := strings.Join(args, " ")

        // 2. 定义 SSH operation
        operation := func(ctx context.Context, host executor.Host,
                         client *ssh.Client) executor.Result {
            return executor.RunShell(ctx, host, client, cmdStr)
        }

        // 3. 批量执行
        results := runner.Run(ctx, e.runnerCfg, operation)

        // 4. 检查结果
        for _, r := range results {
            if r.Error != nil {
                return fmt.Errorf("host %s: phase %q failed: %w",
                    r.Host.Address, phase.Name(), r.Error)
            }
        }
        return nil
    }
}
```

### 3.7 多平面部署（FR-5）

#### FR-5.1 核心概念

多平面部署解决多角色节点组的编排问题。核心概念：

- **Plane**：一组相同角色的节点 + 它们要执行的 Phase 链
- **平面间串行**：按依赖顺序逐个执行平面（如先 init 第一个控制平面，再 join 其他节点）
- **平面内并发**：同一平面的多个节点并发执行（受 runner.Concurrency 控制）

#### FR-5.2 Plane 定义

```go
type Plane struct {
    Name         string                // 平面名称: "cp-1", "workers"
    Role         string                // 角色: "control-plane" | "worker"
    Nodes        []config.ClusterNode  // 该平面的节点列表
    Dependencies []string              // 依赖的前置平面名
    Concurrency  int                   // 平面内节点并发数
    RegisterPhases func(*PhaseRegistry) // 该平面要注册的 phases
}

type ClusterDeployer struct {
    Planes   []Plane
    PhaseCfg kubeadm.PhaseConfig
    Exec     kubeadm.ExecuteFunc  // 注入 SSH 或本地执行器
}
```

#### FR-5.3 执行策略

```
Step 1: Plane-1 (第一个控制平面)
  └─ 串行执行 init phase 链（preflight → certs → ... → addon）
  └─ 执行完毕后集群已就绪

Step 2: Plane-2 (其余控制平面，依赖 cp-1)
  └─ 并发对该平面的所有节点执行 join control-plane phases
  └─ concurrency = len(nodes) 或用户指定

Step 3: Plane-3 (Worker 组，依赖 cp-1)
  └─ 并发对所有 worker 节点执行 join phases
  └─ concurrency = len(nodes) 或用户指定
```

#### FR-5.4 平面间拓扑排序

平面间也可能存在依赖，使用 Kahn 算法对平面进行排序：

```
工 -> cp-others -> workers
```

```go
func (d *ClusterDeployer) Deploy(ctx context.Context) error {
    // 1. 拓扑排序平面
    sorted, err := topoSortPlanes(d.Planes)
    if err != nil {
        return fmt.Errorf("plane dependency cycle: %w", err)
    }

    // 2. 按序执行每个平面
    for _, plane := range sorted {
        registry := kubeadm.NewPhaseRegistry()
        plane.RegisterPhases(registry)

        // 平面内节点并发执行
        for _, node := range plane.Nodes {
            cfg := d.PhaseCfg
            cfg.NodeAddress = node.Address
            if err := registry.RunAll(ctx, cfg, d.Exec); err != nil {
                return fmt.Errorf("plane %q node %s: %w",
                    plane.Name, node.Address, err)
            }
        }
    }
    return nil
}
```

#### FR-5.5 多平面配置示例

```yaml
cluster:
  planes:
    - name: cp-1
      role: control-plane
      nodes:
        - address: 10.0.0.1
      # phases: init（全量）

    - name: cp-others
      role: control-plane
      depends_on: [cp-1]          # 等 cp-1 完成
      nodes:
        - address: 10.0.0.2
        - address: 10.0.0.3       # 两个节点并发 join
      concurrency: 2

    - name: workers
      role: worker
      depends_on: [cp-1]          # 等 cp-1 完成即可
      nodes:
        - address: 10.0.0.10
        - address: 10.0.0.11
      concurrency: 10
```

#### FR-5.6 命令扩展

```bash
kraken cluster deploy                    # 全量多平面部署
kraken cluster deploy --plane workers    # 只部署某个平面
kraken cluster deploy --dry-run          # 预览部署计划
```

---

## 4. 架构设计

### 4.1 分层架构

```
┌──────────────────────────────────────────────────────────────────┐
│                      CLI Layer  (cmd/cluster/)                    │
│                                                                    │
│  kraken cluster init      kraken cluster join                     │
│  kraken cluster reset     kraken cluster upgrade                  │
│  kraken cluster phase list/run                                    │
├──────────────────────────────────────────────────────────────────┤
│                    Phase Abstraction Layer                        │
│                   (internal/cluster/kubeadm/)                     │
│                                                                    │
│  Phase Interface  →  PhaseRegistry  →  Ordered Execution Engine   │
│                                                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │ Init     │  │ Join     │  │ Reset    │  │ Upgrade  │         │
│  │ Phases   │  │ Phases   │  │ Phases   │  │ Phases   │         │
│  │ (12个)   │  │ (5个)    │  │ (未来)    │  │ (未来)    │         │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘         │
├──────────────────────────────────────────────────────────────────┤
│                   kubeadm Command Wrapper                         │
│                                                                    │
│  kubeadm binary discovery  →  config generation  →  cmd exec      │
│                   Executor Layer (桥接层)                          │
│                                                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐       │
│  │  LocalExec   │  │ RemoteExec   │  │ ClusterDeployer  │       │
│  │  (os/exec)   │  │ (SSH runner) │  │ (多平面编排)      │       │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘       │
│         │                 │                    │                  │
│         │    ┌────────────┘                    │                  │
│         ▼    ▼                                 ▼                  │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │            Infrastructure Layer (复用 existing)           │     │
│  │                                                          │     │
│  │  executor.RunShell  │  executor.Executor.Run  │  runner  │     │
│  │  Config Loader      │  Logger                │  Printer  │     │
│  └─────────────────────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────────────────┘
```

### 4.2 现有可复用组件清单

| 组件 | 路径 | 复用方式 |
|---|---|---|
| SSH Executor | `internal/basic/executor/` | 远程执行 kubeadm 命令 |
| Batch Runner | `internal/basic/runner/` | 多节点批量编排 |
| RunShell | `internal/basic/executor/shell.go` | 捕获 stdout/stderr/exitCode |
| Config Loader | `internal/config/config.go` | YAML 加载、Context 注入 |
| Logger | `pkg/logger/` | slog 结构化日志 |
| Result Printer | `utils/print_result.go` | 表格化打印执行结果 |
| CLI Groups | `pkg/cli/groups.go` | `GroupCluster` 已定义 |

---

## 5. 接口设计

### 5.1 Phase 接口

```go
// 文件: internal/cluster/kubeadm/phase.go

type Phase interface {
    Name() string
    Description() string
    Command(cfg PhaseConfig) []string
    Dependencies() []string
}

type PhaseConfig struct {
    KubeadmConfigPath string
    KubeadmBinary     string
    NodeAddress       string
    DryRun            bool
    ExtraArgs         map[string]string
}
```

### 5.2 PhaseRegistry

```go
type PhaseRegistry struct {
    phases map[string]Phase
    order  []string
}

func NewPhaseRegistry() *PhaseRegistry
func (r *PhaseRegistry) Register(p Phase)
func (r *PhaseRegistry) List() []Phase
func (r *PhaseRegistry) Get(name string) Phase
func (r *PhaseRegistry) Sorted() ([]Phase, error)         // 拓扑排序
func (r *PhaseRegistry) RunAll(ctx, cfg, execFunc) error   // 全量执行
func (r *PhaseRegistry) RunOne(ctx, name, cfg, execFunc) error // 单阶段+依赖执行

type ExecuteFunc func(ctx context.Context, phase Phase, cfg PhaseConfig) error
```

### 5.3 Kubeadm Wrapper

```go
// 文件: internal/cluster/kubeadm/kubeadm.go

type Kubeadm struct {
    binaryPath string
    configPath string
}

func NewKubeadm(cfg KubeadmConfig) (*Kubeadm, error)
func (k *Kubeadm) Check(ctx context.Context) error
func (k *Kubeadm) GenerateConfig(clusterCfg ClusterConfig) error
func (k *Kubeadm) BuildPhaseCommand(action, phaseName string, extraArgs ...string) []string
func (k *Kubeadm) RunCommand(ctx context.Context, args ...string) (stdout, stderr string, err error)
```

### 5.4 Init Phase 实现范例

```go
// 文件: internal/cluster/kubeadm/init.go

type PreflightPhase struct{}
func (p *PreflightPhase) Name() string        { return "preflight" }
func (p *PreflightPhase) Description() string { return "Run pre-flight checks" }
func (p *PreflightPhase) Dependencies() []string { return nil }
func (p *PreflightPhase) Command(cfg PhaseConfig) []string {
    return []string{cfg.KubeadmBinary, "init", "phase", "preflight", "--config", cfg.KubeadmConfigPath}
}

type CertsPhase struct{}
func (p *CertsPhase) Name() string        { return "certs" }
func (p *CertsPhase) Description() string { return "Generate all certificates" }
func (p *CertsPhase) Dependencies() []string { return []string{"preflight"} }
func (p *CertsPhase) Command(cfg PhaseConfig) []string {
    return []string{cfg.KubeadmBinary, "init", "phase", "certs", "all", "--config", cfg.KubeadmConfigPath}
}

func RegisterInitPhases(r *PhaseRegistry) {
    r.Register(&PreflightPhase{})
    r.Register(&CertsPhase{})
    // ... 共 12 个
}
```

---

## 6. 配置设计

### 6.1 现有 ClusterConfig（需补全）

当前 [config.go](../../config/config.go) 中 `ClusterConfig` 为空壳：

```go
type ClusterConfig struct {
    //TODO: 实现具体配置
}
```

### 6.2 目标 ClusterConfig

```go
type ClusterConfig struct {
    Name              string                  `yaml:"name"`
    KubernetesVersion string                  `yaml:"kubernetes_version"`
    PodCIDR           string                  `yaml:"pod_cidr"`
    ServiceCIDR       string                  `yaml:"service_cidr"`
    ServiceDNSDomain  string                  `yaml:"service_dns_domain"`
    ControlPlane      ClusterControlPlane     `yaml:"control_plane"`
    ContainerRuntime  ClusterContainerRuntime `yaml:"container_runtime"`
    Nodes             []ClusterNode           `yaml:"nodes"`
    ExtraKubeadm      map[string]interface{}  `yaml:"extra_kubeadm_config"`
}

type ClusterControlPlane struct {
    Endpoint       string `yaml:"endpoint"`
    CertificateKey string `yaml:"certificate_key"`
}

type ClusterContainerRuntime struct {
    Type   string `yaml:"type"`
    Socket string `yaml:"socket"`
}

type ClusterNode struct {
    Address string `yaml:"address"`
    Role    string `yaml:"role"`
    User    string `yaml:"user"`
    Port    int    `yaml:"port"`
}
```

### 6.3 配置示例

```yaml
cluster:
  name: my-cluster
  kubernetes_version: 1.30.0

  pod_cidr: 10.244.0.0/16
  service_cidr: 10.96.0.0/12
  service_dns_domain: cluster.local

  control_plane:
    endpoint: 10.0.0.1:6443

  container_runtime:
    type: containerd
    socket: /run/containerd/containerd.sock

  nodes:
    - address: 10.0.0.1
      role: control-plane
      user: root
    - address: 10.0.0.2
      role: worker
      user: root

  extra_kubeadm_config:
    apiServer:
      extraArgs:
        audit-log-path: /var/log/kubernetes/audit.log
```

### 6.4 配置生成映射

```
Kraken cluster.yaml                          kubeadm InitConfiguration
┌──────────────────────┐                    ┌──────────────────────────────┐
│ cluster:             │     Generate       │ apiVersion: kubeadm.k8s.io/   │
│   name: my-cluster   │  ──────────────→   │ kind: InitConfiguration       │
│   version: 1.30.0    │                    │ localAPIEndpoint:             │
│   pod_cidr: 10.244/16│                    │   advertiseAddress: 10.0.0.1  │
│   svc_cidr: 10.96/16 │                    │ nodeRegistration:             │
│   control_plane:     │                    │   name: cp-1                  │
│     endpoint: ...    │                    │ networking:                   │
│   ...                │                    │   podSubnet: 10.244.0.0/16   │
└──────────────────────┘                    └──────────────────────────────┘
```

---

## 7. 命令设计

### 7.1 命令树

```
kraken
├── bc                              ← 已有
│   ├── cmd
│   ├── script
│   └── adduser
├── cluster                         ← 新增
│   ├── init                        ← kubeadm init（全阶段）
│   │   --config                    ← Kraken 配置文件
│   │   --dry-run                   ← 预览模式
│   │   --phase <name>              ← 指定执行单个阶段
│   │
│   ├── join                        ← kubeadm join（全阶段）
│   │   --config
│   │   --role worker|control-plane
│   │   --dry-run
│   │   --phase <name>
│   │
│   ├── reset                       ← kubeadm reset（后续版本）
│   │
│   ├── upgrade                     ← kubeadm upgrade（后续版本）
│   │
│   └── phase                       ← 阶段管理子命令
│       ├── list                    ← 列出所有阶段及依赖
│       │   --action init|join      ← 按操作类型筛选
│       │
│       └── run                     ← 运行指定阶段
│           --action init|join
│           --phase <name>
│           --config
│           --dry-run
│
├── version                         ← 已有
└── completion                      ← 已有
```

### 7.2 CLI 交互示例

**init 全量执行：**
```bash
$ kraken cluster init --config cluster.yaml
[14:03:01] INF: phase "preflight" started
[14:03:02] INF: phase "preflight" completed ✓
[14:03:02] INF: phase "certs" started
[14:03:05] INF: phase "certs" completed ✓
...
[14:05:30] INF: cluster initialization completed
  kubeconfig saved to: /etc/kubernetes/admin.conf
```

**dry-run 预览：**
```bash
$ kraken cluster init --config cluster.yaml --dry-run
Phase "preflight":      kubeadm init phase preflight --config=/tmp/kubeadm.yaml
  ├── depends on: (none)
Phase "certs":          kubeadm init phase certs all --config=/tmp/kubeadm.yaml
  ├── depends on: preflight
Phase "control-plane":  kubeadm init phase control-plane all --config=/tmp/kubeadm.yaml
  ├── depends on: certs
...
```

**单阶段执行：**
```bash
$ kraken cluster init --config cluster.yaml --phase certs
[14:10:01] INF: resolving dependencies for phase "certs"
[14:10:01] INF: phase "preflight" started (dependency)
[14:10:02] INF: phase "preflight" completed ✓
[14:10:02] INF: phase "certs" started
[14:10:05] INF: phase "certs" completed ✓
```

**阶段列表：**
```bash
$ kraken cluster phase list --action init
Phase                  Dependencies          Description
──────────────────────────────────────────────────────────
preflight              -                     Run pre-flight checks
certs                  preflight             Generate all certificates
kubeconfig             certs                 Generate all kubeconfig files
kubelet-start          preflight             Write kubelet settings
control-plane          certs                 Generate static Pod manifests
etcd                   certs                 Generate etcd static Pod manifest
wait-control-plane     control-plane, etcd   Wait for control plane
upload-config          kubelet-start, ...    Upload kubeadm config
upload-certs           kubelet-start, ...    Upload certificates
mark-control-plane     kubelet-start         Mark node as control-plane
bootstrap-token        kubelet-start         Generate bootstrap tokens
addon                  upload-config, ...    Install addons
```

---

## 8. 执行流程

### 8.1 完整 init 时序

```
CLI 层                  Phase 层                kubeadm Wrapper         SSH Executor
  │                        │                        │                      │
  │  kraken cluster init   │                        │                      │
  │───────────────────────→│                        │                      │
  │                        │                        │                      │
  │                        │  1. LoadConfig()       │                      │
  │                        │───────────────────────→│                      │
  │                        │                        │                      │
  │                        │  2. GenerateConfig()   │                      │
  │                        │───────────────────────→│                      │
  │                        │                        │                      │
  │                        │  3. NewKubeadm()       │                      │
  │                        │───────────────────────→│                      │
  │                        │                        │                      │
  │                        │  4. RegisterInitPhases()                       │
  │                        │  (12个 Phase 注册)                             │
  │                        │                        │                      │
  │                        │  5. registry.Sorted()  │                      │
  │                        │  (拓扑排序→执行顺序)     │                      │
  │                        │                        │                      │
  │                        │  6. For each Phase:    │                      │
  │                        │     phase.Command()    │                      │
  │                        │───────────────────────→│                      │
  │                        │                        │                      │
  │                        │     execute(cmd)       │                      │
  │                        │────────────────────────┼─────────────────────→│
  │                        │                        │   SSH → kubeadm init  │
  │                        │                        │   phase preflight     │
  │                        │                        │←─────────────────────│
  │                        │     result             │                      │
  │                        │←───────────────────────│                      │
  │                        │                        │                      │
  │                        │  (失败终止 / 成功继续)   │                      │
  │                        │                        │                      │
  │  print results         │                        │                      │
  │←───────────────────────│                        │                      │
```

### 8.2 执行策略决策树

```
kraken cluster init --config cluster.yaml
        │
        ▼
  ┌─ 是否指定 --phase? ──────────────────────────┐
  │ YES                                           │ NO
  ▼                                               ▼
RunOne(name)                                 RunAll()
  │                                               │
  ├─ 解析依赖链                                   ├─ 拓扑排序
  ├─ 按依赖顺序执行                               ├─ 遍历执行
  └─ 目标阶段最后执行                             └─ 任一失败 → 终止 + 报错
```

---

## 9. 文件结构

### 9.1 新增文件

```
kraken-toolkit/
│
├── cmd/
│   ├── root.go                          ← [修改] 注册 cluster 命令组
│   └── cluster/
│       ├── cluster.go                   ← [新增] kraken cluster 根命令
│       ├── init.go                      ← [新增] kraken cluster init
│       ├── join.go                      ← [新增] kraken cluster join
│       ├── deploy.go                    ← [新增] kraken cluster deploy（多平面）
│       ├── reset.go                     ← [新增] kraken cluster reset（后续）
│       └── phase.go                     ← [新增] kraken cluster phase list/run
│
├── internal/
│   ├── deploy/
│   │   └── cluster/
│   │       ├── kubeadm/
│   │       │   ├── phase.go             ← [新增] Phase 接口 + PhaseRegistry
│   │       │   ├── phase_test.go        ← [新增] 注册器单元测试
│   │       │   ├── kubeadm.go           ← [新增] kubeadm 二进制封装
│   │       │   ├── kubeadm_test.go      ← [新增] kubeadm 命令构建测试
│   │       │   ├── init.go              ← [新增] init 12个 Phase 实现
│   │       │   ├── init_test.go         ← [新增] init phase 单元测试
│   │       │   ├── join.go              ← [新增] join 5个 Phase 实现
│   │       │   ├── join_test.go         ← [新增] join phase 单元测试
│   │       │   └── config.go            ← [新增] KrakenConfig → kubeadm YAML 生成
│   │       └── executor/
│   │           ├── remote.go            ← [新增] SSH 远程执行器（桥接 basic/executor）
│   │           ├── local.go             ← [新增] 本地执行器（os/exec 调试用）
│   │           ├── deployer.go          ← [新增] 多平面部署编排器
│   │           └── deployer_test.go     ← [新增] 多平面部署单元测试
│   │
│   └── config/
│       └── config.go                    ← [修改] ClusterConfig 补全
│
└── pkg/
    └── cli/
        └── groups.go                    ← [修改] 新增 cluster help 模板
```

### 9.2 文件依赖关系

```
cmd/cluster/cluster.go
    ├── cmd/cluster/init.go ──→ internal/cluster/kubeadm/init.go ──→ phase.go
    ├── cmd/cluster/join.go ──→ internal/cluster/kubeadm/join.go ──→ phase.go
    └── cmd/cluster/phase/phase.go ──→ internal/cluster/kubeadm/phase.go

internal/deploy/cluster/kubeadm/kubeadm.go
    ├── internal/deploy/cluster/kubeadm/config.go ──→ internal/config/config.go
    └── internal/deploy/cluster/executor/remote.go ──→ internal/basic/executor/
                                                      internal/basic/runner/

internal/deploy/cluster/executor/
    ├── local.go ──→ kubeadm/phase.go
    ├── remote.go ──→ kubeadm/phase.go, basic/executor/, basic/runner/
    └── deployer.go ──→ kubeadm/phase.go, executor/remote.go
```

---

## 10. 开发计划

### Phase 1：核心框架（优先级 P0）

| 任务 | 文件 | 预计工作量 |
|---|---|---|
| Phase 接口 + PhaseRegistry | `internal/deploy/cluster/kubeadm/phase.go` | 1d |
| PhaseRegistry 单元测试 | `internal/deploy/cluster/kubeadm/phase_test.go` | 0.5d |
| kubeadm 二进制封装 | `internal/deploy/cluster/kubeadm/kubeadm.go` | 1d |
| 配置生成器 | `internal/deploy/cluster/kubeadm/config.go` | 1d |
| ClusterConfig 补全 | `internal/config/config.go` | 0.5d |

### Phase 2：init 实现（优先级 P0）

| 任务 | 文件 | 预计工作量 |
|---|---|---|
| 12 个 Init Phase 实现 | `internal/deploy/cluster/kubeadm/init.go` | 1d |
| Init Phase 单元测试 | `internal/deploy/cluster/kubeadm/init_test.go` | 0.5d |

### Phase 3：join 实现（优先级 P1）

| 任务 | 文件 | 预计工作量 |
|---|---|---|
| 5 个 Join Phase 实现 | `internal/deploy/cluster/kubeadm/join.go` | 1d |
| Join Phase 单元测试 | `internal/deploy/cluster/kubeadm/join_test.go` | 0.5d |

### Phase 4：SSH 执行层（优先级 P1）

| 任务 | 文件 | 预计工作量 |
|---|---|---|
| LocalExecutor 本地执行器 | `internal/deploy/cluster/executor/local.go` | 0.5d |
| RemoteExecutor SSH 执行器 | `internal/deploy/cluster/executor/remote.go` | 1d |
| ClusterDeployer 多平面编排 | `internal/deploy/cluster/executor/deployer.go` | 1d |
| 多平面部署单元测试 | `internal/deploy/cluster/executor/deployer_test.go` | 0.5d |

### Phase 5：CLI 命令层（优先级 P1）

| 任务 | 文件 | 预计工作量 |
|---|---|---|
| cluster 根命令 | `cmd/cluster/cluster.go` | 0.5d |
| cluster init 命令 | `cmd/cluster/init.go` | 0.5d |
| cluster join 命令 | `cmd/cluster/join.go` | 0.5d |
| cluster deploy 命令（多平面） | `cmd/cluster/deploy.go` | 0.5d |
| phase list/run 命令 | `cmd/cluster/phase.go` | 1d |
| 注册到 root | `cmd/root.go` | 0.5d |

### Phase 6：后续扩展（优先级 P2）

| 任务 | 说明 |
|---|---|
| cluster reset | kubeadm reset 阶段化 |
| cluster upgrade | kubeadm upgrade 阶段化 |
| 集成测试 | 端到端集群创建/加入测试 |
| kubeadm_test.go | kubeadm 二进制封装单元测试 |

---

## 11. 扩展性范例

以添加 `kubeadm reset` 为例，展示三步扩展法：

### Step 1：实现 Phase

```go
// internal/cluster/kubeadm/reset.go

type ResetPreflightPhase struct{}
func (p *ResetPreflightPhase) Name() string        { return "reset-preflight" }
func (p *ResetPreflightPhase) Description() string { return "Run reset pre-flight checks" }
func (p *ResetPreflightPhase) Dependencies() []string { return nil }
func (p *ResetPreflightPhase) Command(cfg PhaseConfig) []string {
    return []string{cfg.KubeadmBinary, "reset", "phase", "preflight"}
}

func RegisterResetPhases(r *PhaseRegistry) {
    r.Register(&ResetPreflightPhase{})
    r.Register(&ResetRemoveEtcdPhase{})
    // ...
}
```

### Step 2：添加命令

```go
// cmd/cluster/reset.go

func NewResetCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "reset",
        Short: "Reset a Kubernetes cluster (kubeadm reset)",
        RunE: func(cmd *cobra.Command, args []string) error {
            registry := kubeadm.NewPhaseRegistry()
            kubeadm.RegisterResetPhases(registry)
            return registry.RunAll(cmd.Context(), phaseCfg, executeRemote)
        },
    }
    return cmd
}
```

### Step 3：注册到父命令

```go
// cmd/cluster/cluster.go
cmd.AddCommand(NewInitCmd(), NewJoinCmd(), NewResetCmd())
```

---

## 12. 设计决策记录

| 决策点 | 选择 | 理由 |
|---|---|---|
| Phase vs 直接调 kubeadm | Phase 接口 | 可编排、可观测、可测试、可扩展 |
| 本地 vs SSH 执行 | 可切换 | 通过 ExecuteFunc 注入，单节点本地，多节点 SSH |
| SSH 桥接方式 | 新建 executor 层，不修改 Phase 层 | Phase 层只关心命令构建，不关心执行方式 |
| 复用 SSH 基础设施 | 复用 basic/executor + basic/runner | 不重新发明 SSH 连接管理、并发控制、错误处理 |
| 命令切片 → shell 字符串 | strings.Join | RunShell 接受字符串，非切片参数 |
| 多节点编排 | ClusterDeployer + Plane 概念 | 按角色分组、平面间串行、平面内并发 |
| 平面间排序 | Kahn 算法（复用 Phase 层模式） | 与 PhaseRegistry 一致的 DAG 排序经验 |
| 配置文件 | 单一 cluster.yaml | 与现有 config 模型一致，PersistentPreRun 注入 Context |
| 错误处理 | 阶段失败 → 终止 | 保证集群状态一致，不支持部分执行（除非显式 --phase） |
| kubeadm 配置 | 代码生成 | 从 Kraken ClusterConfig 自动生成，避免用户手写两份配置 |
| 命令分组 | GroupCluster | 已在 groups.go 中预定义 |
| 拓扑排序 | Kahn 算法 | 标准 DAG 排序算法，可检测循环依赖 |

---

> 下一步：确认本需求文档后，按 Phase 1 → Phase 2 → Phase 3 → Phase 4 顺序进入开发。