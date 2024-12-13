package daemon_cmd

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Hidden: true,
	Use:    "daemon",
	Short:  "daemon commands to manage kloudlite daemon",
}

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.Aliases = append(Cmd.Aliases, "d")
}
