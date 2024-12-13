package vpn

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Hidden: true,
	Use:    "vpn",
	Short:  "vpn related commands",
	// Example: Example,
	Long: `vpn related commands
Examples:
	`,
}

func init() {
	Cmd.Aliases = append(Cmd.Aliases, "dev")
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(startFgCmd)
	Cmd.AddCommand(restartCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(statusCmd)
}
