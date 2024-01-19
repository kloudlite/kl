package vpn

import (
	"github.com/kloudlite/kl/domain/client"
	"os"
	"time"

	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var reconnectVerbose bool
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart vpn device",
	Long: `This command let you restart vpn device.
Example:
  # restart vpn device
  sudo kl vpn restart
	`,
	Run: func(_ *cobra.Command, _ []string) {
		err := restartVPN()
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func restartVPN() error {

	if euid := os.Geteuid(); euid != 0 {
		fn.Log(
			text.Colored("make sure you are running command with sudo", 209),
		)
		return nil
	}

	wgInterface, err := wgc.Show(&wgc.WgShowOptions{
		Interface: "interfaces",
	})

	if err != nil {
		return err
	}

	if len(wgInterface) == 0 {
		fn.Log(text.Colored("[#] no devices connected yet", 209))
	} else {
		if err := disconnect(reconnectVerbose); err != nil {
			return err
		}
		fn.Log("[#] disconnected")
	}
	fn.Log("[#] connecting")
	time.Sleep(time.Second * 1)

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	startServiceInBg(devName)
	if err := connect(reconnectVerbose); err != nil {
		return err
	}

	fn.Log("[#] connected")
	fn.Log("[#] reconnection done")

	if _, err = wgc.Show(nil); err != nil {
		return err
	}

	s, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	fn.Log(text.Bold(text.Green("\n[#]Selected Device:")),
		text.Red(s),
	)
	return nil
}

func init() {
	restartCmd.Flags().BoolVarP(&reconnectVerbose, "verbose", "v", false, "show verbose")
	restartCmd.Aliases = append(restartCmd.Aliases, "reconnect")
}
