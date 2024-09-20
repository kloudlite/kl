package cluster

import "github.com/spf13/cobra"

var Command = &cobra.Command{
	Use:   "cluster",
	Short: "manage local clusters",
	Long:  "This command is used to manage local clusters",
}

func init() {
	Command.AddCommand(connectCmd)
	Command.Aliases = []string{"clus", "clusters"}
}
