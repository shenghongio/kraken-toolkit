package cluster

import (
	"context"
	"fmt"
	
	"github.com/kraken-toolkit/internal/config"
	"github.com/kraken-toolkit/internal/deploy/cluster/kubeadm"
	"github.com/spf13/cobra"
)

// NewInitCmd 创建 kraken cluster init 命令
func NewInitCommand() *cobra.Command {
	var (
		dryRun bool
		phase  string
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Kubernetes cluster",
		Long:  "Run kubeadm init phases to bootstrap a new Kubernetes control plane.",
		Example: "kraken cluster init --config cluster.yaml\n" +
			"  kraken cluster init --config cluster.yaml --dry-run\n" +
			"  kraken cluster init --config cluster.yaml --phase certs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fromContext, ok := config.FromContext(cmd.Context())
			if !ok {
				return fmt.Errorf("config not found in context")
			}
			// 创建kubeadm 实例
			kadm, err := kubeadm.NewKubeadm(kubeadm.KubeadmConfig{
				BinaryPath:      "",
				ConfigPath:      "",
				ExpectedVersion: fromContext.Cluster.KubernetesVersion,
			})
			if err != nil {
				return fmt.Errorf("create kubeadm: %w", err)
			}
			// 生成kubeadm 配置
			nodeAddr := fromContext.Cluster.Nodes[0].Address // 简化取第一个节点
			if err := kadm.GenerateInitConfig(&fromContext.Cluster, nodeAddr); err != nil {
				return fmt.Errorf("generate init config: %w", err)
			}
			// 注册 init phases
			registry := kubeadm.NewPhaseRegistry()
			kubeadm.RegisterInitPhases(registry)
			
			// 构建 PhaseConfig
			phaseConfig := kubeadm.PhaseConfig{
				KubeadmConfigPath: kadm.ConfigFilePath(),
				KubeadmBinary:     kadm.BinaryPath(),
				NodeAddress:       nodeAddr,
				DryRun:            dryRun,
				ExtraArgs:         nil,
			}
			
			// 执行
			exec := func(ctx context.Context, p kubeadm.Phase, cfg kubeadm.PhaseConfig) error {
				command := p.Command(cfg)
				fmt.Printf("  -> %s: %v\n", p.Name(), command)
				
				// TODO: 实际执行，本地或者ssh
				return nil
			}
			if phase != "" {
				return registry.RunOne(cmd.Context(), phase, phaseConfig, exec)
			}
			return registry.RunAll(cmd.Context(), phaseConfig, exec)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Don't actually do anything")
	cmd.Flags().StringVar(&phase, "phase", "", "Run a specific phase and its dependencies")
	return cmd
}
