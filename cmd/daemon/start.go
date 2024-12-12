package daemon_cmd

import (
	"os"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/daemon"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "up",
	Short: "start the kloudlite daemon service",
	Run: func(c *cobra.Command, _ []string) {

		if runtime.GOOS != constants.RuntimeWindows {
			if euid := os.Geteuid(); euid != 0 {
				fn.Log(text.Colored("make sure you are running command with sudo", 209))
				return
			}
		}

		if err := daemon.RunApp(c.Parent().Parent().Name()); err != nil {
			fn.PrintError(err)
		}
	},
}

func init() {
	startCmd.Aliases = append(startCmd.Aliases, "start")
}
