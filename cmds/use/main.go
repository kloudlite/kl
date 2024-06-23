package use

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "use",
	Short: "to switch env use `kl use env` command",
}

func init() {
	Cmd.AddCommand(envCmd)
}
