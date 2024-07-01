package cluster

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "manage clusters",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "clusters")
	Cmd.Aliases = append(Cmd.Aliases, "clus")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(removeCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
}
