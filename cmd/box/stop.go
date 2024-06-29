package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop running box",
	Run: func(cmd *cobra.Command, args []string) {

		fn.Logf(text.Yellow("[#] current process will be stopped, do you want to stop the container? [Y/n] "))
		if !fn.Confirm("Y", "Y") {
			return
		}

		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Stop(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}
