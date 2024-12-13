package cluster

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage clusters",
	Long:  `start and stop clusters.`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "clus", "clusters")

	upCmd.Aliases = append(upCmd.Aliases, "start", "connect")
	downCmd.Aliases = append(downCmd.Aliases, "stop", "disconnect")
	cleanCmd.Aliases = append(cleanCmd.Aliases, "delete", "clean")

	Cmd.AddCommand(downCmd)
	Cmd.AddCommand(upCmd)
	Cmd.AddCommand(cleanCmd)
}
