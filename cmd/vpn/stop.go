package vpn

import (
	"os"

	"github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop vpn device",
	Long: `This command let you stop running vpn device.
Example:
  # stop vpn device
  sudo kl vpn stop
	`,
	Run: func(cmd *cobra.Command, _ []string) {

		verbose := fn.ParseBoolFlag(cmd, "verbose")

		// if runtime.GOOS == constants.RuntimeWindows {
		// 	if err := disconnect(verbose); err != nil {
		// 		fn.Notify("Error:", err.Error())
		// 		fn.PrintError(err)
		// 	}
		// 	return
		// }

		if euid := os.Geteuid(); euid != 0 {
			if os.Getenv("KL_APP") != "true" {
				if err := func() error {

					if err := client.EnsureAppRunning(); err != nil {
						return err
					}

					p, err := proxy.NewProxy(true)
					if err != nil {
						return err
					}

					out, err := p.Stop()
					if err != nil {
						return err
					}

					fn.Log(string(out))
					return nil
				}(); err != nil {
					fn.PrintError(err)
					return
				}

				return
			}
		}

		// if euid := os.Geteuid(); euid != 0 {
		// 	fn.Log(
		// 		text.Colored("make sure you are running command with sudo", 209),
		// 	)
		// 	return
		// }

		wgInterface, err := wgc.Show(&wgc.WgShowOptions{
			Interface: "interfaces",
		})

		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(wgInterface) == 0 {
			fn.Log(text.Colored("[#] no device connected yet", 209))
			return
		}

		err = disconnect(verbose)
		if err != nil {
			fn.PrintError(err)
			return
		}

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.Logf(text.Bold("\n [#] disconnected device"), text.Blue(s))
			fn.PrintError(err)
			return
		}

		fn.Logf(text.Bold("\n[#] disconnected device %s"), text.Blue(s))
	},
}

func init() {
	stopCmd.Flags().BoolP("verbose", "v", false, "run in debug mode")

	stopCmd.Aliases = append(stopCmd.Aliases, "disconnect")
}
