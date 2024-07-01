package cluster

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove cluster",
	Run: func(cmd *cobra.Command, args []string) {
		err := removeCluster(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	removeCmd.Aliases = append(removeCmd.Aliases, "rm")
	startCmd.Flags().StringP("name", "n", "", "cluster name")

}
