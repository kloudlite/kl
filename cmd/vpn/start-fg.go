package vpn

import (
	"errors"
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var startFgCmd = &cobra.Command{
	Use:    "start-fg",
	Short:  "start vpn foreground",
	Hidden: true,
	Run: func(cmd *cobra.Command, _ []string) {
		err := startVPNFG(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startVPNFG(cmd *cobra.Command) error {

	if euid := os.Geteuid(); euid != 0 {
		fn.Log(
			text.Colored("make sure you are running command with sudo", 209),
		)
		return nil
	}

	devName := fn.ParseStringFlag(cmd, "device")
	if devName == "" {
		return errors.New("device name is required")
	}

	if err := startService(devName, false); err != nil {
		return err
	}
	return nil
}

func init() {
	startFgCmd.Flags().StringP("device", "d", "", "device name")
}
