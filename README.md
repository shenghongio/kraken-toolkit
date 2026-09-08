# Kraken Pedestal

<p align="center">
  <strong>Kraken Pedestal</strong>
</p>

<p align="center">
  Kubernetes & AI Infrastructure Operations Toolkit
</p>

<p align="center">
  A Go-based toolkit for system operations, Kubernetes cluster management,
  Helm deployment, and GPU resource scheduling.
</p>

<p align="center">
  <a href="https://github.com/KrakenStack/kraken-pedestal">GitHub</a>
  · 
  <a href="https://github.com/KrakenStack/kraken-pedestal/issues">Issues</a>
</p>


<p align="center">
  <img src="https://cdn.simpleicons.org/kubernetes" width="36" alt="Kubernetes">
  <img src="https://cdn.simpleicons.org/go" width="36" alt="Go">
  <img src="https://cdn.simpleicons.org/helm" width="36" alt="Helm">
  <img src="https://cdn.simpleicons.org/linux" width="36" alt="Linux">
  <img src="https://cdn.simpleicons.org/openai" width="36" alt="AI">
  <img src="https://cdn.simpleicons.org/nvidia" width="36" alt="NVIDIA">
</p>

<h1 align="center">Kraken Pedestal</h1>

<p align="center">
  <strong>Kubernetes & AI Infrastructure Operations Toolkit</strong>
</p>

<p align="center">
  System Operations · Kubernetes · Helm · GPU · HAMi · DRA
</p>

<p align="center">
  <a href="https://github.com/KrakenStack/kraken-pedestal">
    <img src="https://img.shields.io/github/stars/KrakenStack/kraken-pedestal?style=flat-square&logo=github">
  </a>
  <a href="https://github.com/KrakenStack/kraken-pedestal">
    <img src="https://img.shields.io/github/license/KrakenStack/kraken-pedestal?style=flat-square">
  </a>
  <a href="https://github.com/KrakenStack/kraken-pedestal">
    <img src="https://img.shields.io/github/last-commit/KrakenStack/kraken-pedestal?style=flat-square">
  </a>
</p>

---

## Overview

**Kraken Pedestal** is a Go-based infrastructure operations toolkit for managing and operating **Linux systems, Kubernetes clusters, cloud-native applications, and GPU workloads**.

The project brings commonly used infrastructure operations into a unified CLI and provides reusable components for higher-level automation.

It currently focuses on the following areas:

* **Basic** — system information and basic host operations
* **Ops Tools** — infrastructure and system diagnostic utilities
* **Kubernetes Cluster** — Kubernetes cluster lifecycle and management
* **Helm Deploy** — Helm-based application deployment and management
* **GPU Scheduling** — GPU resource management based on HAMi and Kubernetes DRA

The long-term goal is to provide a unified foundation for infrastructure engineers to **inspect, provision, configure, deploy, and operate Kubernetes-based infrastructure**.

---

## Why Kraken Pedestal?

Infrastructure operations often require many independent tools:

```text
Linux
 ├── systemctl
 ├── ip
 ├── mount
 ├── lsof
 ├── dmidecode
 └── ...

Kubernetes
 ├── kubectl
 ├── kubeadm
 ├── kubelet
 └── ...

Helm
 └── helm

GPU
 ├── nvidia-smi
 ├── HAMi
 └── DRA
```

Each tool has its own command interface, configuration model, output format, and error handling.

Kraken Pedestal attempts to provide a unified operational layer:

```text
                    Kraken
                      │
          ┌───────────┴───────────┐
          │                       │
      CLI Interface          Common Runtime
          │                       │
          │          ┌────────────┼────────────┐
          │          │            │            │
          ▼          ▼            ▼            ▼
       Basic      Config       Logger       Error
          │
          ├──────── Ops Tools
          │
          ├──────── Kubernetes
          │
          ├──────── Helm
          │
          └──────── GPU
                       │
                  ┌────┴────┐
                  │         │
                 HAMi      DRA
```

The project is therefore not intended to replace every existing infrastructure tool.

Instead, it provides a **common operational interface and reusable implementation layer** around infrastructure workflows.

---

# Features

## Basic

The `basic` module provides fundamental host and operating-system inspection capabilities.

Typical functionality includes:

* Host information
* Operating system information
* Kernel information
* CPU architecture
* Machine ID
* Boot ID
* Hostname
* Filesystem information
* Runtime environment information

Example:

```bash
kraken basic system-info
```

The collected information is represented as structured data and can be rendered in different formats.

```bash
kraken basic system-info --output table
```

```bash
kraken basic system-info --output json
```

```bash
kraken basic system-info --output yaml
```

The purpose of the `basic` module is to provide the foundation for higher-level system diagnostics.

---

# Ops Tools

The `ops-tools` area contains commonly required infrastructure and system operation capabilities.

The goal is to provide frequently used operational functions through a consistent CLI rather than requiring engineers to remember a large collection of independent commands.

Typical areas include:

```text
Ops Tools
│
├── System
│   ├── CPU
│   ├── Memory
│   ├── Disk
│   ├── Filesystem
│   └── Process
│
├── Network
│   ├── Interface
│   ├── Route
│   ├── DNS
│   └── Connectivity
│
├── Storage
│   ├── Block Device
│   ├── Mount
│   └── Filesystem
│
└── Diagnostics
    ├── Process
    ├── IO
    └── System Health
```

This area is intentionally modular so that new operational tools can be added without changing the overall CLI architecture.

---

# Kubernetes Cluster

Kubernetes cluster management is one of the core capabilities of Kraken Pedestal.

The Kubernetes module is intended to provide a unified interface for common cluster lifecycle operations.

Conceptually:

```text
Kubernetes
│
├── Cluster
│   ├── Create
│   ├── Configure
│   ├── Join
│   ├── Upgrade
│   └── Delete
│
├── Node
│   ├── Add
│   ├── Remove
│   ├── Drain
│   └── Inspect
│
├── Component
│   ├── Control Plane
│   ├── kubelet
│   ├── CNI
│   └── Runtime
│
└── Diagnostics
    ├── Cluster
    ├── Node
    └── Workload
```

The project is intended to work with common Kubernetes components and infrastructure tooling rather than implementing an alternative Kubernetes distribution.

For example:

```text
Kraken
   │
   ├── kubeadm
   ├── containerd
   ├── kubectl
   ├── CNI
   └── Kubernetes API
```

Kraken acts as an orchestration and operational layer around these components.

---

# Helm Deployment

Kraken Pedestal provides Helm-oriented deployment capabilities for Kubernetes applications.

The Helm module is designed around the application lifecycle:

```text
Chart
  │
  ▼
Repository / Local Chart
  │
  ▼
Values
  │
  ▼
Release
  │
  ├── Install
  ├── Upgrade
  ├── Rollback
  ├── Status
  └── Uninstall
```

Typical operations include:

```bash
kraken helm deploy
```

```bash
kraken helm upgrade
```

```bash
kraken helm status
```

```bash
kraken helm rollback
```

```bash
kraken helm uninstall
```

The exact command structure may evolve as the Helm module develops.

The design goal is to make application deployment consistent with the rest of the Kraken CLI.

---

# GPU Infrastructure

GPU infrastructure is an important part of the Kraken roadmap.

Modern Kubernetes GPU environments increasingly require more than simple device discovery.

A complete GPU platform may involve:

```text
GPU
│
├── Discovery
│
├── Device Plugin
│
├── Resource Allocation
│
├── Scheduling
│
├── Isolation
│
└── Workload Management
```

Kraken Pedestal currently focuses on integrating GPU scheduling and resource management with Kubernetes.

---

# HAMi

[HAMi](https://github.com/Project-HAMi/HAMi) provides GPU virtualization and scheduling capabilities for Kubernetes environments.

Kraken Pedestal integrates HAMi-related operations into the broader infrastructure workflow.

The intended workflow is:

```text
Kubernetes Cluster
        │
        ▼
     GPU Node
        │
        ▼
       HAMi
        │
        ├── GPU Discovery
        ├── GPU Allocation
        ├── GPU Isolation
        └── GPU Scheduling
        │
        ▼
     Workload
```

This allows GPU-related infrastructure operations to be managed together with the rest of the Kubernetes environment.

---

# Kubernetes DRA

Kraken Pedestal also targets Kubernetes **Dynamic Resource Allocation (DRA)**.

DRA provides a Kubernetes-native mechanism for allocating specialized resources to workloads.

For GPU infrastructure, the conceptual workflow is:

```text
Pod
 │
 ▼
ResourceClaim
 │
 ▼
DRA
 │
 ▼
GPU Resource
 │
 ▼
Node
```

Kraken Pedestal aims to provide operational tooling around this workflow.

The GPU scheduling layer is intended to evolve toward supporting multiple resource allocation mechanisms rather than coupling the entire project to a single GPU implementation.

---

# HAMi vs DRA

HAMi and DRA address GPU resource management from different architectural perspectives.

A simplified view:

```text
                  GPU Workload
                       │
             ┌─────────┴─────────┐
             │                   │
            HAMi                DRA
             │                   │
       GPU virtualization    Resource allocation
             │                   │
             └─────────┬─────────┘
                       │
                    Kubernetes
```

Kraken Pedestal keeps these integrations modular so that GPU infrastructure can evolve independently from the core Kubernetes functionality.

This also allows the project to support different GPU environments without making the entire CLI dependent on a single scheduler implementation.

---

# Architecture

Kraken Pedestal follows a layered architecture.

```text
┌───────────────────────────────────────────────┐
│                    CLI                        │
│                                               │
│  basic / ops / cluster / helm / gpu / ...    │
└───────────────────────┬───────────────────────┘
                        │
┌───────────────────────▼───────────────────────┐
│                 Application                   │
│                                               │
│   Cluster / Deploy / GPU / System Operations │
└───────────────────────┬───────────────────────┘
                        │
┌───────────────────────▼───────────────────────┐
│                  Runtime                      │
│                                               │
│ Configuration / Logger / Clients / Context   │
└───────────────────────┬───────────────────────┘
                        │
        ┌───────────────┼────────────────┐
        │               │                │
        ▼               ▼                ▼
   Kubernetes         Helm             System
       │                │                │
       ▼                ▼                ▼
 Kubernetes API      Helm API        Linux / OS
        │
        ▼
   GPU Infrastructure
      ┌──────┴──────┐
      │             │
     HAMi          DRA
```

The architecture is intentionally divided into several layers.

---

# CLI Layer

The CLI layer is responsible for:

* Command registration
* Command hierarchy
* Flag parsing
* Input validation
* Command lifecycle
* Output selection

The CLI should remain relatively thin.

It should orchestrate application logic rather than contain large amounts of infrastructure implementation.

---

# Runtime Layer

Runtime initialization connects the application's configuration and infrastructure dependencies.

Conceptually:

```text
CLI
 │
 ▼
Configuration
 │
 ▼
Runtime
 │
 ├── Logger
 ├── Kubernetes Client
 ├── Helm Client
 ├── System Runtime
 └── GPU Runtime
```

This allows individual commands to consume initialized dependencies without having to repeatedly implement initialization logic.

---

# Configuration

Kraken Pedestal is designed around a unified configuration model.

Instead of maintaining independent configuration files for every module, the application can use a single configuration structure containing multiple sections.

Example:

```yaml
basic:
  concurrency: 8

kubernetes:
  kubeconfig: ~/.kube/config

helm:
  timeout: 5m

gpu:
  runtime: hami
```

Conceptually:

```text
config.yaml
│
├── basic
│
├── ops
│
├── kubernetes
│
├── helm
│
└── gpu
    ├── hami
    └── dra
```

This provides a consistent configuration model as new modules are introduced.

---

# Logging

Kraken Pedestal uses Go's `log/slog` as the foundation of its logging system.

The project provides a common logging abstraction for all modules.

The logger is designed to support:

* Debug
* Info
* Warn
* Error
* Structured attributes
* Human-readable terminal output
* Colored log levels
* File output
* Log rotation
* Context-aware logging

Example:

```text
[ 26-09-06/13:10:20 ] INF: cluster initialization started
[ 26-09-06/13:10:21 ] DEB: loading cluster configuration
[ 26-09-06/13:10:22 ] INF: kubelet configuration completed
[ 26-09-06/13:10:23 ] ERR: failed to initialize CNI
```

The goal is to provide consistent logging behavior across:

```text
Basic
Ops Tools
Kubernetes
Helm
GPU
```

---

# Error Handling

Infrastructure operations often fail because of external conditions:

* missing system dependencies
* invalid configuration
* unavailable nodes
* Kubernetes API errors
* Helm failures
* network problems
* GPU device problems

Kraken Pedestal therefore uses a common error model to preserve the original error while adding operational context.

Conceptually:

```text
System Error
     │
     ▼
Infrastructure Error
     │
     ▼
Kraken Error
     │
     ▼
CLI
```

Example:

```go
return kerror.Wrap(
    kerror.CodeSystem,
    "failed to collect system information",
    err,
)
```

This allows errors to remain useful for both humans and higher-level callers.

---

# Output

Kraken separates data collection from presentation.

```text
                 Operation
                    │
                    ▼
                Result Model
                    │
          ┌─────────┼─────────┐
          │         │         │
          ▼         ▼         ▼
        Table      JSON      YAML
```

Human-readable output:

```bash
kraken basic system-info
```

Machine-readable output:

```bash
kraken basic system-info --output json
```

This separation is particularly important for infrastructure automation.

---

# Command Structure

The CLI is organized by infrastructure domain.

The current conceptual command tree is:

```text
kraken
│
├── basic
│   └── system-info
│
├── ops
│   ├── system
│   ├── network
│   ├── storage
│   └── diagnostics
│
├── cluster
│   ├── create
│   ├── join
│   ├── node
│   ├── upgrade
│   └── ...
│
├── helm
│   ├── deploy
│   ├── upgrade
│   ├── status
│   ├── rollback
│   └── uninstall
│
└── gpu
    ├── hami
    ├── dra
    └── ...
```

The exact commands may change as the project evolves.

The important principle is that the command hierarchy reflects the infrastructure domain rather than individual implementation details.

---

# Project Structure

The project follows a Go-oriented modular structure.

A conceptual structure is:

```text
.
├── cmd/
│   ├── basic/
│   ├── ops/
│   ├── cluster/
│   ├── helm/
│   └── gpu/
│
├── internal/
│   ├── basic/
│   ├── ops/
│   ├── cluster/
│   ├── helm/
│   └── gpu/
│
├── pkg/
│   ├── logger/
│   ├── error/
│   ├── printer/
│   ├── config/
│   └── ...
│
├── models/
│   └── ...
│
├── configs/
│   └── ...
│
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

The actual structure may evolve as individual modules become more complex.

---

# Supported Infrastructure

Kraken Pedestal is primarily designed for Linux and Kubernetes infrastructure environments.

The project is expected to interact with components such as:

```text
Operating System
      │
      ├── Linux
      │
      ├── systemd
      │
      ├── networking
      │
      └── storage
      │
      ▼
Container Runtime
      │
      └── containerd
      │
      ▼
Kubernetes
      │
      ├── kubeadm
      ├── kubelet
      ├── kubectl
      └── Kubernetes API
      │
      ├── Helm
      │
      └── GPU
           ├── HAMi
           └── DRA
```

The exact supported versions depend on the implementation of each module.

---

# Development

## Requirements

A development environment should provide:

* Go
* Git
* Make
* Linux environment for system-level functionality
* Kubernetes environment for cluster-related functionality
* Helm for Helm-related development
* GPU-enabled Kubernetes environment for GPU functionality

Some functionality can be developed and tested independently.

For example:

```text
Basic / Ops Tools
    │
    └── Linux environment

Cluster
    │
    └── Kubernetes environment

Helm
    │
    └── Kubernetes + Helm

GPU
    │
    └── Kubernetes + GPU + HAMi / DRA
```

---

# Clone

```bash
git clone https://github.com/KrakenStack/kraken-pedestal.git

cd kraken-pedestal
```

---

# Build

Build the project using the provided Makefile:

```bash
make build
```

Or build directly using Go:

```bash
go build ./...
```

The generated binary is placed under the project's build output directory according to the Makefile configuration.

---

# Development Workflow

A typical development workflow is:

```bash
make fmt
make check
make test
make build
```

Individual checks can also be run directly:

```bash
go test ./...
```

```bash
go vet ./...
```

```bash
gofmt -w .
```

---

# Testing

The project uses Go's standard testing framework.

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run coverage:

```bash
make coverage
```

Infrastructure-related functionality should be tested at multiple levels where practical:

```text
Unit Tests
    │
    ▼
Component Tests
    │
    ▼
Integration Tests
    │
    ▼
Cluster Tests
```

Kubernetes and GPU functionality may require dedicated test environments.

---

# Design Principles

## 1. Infrastructure First

Kraken Pedestal focuses on infrastructure operations rather than application business logic.

Its primary concerns are:

```text
System
Kubernetes
Deployment
GPU
Operations
```

---

## 2. CLI Should Stay Thin

Command implementations should primarily handle:

* argument parsing
* flag parsing
* validation
* runtime invocation
* output

Infrastructure logic should remain outside the Cobra command implementation.

Prefer:

```text
cmd
 │
 ▼
service / runtime
 │
 ▼
implementation
```

instead of:

```text
cmd
 └── all infrastructure logic
```

---

## 3. Separate Operation From Presentation

Infrastructure components should return structured results.

The CLI decides how those results are rendered.

```text
Operation
    │
    ▼
Model
    │
    ├── Table
    ├── JSON
    └── YAML
```

---

## 4. Modular Infrastructure

Kubernetes, Helm, and GPU functionality should remain independently evolvable.

For example:

```text
Kubernetes
    │
    ├── Cluster
    │
    ├── Helm
    │
    └── GPU
          ├── HAMi
          └── DRA
```

A change to HAMi integration should not require changes to unrelated system-information functionality.

---

## 5. Reusable Runtime

Common infrastructure such as:

* configuration
* logging
* error handling
* Kubernetes clients
* output formatting

should be initialized and reused through the runtime layer.

---

# Roadmap

The project is currently focused on establishing a unified infrastructure operations foundation.

### Current

* [x] Go CLI foundation
* [x] Cobra command hierarchy
* [x] Basic system information
* [x] Common configuration
* [x] Runtime initialization
* [x] Structured logging
* [x] Common error handling
* [x] Table / JSON / YAML output
* [x] Basic Ops Tools
* [x] Kubernetes cluster functionality
* [x] Helm deployment functionality
* [x] HAMi integration
* [x] Kubernetes DRA exploration / implementation

### Planned

* [ ] Expand system diagnostic tools
* [ ] Expand Kubernetes cluster lifecycle operations
* [ ] Improve cluster preflight checks
* [ ] Improve Helm application lifecycle management
* [ ] Expand GPU diagnostics
* [ ] Expand HAMi operations
* [ ] Expand DRA resource management
* [ ] GPU health and fault diagnostics
* [ ] More comprehensive integration tests
* [ ] Cross-platform improvements
* [ ] Automated release pipeline
* [ ] More detailed documentation and examples

---

# Roadmap Direction

The long-term direction of Kraken Pedestal can be summarized as:

```text
                 Kraken Pedestal
                       │
        ┌──────────────┼──────────────┐
        │              │              │
      System        Kubernetes      GPU
        │              │              │
        │              │        ┌─────┴─────┐
        │              │        │           │
        │            Helm      HAMi        DRA
        │              │        │           │
        └──────────────┴────────┴───────────┘
                       │
                       ▼
             Infrastructure Platform
```

The ultimate goal is to make common infrastructure workflows accessible through one consistent operational framework.

---

# Contributing

Contributions are welcome.

Before submitting a pull request:

1. Format the code.
2. Run static analysis.
3. Run tests.
4. Build the project.
5. Keep commits focused.
6. Update documentation when behavior changes.

Recommended validation:

```bash
make fmt
make check
make test
make build
```

For infrastructure-related changes, please also describe:

* supported environment
* Kubernetes version
* runtime requirements
* required system dependencies
* GPU requirements, if applicable
* expected operational behavior

---

# Commit Convention

The project follows a conventional commit style:

```text
<type>(<scope>): <description>
```

Examples:

```text
feat(basic): add system information command
```

```text
feat(cluster): add kubeadm cluster initialization
```

```text
feat(helm): add application deployment
```

```text
feat(gpu): add hami resource management
```

```text
feat(dra): add gpu resource claim support
```

```text
fix(cluster): fix kubelet initialization
```

```text
refactor(logger): simplify slog handler
```

```text
test(gpu): add hami integration tests
```

```text
docs(readme): update project architecture
```

Common commit types:

| Type       | Description             |
| ---------- | ----------------------- |
| `feat`     | New functionality       |
| `fix`      | Bug fix                 |
| `refactor` | Code restructuring      |
| `perf`     | Performance improvement |
| `test`     | Tests                   |
| `docs`     | Documentation           |
| `build`    | Build system            |
| `ci`       | CI/CD                   |
| `chore`    | Maintenance             |

---

# Project Status

Kraken Pedestal is currently under active development.

The project is establishing a unified infrastructure operations framework covering:

```text
Linux
  ↓
System Operations
  ↓
Kubernetes
  ↓
Helm
  ↓
GPU Infrastructure
  ↓
HAMi / DRA
```

APIs, command structures, configuration formats, and internal package layouts may change before the project reaches a stable release.

For production environments, use a tagged release when available.

---

# License

See the `LICENSE` file for the applicable license.

---

# Acknowledgements

Kraken Pedestal builds upon the Kubernetes and cloud-native ecosystem.

The project integrates with or is designed to work alongside technologies including:

* Kubernetes
* kubeadm
* containerd
* Helm
* HAMi
* Kubernetes DRA
* Go
* Cobra
* `log/slog`

---

<p align="center">
  <strong>Kraken Pedestal</strong>
</p>

<p align="center">
  Kubernetes · Infrastructure · Operations · GPU
</p>
