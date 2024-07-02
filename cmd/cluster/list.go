package cluster

import (
	"fmt"

	"github.com/kloudlite/kl/cmd/cluster/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list clusters",
	Run: func(cmd *cobra.Command, args []string) {

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fn.PrintError(err)
			return
		}

		clusterClient, err := k3s.NewK3sClient(verbose)
		if err != nil {
			fn.PrintError(err)
			return
		}
		clusters, err := clusterClient.ListClusters()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(clusters) == 0 {
			fn.Log("No clusters created yet. Create a cluster using `kl cluster create` command.")
			return
		}

		header := table.Row{
			table.HeaderText("Cluster"),
			table.HeaderText("Account"),
			table.HeaderText("Status"),
		}
		rows := []table.Row{}
		for _, fCluster := range clusters {
			rows = append(rows, table.Row{
				fCluster.Name,
				fCluster.AccountName,
				fCluster.Status,
			})
		}
		fmt.Println(table.Table(&header, rows))
	},
}

func init() {
	listCmd.Aliases = append(listCmd.Aliases, "ls", "list")
}
