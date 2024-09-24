package vpn

import (
	"fmt"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"os/exec"
	"os/user"
	"runtime"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop vpn",
	Long:  `stop vpn`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopVPN(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func stopVPN() error {

	if runtime.GOOS != "linux" {
		fn.Log(text.Green("stop vpn from your wireguard client"))
		return nil
	}

	current, err := user.Current()
	if err != nil {
		return fn.NewE(err)
	}

	if current.Uid != "0" {
		return fmt.Errorf("root permission required")
	}

	err = exec.Command("wg-quick", "down", "kl").Run()
	if err != nil {
		return fn.NewE(err)
	}

	fn.Log(text.Green("kloudlite vpn has been stopped"))

	return nil
}
