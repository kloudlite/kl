package context

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "context",
	Short: "create new context and manage existing contexts",
	Long: `Create new context and manage existing contexts
Examples:
  # creating new context
  kl context new

  # list all contexts
  kl context list

  # switch to context
  kl context switch <context_name>

  # remove context
  kl context remove <context_name>
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "ctx")

	Cmd.AddCommand(newCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(removeCmd)
}