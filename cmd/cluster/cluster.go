package cluster

import (
	"log/slog"
	
	"github.com/kraken-pedestal/pkg/cli"
	"github.com/spf13/cobra"
)

// NewClusterCommand  创建 cluster 根命令
func NewClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Cluster Commands",
		Long: "Kubernetes cluster lifecycle management:\n" +
			"    * init   - Initialize a new Kubernetes cluster\n" +
			"    * join   - Join a node to an existing cluster\n" +
			"    * phase  - List or run individual kubeadm phases",
		Example: "kraken cluster init --config cluster.yaml\n" +
			"  kraken cluster join --config cluster.yaml --role worker\n" +
			"  kraken cluster phase list --action init",
		GroupID:      cli.GroupCluster,
		SilenceUsage: true,
	}
	slog.Debug("register cluster subcommands")
	cmd.AddCommand(
		NewInitCommand(),
		NewJoinCommand(),
		NewPhaseCommand(),
	)
	return cmd
}
