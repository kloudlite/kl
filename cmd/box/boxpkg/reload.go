package boxpkg

import (
	"fmt"
	"github.com/kloudlite/kl/domain/server"

	cl "github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func (c *client) Reload() error {
	defer spinner.Client.Start("Reloading environments please wait")()

	if err := server.SyncBoxHash(); err != nil {
		return err
	}
	//if err := server.SyncDevboxJsonFile(); err != nil {
	//	return err
	//}

	return cl.ExecPackageCommand(fmt.Sprintf("devbox install%s", func() string {
		if c.verbose {
			return ""
		}
		return " -q"
	}()), c.cmd)
}
