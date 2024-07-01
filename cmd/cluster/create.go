package cluster

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create cluster",
	Run: func(cmd *cobra.Command, args []string) {
		err := createCluster(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	createCmd.Aliases = append(createCmd.Aliases, "add")
	startCmd.Flags().StringP("name", "n", "", "cluster name")
}
