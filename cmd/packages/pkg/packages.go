package pkg

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pkg",
	Short: "packages util to manage nix packages of kl shell",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "packages")
	Cmd.Aliases = append(Cmd.Aliases, "package")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(rmCmd)
}
