package vpn

import (
	"fmt"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"os/exec"
	"os/user"
	"runtime"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start vpn",
	Long:  `start vpn`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := startVPN(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startVPN() error {
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}
	config, err := fc.GetWGConfig()
	if err != nil {
		return fn.NewE(err)
	}

	wgConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 198.18.0.2/32

[Peer]
PublicKey = %s
AllowedIPs = 198.18.0.1/32, 100.64.0.0/10
PersistentKeepalive = 25
Endpoint = 127.0.0.1:51820
`, config.Host.PrivateKey, config.Proxy.PublicKey)

	if runtime.GOOS != "linux" {
		fn.Log(text.Green("add below config to your wireguard client and start vpn"))
		fn.Log(wgConfig)
		return nil
	}

	current, err := user.Current()
	if err != nil {
		return fn.NewE(err)
	}
	if current.Uid != "0" {
		return fmt.Errorf("root permission required")
	}

	if err := fc.SetWGConfig(wgConfig); err != nil {
		return fn.NewE(err)
	}

	err = exec.Command("wg-quick", "up", "kl").Run()
	if err != nil {
		return fn.NewE(err)
	}

	fn.Log(text.Green("kloudlite vpn has been started"))

	return nil
}
