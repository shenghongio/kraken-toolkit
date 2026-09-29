package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/kraken-pedestal/internal/deploy/cluster/kubeadm"
	"github.com/spf13/cobra"
)

// NewPhaseCommand  创建kraken cluster phase 子命令
func NewPhaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phase",
		Short: "Manage kubeadm phases",
		Long:  "list or run individual kubeadm phases",
	}
	cmd.AddCommand(
		NewPhaseListCmd(),
		NewPhaseRunCmd(),
	)
	return cmd
}

// NewPhaseListCmd 列出所有phase
func NewPhaseListCmd() *cobra.Command {
	var action string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all kubeadm phases and their dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := kubeadm.NewPhaseRegistry()
			// 按照关键匹配
			switch action {
			case "init":
				kubeadm.RegisterInitPhases(registry)
			case "join":
				kubeadm.RegisterJoinPhases(registry)
			default:
				return fmt.Errorf("unknown phase action: %s (use init or join)", action)
			}
			sorted, err := registry.Sorted()
			if err != nil {
				return err
			}
			fmt.Printf("%-25s %-30s %s\n", "Phase", "Dependencies", "Description")
			fmt.Println(strings.Repeat("-", 80))
			for _, p := range sorted {
				deps := strings.Join(p.Dependencies(), ", ")
				if deps == "" {
					deps = "-"
				}
				fmt.Printf("%-25s %-30s %s\n", p.Name(), deps, p.Description())
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&action, "action", "init", "Action to perform [init|join]")
	return cmd
}

// NewPhaseRunCmd 运行自定phase
func NewPhaseRunCmd() *cobra.Command {
	var (
		action string
		name   string
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a specific kubeadm phases",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--phase is required")
			}
			registry := kubeadm.NewPhaseRegistry()
			switch action {
			case "init":
				kubeadm.RegisterInitPhases(registry)
			case "join":
				kubeadm.RegisterJoinPhases(registry)
			default:
				return fmt.Errorf("unknown action: %s", action)
			}
			phaseConfig := kubeadm.PhaseConfig{
				KubeadmConfigPath: "",
				KubeadmBinary:     "kubeadm",
				DryRun:            dryRun,
			}
			exec := func(ctx context.Context, p kubeadm.Phase, cfg kubeadm.PhaseConfig) error {
				fmt.Printf("  -> %s: %v\n", p.Name(), p.Command(cfg))
				return nil
			}
			return registry.RunOne(cmd.Context(), name, phaseConfig, exec)
		},
	}
	cmd.Flags().StringVar(&action, "action", "init", "Action to perform [init|join]")
	cmd.Flags().StringVar(&name, "phase", "", "Phase name to run")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without executing")
	return cmd
}
