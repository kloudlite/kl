package vpn

import (
	"github.com/kloudlite/kl/domain/envclient"
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

	if !envclient.IsBoxMode() {
		Cmd.AddCommand(startCmd)
		Cmd.AddCommand(startFgCmd)
		Cmd.AddCommand(restartCmd)
		Cmd.AddCommand(stopCmd)
	}

	Cmd.AddCommand(statusCmd)
}
