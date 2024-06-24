package box

import (
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start new devbox",
	Run: func(_ *cobra.Command, _ []string) {
		oswd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err = devbox.Start(oswd); err != nil {
			fn.PrintError(err)
			if err := devbox.Stop(oswd); err != nil {
				fn.PrintError(err)
			}
		}
	},
}
