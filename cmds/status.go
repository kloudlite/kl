package cmds

import (
	"fmt"
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "get status of your current context (user, account, environment, vpn status)",
	Example: fn.Desc("{cmd} status"),
	Run: func(_ *cobra.Command, _ []string) {

		if user, err := server.GetCurrentUser(); err == nil {
			fmt.Printf("You are logged in as %s (%s)\n",
				text.Bold(text.Green(user.Name)),
				text.Blue(user.Email),
			)
		}

		fn.Println()

		klFile, err := klfile.GetKlFile("")
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold("account: "), text.Blue(klFile.AccountName)))
		}

		if cwd, err := os.Getwd(); err == nil {
			if le, err := server.EnvAtPath(cwd); err == nil {
				fn.Log(fmt.Sprint(text.Bold("environment: "), text.Blue(le.Name)))
			}
		}

		// if s, err := client.CurrentDeviceName(); err == nil {
		//
		// 	// dev, err := server.GetDevice(fn.MakeOption("deviceName", s))
		// 	// if err != nil {
		// 	// 	fn.PrintError(err)
		// 	// 	return
		// 	// }
		//
		// 	// switch flags.CliName {
		// 	// case constants.InfraCliName:
		// 	// 	fn.Log(fmt.Sprint(text.Bold("Cluster:"), dev.ClusterName))
		// 	// }
		//
		// 	b := server.CheckDeviceStatus()
		// 	fn.Log(fmt.Sprint(text.Bold(text.Blue("Device: ")), s, func() string {
		// 		if b {
		// 			return text.Bold(text.Green(" (Connected) "))
		// 		} else {
		// 			return text.Bold(text.Red(" (Disconnected) "))
		// 		}
		// 	}()))
		//
		// 	ip, err := client.CurrentDeviceIp()
		// 	if err == nil {
		// 		fn.Logf("%s %s", text.Bold(text.Blue("Device IP:")), *ip)
		// 	}
		// }
	},
}

