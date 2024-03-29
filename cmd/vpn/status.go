package vpn

import (
	"fmt"
	"os"
	"runtime"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
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
	Run: func(cmd *cobra.Command, _ []string) {

		if runtime.GOOS != "windows" {
			if euid := os.Geteuid(); euid != 0 {
				fn.Log(
					text.Colored("make sure you are running command with sudo", 209),
				)
				return
			}
		}

		_, err := wgc.Show(nil)
		if err != nil {
			fn.PrintError(err)
			return
		}

		s, err := client.CurrentDeviceName()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log(text.Bold(text.Green("\n[#]Selected Device: ")), text.Red(s), "\n")

		dev, err := server.GetDevice([]fn.Option{
			fn.MakeOption("deviceName", s),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(dev.Spec.Ports) != 0 {
			fn.Log(text.Bold(text.Green("\n[#]Exposed Ports: ")))
			for _, v := range dev.Spec.Ports {
				fn.Log(text.Blue(fmt.Sprintf("%d:%d\t", v.Port, v.TargetPort)))
			}
		} else {
			fn.Warn(fmt.Sprintf("[#] no ports exposed, you can expose ports using `%s vpn expose` command", flags.CliName))
		}

	},
}

func init() {
	statusCmd.Aliases = append(statusCmd.Aliases, "show")
}
