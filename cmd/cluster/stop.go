package cluster

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop cluster",
	Run: func(cmd *cobra.Command, args []string) {
		err := stopCluster(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	stopCmd.Aliases = append(stopCmd.Aliases, "down")
	startCmd.Flags().StringP("name", "n", "", "cluster name")
}
