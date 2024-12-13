package vpn

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/domain/clients"
	daemon_server "github.com/kloudlite/kl/domain/daemon-server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Hidden: true,
	Use:    "status",
	Short:  "show vpn status",
	Long: `This command let you show vpn status.
Example:
  # show vpn status
  sudo kl vpn status
	`,
	Run: func(_ *cobra.Command, _ []string) {

		if euid := os.Geteuid(); euid != 0 {
			if os.Getenv("KL_APP") != "true" {
				if err := func() error {

					p, err := daemon_server.NewProxyWithService(true)
					if err != nil {
						return err
					}

					out, err := p.WgStatus()
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

		res, err := wgc.Show(nil)
		if err != nil {
			fn.PrintError(err)
			return
		}

		print := func(devName string) {
			d := ""
			if devName != "" {
				d = fmt.Sprintf("%s ", devName)
			}
			n := ""

			if len(res) == 0 {
				n = "not "
			}
			fn.Logf(text.Bold("\n[#] device %s%s"), text.Red(d), text.Bold(fmt.Sprintf("is %sconnected", n)))
		}

		apic := clients.Api
		if err != nil {
			fn.PrintError(err)
			return
		}

		dev, _ := apic.EnsureDevice()
		if err != nil {
			fn.Logf(text.Bold("\n [#] disconnected device"))
			fn.PrintError(err)
			return
		}

		print(dev.DeviceName)
	},
}

func init() {
	statusCmd.Aliases = append(statusCmd.Aliases, "show")
}
