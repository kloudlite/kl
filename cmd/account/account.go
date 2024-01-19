package account

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "account",
	Short: "switch account context",
	Long: `Use this command to switch account context
Examples:
  # switch account context
  kl account switch <account_name>
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "acc")
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(switchCmd)
}
