package packages

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

var LibCmd = &cobra.Command{
	Use:   "lib",
	Short: "libraries util to manage nix libraries of kl box",
}

func init() {
	LibCmd.Aliases = append(LibCmd.Aliases, "libraries")
	LibCmd.Aliases = append(LibCmd.Aliases, "library")

	//fileclient.OnlyInsideBox(listCmd)
	//fileclient.OnlyInsideBox(addCmd)
	//fileclient.OnlyInsideBox(rmCmd)

	LibCmd.AddCommand(listCmd)
	fileclient.OnlyInsideBox(addLibCmd)
	LibCmd.AddCommand(addLibCmd)
	fileclient.OnlyInsideBox(rmLibCmd)
	LibCmd.AddCommand(rmLibCmd)
	fileclient.OnlyInsideBox(searchCmd)
	LibCmd.AddCommand(searchCmd)
}
