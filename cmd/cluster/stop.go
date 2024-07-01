package cluster

import (
	"github.com/kloudlite/kl/cmd/cluster/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fn.PrintError(fn.Error("cluster name is required"))
			cmd.Help()
			return
		}
		clusterClient, err := k3s.NewK3sClient()
		if err != nil {
			fn.PrintError(err)
			return
		}
		err = clusterClient.StopCluster(args[0])
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	stopCmd.Aliases = append(stopCmd.Aliases, "down")
}
