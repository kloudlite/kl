package pkgs

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pkg",
	Short: "packages util to manage nix packages of kl box",
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "packages")
	Cmd.Aliases = append(Cmd.Aliases, "package")
	Cmd.AddCommand(addCmd)
	removeCmd.Aliases = append(removeCmd.Aliases, "rm")
	Cmd.AddCommand(removeCmd)
}
