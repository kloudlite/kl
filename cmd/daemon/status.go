package daemon_cmd

import (
	daemon_server "github.com/kloudlite/kl/domain/daemon-server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "stop the kloudlite daemon service",
	Run: func(_ *cobra.Command, _ []string) {
		if err := Status(); err != nil {
			fn.PrintError(err)
		}
	},
}

func Status() error {
	p, err := daemon_server.NewProxyWithService(true, false)
	if err != nil {
		return err
	}

	if p.Status() {
		fn.Log("kloudlite daemon is running")
		return nil
	}

	fn.Log("kloudlite daemon is not running")
	return nil
}
