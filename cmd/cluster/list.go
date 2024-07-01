package cluster

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list clusters",
	Run: func(cmd *cobra.Command, args []string) {
		err := listClusters(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	listCmd.Aliases = append(listCmd.Aliases, "ls")
	startCmd.Flags().StringP("name", "n", "", "cluster name")
}
