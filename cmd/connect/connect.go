package connect

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "connect",
	Short: "start the wireguard connection",
	Long:  "This command will start the wireguard connection",
	Run: func(cmd *cobra.Command, _ []string) {
		if err := startWg(cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startWg(cmd *cobra.Command) error {
	defer spinner.Client.UpdateMessage("connecting your device")()
	k3sClient, err := k3s.NewClient(cmd)
	if err != nil {
		return fn.NewE(err)
	}

	if err := k3sClient.RestartWgProxyContainer(); err != nil {
		return fn.NewE(err)
	}

	fn.Log(text.Green("device connected"))

	return nil
}

func ChekcWireguardConnection() bool {
	file, err := os.Open("/kl-tmp/online.status")
	if err != nil {
		return false
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	status, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false
	}

	if strings.TrimSpace(status) == "online" {
		return true
	}

	return false
}
