package app

import (
	daemon_server "github.com/kloudlite/kl/domain/daemon-server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop the kloudlite controller app",
	Long:  `This is internal command`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := Stop(); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("app stopped successfully")
	},
}

func Stop() error {
	p, err := daemon_server.NewProxyWithService(true, false)
	if err != nil {
		return err
	}

	if err := p.Exit(); err != nil {
		return err
	}

	return nil
}
