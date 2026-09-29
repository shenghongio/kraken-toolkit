package kubeadm

// PreflightPhase 初始化前校验

type PreflightPhase struct{}

func (p *PreflightPhase) Name() string           { return "preflight" }
func (p *PreflightPhase) Description() string    { return "run pre-flight checks" }
func (p *PreflightPhase) Dependencies() []string { return nil }
func (p *PreflightPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "preflight", "--config", cfg.KubeadmConfigPath}
}

// CertsPhase  生成证书
type CertsPhase struct{}

func (c *CertsPhase) Name() string           { return "certs" }
func (c *CertsPhase) Description() string    { return "Generate all certificates" }
func (c *CertsPhase) Dependencies() []string { return []string{"preflight"} }
func (c *CertsPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "certs", "all", "--config", cfg.KubeadmConfigPath}
}

// KubeconfigPhase 生成所有kubeconfig文件
type KubeconfigPhase struct{}

func (p *KubeconfigPhase) Name() string           { return "kubeconfig" }
func (p *KubeconfigPhase) Description() string    { return "Generate all kubeconfig files" }
func (p *KubeconfigPhase) Dependencies() []string { return []string{"certs"} }
func (p *KubeconfigPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "kubeconfig", "--config", cfg.KubeadmConfigPath}
}

//ControlPlanePhase 生成静态pod清单
type ControlPlanePhase struct{}

func (p *ControlPlanePhase) Name() string { return "control-plane" }
func (p *ControlPlanePhase) Description() string {
	return "Generate static pod manifest for control plane components"
}
func (p *ControlPlanePhase) Dependencies() []string { return []string{"certs"} }
func (p *ControlPlanePhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "control-plane"}
}

// EtcdPhase 生成 etcd静态 Pod 清单
type EtcdPhase struct{}

func (p *EtcdPhase) Name() string           { return "etcd" }
func (p *EtcdPhase) Description() string    { return "Generate static pod manifest for etcd components" }
func (p *EtcdPhase) Dependencies() []string { return []string{"certs"} }
func (p *EtcdPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "etcd", "local", "--config", cfg.KubeadmConfigPath}
}

// WaitControlPlanePhase 等待控制平台就绪
type WaitControlPlanePhase struct{}

func (p *WaitControlPlanePhase) Name() string           { return "wait-control-plane" }
func (p *WaitControlPlanePhase) Description() string    { return "Wait for the control plane to be ready" }
func (p *WaitControlPlanePhase) Dependencies() []string { return []string{"certs"} }
func (p *WaitControlPlanePhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "wait-control-plane", cfg.KubeadmConfigPath}
}

// KubeletStartPhase 写入kubelet 配置并且启动
type KubeletStartPhase struct{}

func (p *KubeletStartPhase) Name() string           { return "kubelet-start" }
func (p *KubeletStartPhase) Description() string    { return "Write kubelet config and start kubelet" }
func (p *KubeletStartPhase) Dependencies() []string { return []string{"preflight"} }
func (p *KubeletStartPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "kubelet-start", "--config", cfg.KubeadmConfigPath}
}

// UploadConfigPhase 上传kubeadm 配置到集群
type UploadConfigPhase struct{}

func (p *UploadConfigPhase) Name() string           { return "upload-config" }
func (p *UploadConfigPhase) Description() string    { return "Upload kubeadm and kubelet configuration " }
func (p *UploadConfigPhase) Dependencies() []string { return []string{"kubelet-start", "control-plane"} }
func (p *UploadConfigPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "upload-config", "all", "--config", cfg.KubeadmConfigPath}
}

// UploadCertsPhase 上传证书到集群
type UploadCertsPhase struct{}

func (p *UploadCertsPhase) Name() string           { return "upload-certs" }
func (p *UploadCertsPhase) Description() string    { return "Upload certificates to cluster" }
func (p *UploadCertsPhase) Dependencies() []string { return []string{"kubelet-start", "control-plane"} }
func (p UploadCertsPhase) Command(cfg PhaseConfig) []string {
	args := []string{cfg.KubeadmBinary, "init", "phase", "upload-certs", "--upload-certs", "--config", cfg.KubeadmConfigPath}
	if cfg.ExtraArgs != nil {
		if certKey, ok := cfg.ExtraArgs["certificate-key"]; ok {
			args = append(args, "--certificate-key", certKey)
		}
	}
	return args
}

// MarkControlPlanePhase 标记控制平台节点
type MarkControlPlanePhase struct{}

func (p *MarkControlPlanePhase) Name() string           { return "mark-control-plane" }
func (p *MarkControlPlanePhase) Description() string    { return "Mark node as control-plane" }
func (p *MarkControlPlanePhase) Dependencies() []string { return []string{"kubelet-start"} }
func (p *MarkControlPlanePhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "mark-control-plane", "--config", cfg.KubeadmConfigPath}
}

// BootstrapTokenPhase 生成引导令牌
type BootstrapTokenPhase struct{}

func (p *BootstrapTokenPhase) Name() string { return "bootstrap-token" }
func (p *BootstrapTokenPhase) Description() string {
	return "Generate bootstrap token for joining nodes"
}
func (p *BootstrapTokenPhase) Dependencies() []string { return []string{"kubelet-start"} }
func (p *BootstrapTokenPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "bootstrap-token", "--config", cfg.KubeadmConfigPath}
}

// AddonPhase 安装插件(kube-proxy 和CoreDNS)
type AddonPhase struct{}

func (p *AddonPhase) Name() string        { return "addon" }
func (p *AddonPhase) Description() string { return "Install kube-proxy and CoreDNS addon" }
func (p *AddonPhase) Dependencies() []string {
	return []string{"upload-config", "upload-certs"}
}
func (p *AddonPhase) Command(cfg PhaseConfig) []string {
	return []string{cfg.KubeadmBinary, "init", "phase", "addon", "all", "all", "--config", cfg.KubeadmConfigPath}
}

// RegisterInitPhases 将12个init phase 注册到PhaseRegistry
func RegisterInitPhases(registry *PhaseRegistry) {
	registry.Register(&PreflightPhase{})
	registry.Register(&CertsPhase{})
	registry.Register(&KubeconfigPhase{})
	registry.Register(&ControlPlanePhase{})
	registry.Register(&EtcdPhase{})
	registry.Register(&WaitControlPlanePhase{})
	registry.Register(&KubeletStartPhase{})
	registry.Register(&UploadConfigPhase{})
	registry.Register(&UploadCertsPhase{})
	registry.Register(&MarkControlPlanePhase{})
	registry.Register(&BootstrapTokenPhase{})
	registry.Register(&AddonPhase{})
}
