package lib

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "lib",
	Short: "library util to manage nix libraries of kl shell",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "libs")
	Cmd.Aliases = append(Cmd.Aliases, "libraries")
	Cmd.Aliases = append(Cmd.Aliases, "library")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(rmCmd)
}
