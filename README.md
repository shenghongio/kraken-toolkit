# Kraken Toolkit

<p align="center">
  <strong>Kraken Toolkit</strong> — Kubernetes & AI Infrastructure Operations Toolkit
</p>

<p align="center">
  A Go-based toolkit for system operations, Kubernetes cluster management,
  Helm deployment, and GPU resource scheduling.
</p>

<p align="center">
  <a href="https://github.com/shenghongio/kraken-toolkit">GitHub</a>
  · 
  <a href="https://github.com/shenghongio/kraken-toolkit/issues">Issues</a>
</p>

<p align="center">
  <img src="https://cdn.simpleicons.org/kubernetes" width="36" alt="Kubernetes">
  <img src="https://cdn.simpleicons.org/go" width="36" alt="Go">
  <img src="https://cdn.simpleicons.org/helm" width="36" alt="Helm">
  <img src="https://cdn.simpleicons.org/linux" width="36" alt="Linux">
  <img src="https://cdn.simpleicons.org/nvidia" width="36" alt="NVIDIA">
</p>

---

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                          CLI                               │
│  root (kraken)                                             │
│    ├── bc        Basic commands                            │
│    │   ├── cmd       Execute shell commands on hosts       │
│    │   ├── adduser   Create management user + deploy SSH   │
│    │   └── script    Execute local scripts on hosts        │
│    ├── version       Print version information             │
│    └── completion    Generate shell completion             │
├────────────────────────────────────────────────────────────┤
│                       Internal                             │
│  ┌──────────┐  ┌───────────┐  ┌────────────────────────┐  │
│  │ config   │  │ executor  │  │ kubeadm                │  │
│  │ YAML     │  │ SSH-based │  │ Phase interface        │  │
│  │ loading  │  │ concurrent│  │ PhaseRegistry (DAG)    │  │
│  │ & inject │  │ execution │  │ Kubeadm wrapper        │  │
│  └──────────┘  └───────────┘  └────────────────────────┘  │
├────────────────────────────────────────────────────────────┤
│                      pkg / utils                           │
│  ┌──────────┐  ┌──────────┐  ┌────────────────────────┐  │
│  │ logger   │  │ cli      │  │ utils                  │  │
│  │ slog +   │  │ groups,  │  │ result display,        │  │
│  │ console  │  │ version  │  │ help formatting        │  │
│  └──────────┘  └──────────┘  └────────────────────────┘  │
└────────────────────────────────────────────────────────────┘
```

---

## Quick Start

```bash
# Build
make build

# Run
./bin/kraken --help
./bin/kraken bc cmd -c "uptime" --ssh-host=<IP>
```

## License

MIT