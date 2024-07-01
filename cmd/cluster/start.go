package cluster

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start cluster",
	Run: func(cmd *cobra.Command, args []string) {
		err := startCluster(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	startCmd.Aliases = append(startCmd.Aliases, "up")
	startCmd.Flags().StringP("name", "n", "", "cluster name")
}
