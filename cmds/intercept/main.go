package intercept

import "github.com/spf13/cobra"

var Command = &cobra.Command{
	Use:   "add",
	Short: "add environment resources to your kl-config file",
	Long:  "This command will add the environment resources to your kl-config file",
}

func init() {
	Command.AddCommand(startCmd)
	Command.AddCommand(stopCmd)
}
