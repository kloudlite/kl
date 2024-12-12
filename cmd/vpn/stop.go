package vpn

import (
	"os"

	"github.com/kloudlite/kl/domain/apiclient"
	daemon_server "github.com/kloudlite/kl/domain/daemon-server"
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

					p, err := daemon_server.NewProxyWithService(true)
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

		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		dev, err := apic.EnsureDevice()
		if err != nil {
			fn.Logf(text.Bold("\n [#] disconnected device"))
			fn.PrintError(err)
			return
		}

		fn.Logf(text.Bold("\n[#] disconnected device %s"), text.Blue(dev.DeviceName))
	},
}

func init() {
	stopCmd.Flags().BoolP("verbose", "v", false, "run in debug mode")

	stopCmd.Aliases = append(stopCmd.Aliases, "disconnect")
}
