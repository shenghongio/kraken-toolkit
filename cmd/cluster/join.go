package cluster

import (
	"context"
	"fmt"
	
	"github.com/kraken-pedestal/internal/config"
	"github.com/kraken-pedestal/internal/deploy/cluster/kubeadm"
	"github.com/spf13/cobra"
)

// NewJoinCmd 创建kraken cluster join命令
func NewJoinCommand() *cobra.Command {
	var (
		dryRun bool
		role   string
		phase  string
	)
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join a node to an existing Kubernetes cluster",
		Long:  "Run kubeadm join phases to add a node to an existing cluster.",
		Example: "kraken cluster join --config cluster.yaml --role worker\n" +
			"  kraken cluster join --config cluster.yaml --role control-plane --dry-run",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, ok := config.FromContext(cmd.Context())
			if !ok {
				return fmt.Errorf("config not found in context")
			}
			// 创建Kubeadm 实例
			newKubeadm, err := kubeadm.NewKubeadm(kubeadm.KubeadmConfig{
				BinaryPath:      "",
				ConfigPath:      "",
				ExpectedVersion: cfg.Cluster.KubernetesVersion,
			})
			if err != nil {
				return fmt.Errorf("create kubeadm: %w", err)
			}
			// 生成join 配置 (简化 toke 等参数)
			// TODO: 从配置或者命令行读取 token,apiServerEndpoint,caCertHash
			nodeAddress := cfg.Cluster.Nodes[0].Address
			if err := newKubeadm.GenerateJoinConfig(&cfg.Cluster, nodeAddress, "", "", "", role == "control-plane"); err != nil {
				return fmt.Errorf("generate join config: %w", err)
			}
			// 注册join phases
			phaseRegistry := kubeadm.NewPhaseRegistry()
			kubeadm.RegisterJoinPhases(phaseRegistry)
			
			// 构建 PhaseConfig
			phaseConfig := kubeadm.PhaseConfig{
				KubeadmConfigPath: newKubeadm.ConfigFilePath(),
				KubeadmBinary:     newKubeadm.BinaryPath(),
				NodeAddress:       nodeAddress,
				DryRun:            dryRun,
			}
			// 执行
			exec := func(ctx context.Context, p kubeadm.Phase, cfg kubeadm.PhaseConfig) error {
				args := p.Command(cfg)
				fmt.Printf(" -> %s: %v\n", p.Name(), args)
				return nil
			}
			if phase != "" {
				return phaseRegistry.RunOne(cmd.Context(), phase, phaseConfig, exec)
			}
			return phaseRegistry.RunAll(cmd.Context(), phaseConfig, exec)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview phases without executing")
	cmd.Flags().StringVar(&role, "role", "worker", "Node role: worker or control-plane")
	cmd.Flags().StringVar(&phase, "phase", "", "Run a specific phase and its dependencies")
	return cmd
}
