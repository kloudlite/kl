package intercept

import (
	"github.com/kloudlite/kl/domain/clients"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [app_name]",
	Short: "stop tunneling the traffic to your device",
	Long: `stop intercept service to stop tunnel traffic to your device
Examples:
	# close intercept service
  kl intercept stop [app_name]
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		apic := clients.Api
		fc := clients.File

		currentAcc, err := fc.GetDirTeam()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := apic.RemoveAllIntercepts(fn.MakeOption("teamName", currentAcc)); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("intercepted service stopped successfully")
	},
}

func init() {
	// stopCmd.Flags().StringP("service", "a", "", "service name")

	stopCmd.Aliases = append(stopCmd.Aliases, "close", "end", "leave", "quit", "terminate", "exit", "remove", "disconnect")
}
