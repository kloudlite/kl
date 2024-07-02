package cluster

import (
	"github.com/kloudlite/kl/cmd/cluster/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove cluster",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fn.PrintError(err)
			return
		}
		if len(args) == 0 {
			fn.PrintError(fn.Error("cluster name is required"))
			cmd.Help()
			return
		}
		clusterClient, err := k3s.NewK3sClient(verbose)
		if err != nil {
			fn.PrintError(err)
			return
		}
		err = clusterClient.RemoveCluster(args[0])
		if err != nil {
			fn.PrintError(err)
			return
		}
		fn.Log("cluster removed")
	},
}

func init() {
	removeCmd.Aliases = append(removeCmd.Aliases, "rm", "delete")
}
