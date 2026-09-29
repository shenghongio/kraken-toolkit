package kubeadm

// JoinPreflightPhase join 预防检查
type JoinPreflightPhase struct{}

func (p *JoinPreflightPhase) Name() string           { return "preflight" }
func (p *JoinPreflightPhase) Description() string    { return "Run join pre-flight checks" }
func (p *JoinPreflightPhase) Dependencies() []string { return nil }
func (p *JoinPreflightPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "join", "phase", "preflight", "--config", cfg.KubeadmConfigPath}
}

// ControlPlanePreparePhase 控制平面准备 -- 仅 control-plans角色
type ControlPlanePreparePhase struct{}

func (p *ControlPlanePreparePhase) Name() string           { return "control-plane-prepare" }
func (p *ControlPlanePreparePhase) Description() string    { return "Prepare control plane components" }
func (p *ControlPlanePreparePhase) Dependencies() []string { return []string{"preflight"} }
func (p *ControlPlanePreparePhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "join", "phase", "control-plane-prepare", "all", "--config", cfg.KubeadmConfigPath}
}

// JoinKubeletStartPhase 启动kubelet
type JoinKubeletStartPhase struct{}

func (p *JoinKubeletStartPhase) Name() string { return "kubelet-start" }
func (p *JoinKubeletStartPhase) Description() string {
	return "Write kubelet settings and start kubelet"
}
func (p *JoinKubeletStartPhase) Dependencies() []string { return []string{"preflight"} }
func (p *JoinKubeletStartPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "join", "phase", "kubelet-start", "--config", cfg.KubeadmConfigPath}
}

// ControlPlaneJoinPhase 加入控制平面 -- control-plane角色
type ControlPlaneJoinPhase struct{}

func (p *ControlPlaneJoinPhase) Name() string        { return "control-plane-join" }
func (p *ControlPlaneJoinPhase) Description() string { return "join nods as control plane" }
func (p *ControlPlaneJoinPhase) Dependencies() []string {
	return []string{"control-plane-prepare", "kubelet-start"}
}
func (p *ControlPlaneJoinPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "join", "phase", "control-plane-join", "all", "--config", cfg.KubeadmConfigPath}
}

// JoinKubeconfigPhase  生成 kubeconfig -- 仅control-plane 角色
type JoinKubeconfigPhase struct{}

func (p *JoinKubeconfigPhase) Name() string        { return "kubeconfig" }
func (p *JoinKubeconfigPhase) Description() string { return "Generate kubeconfig for control plane" }
func (p *JoinKubeconfigPhase) Dependencies() []string {
	return []string{"control-plane-prepare"}
}
func (p *JoinKubeconfigPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "join", "phase", "kubeconfig", "--config", cfg.KubeadmConfigPath}
}

// RegisterJoinPhases 将5个 join phase 注册到 PhaseRegistry
func RegisterJoinPhases(r *PhaseRegistry) {
	r.Register(&JoinPreflightPhase{})
	r.Register(&ControlPlanePreparePhase{})
	r.Register(&JoinKubeletStartPhase{})
	r.Register(&ControlPlaneJoinPhase{})
	r.Register(&JoinKubeconfigPhase{})
}
