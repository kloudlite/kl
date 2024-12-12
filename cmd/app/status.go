package app

import (
	daemon_server "github.com/kloudlite/kl/domain/daemon-server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "stop the kloudlite controller app",
	Long:  `This is internal command`,
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
		fn.Log("app is running")
		return nil
	}

	fn.Log("app is not running")
	return nil
}
