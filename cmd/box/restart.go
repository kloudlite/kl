package box

import (
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart the box according to the current kl.yml configuration",
	Run: func(cmd *cobra.Command, args []string) {

		fn.Logf(text.Yellow("[#] current process will be stopped, do you want to restart the container? [Y/n] "))
		if !fn.Confirm(strings.ToUpper("Y"), strings.ToUpper("Y")) {
			return
		}

		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Restart(); err != nil {
			fn.PrintError(err)
			return
		}

	},
}
