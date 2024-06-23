package get

import "github.com/spf13/cobra"

var Command = &cobra.Command{
	Use:   "get",
	Short: "get config/secret entries",
}

func init() {
	Command.AddCommand(configCmd)
	Command.AddCommand(secretCmd)
}
