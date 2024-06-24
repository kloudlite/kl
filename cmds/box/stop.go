package box

import (
	"os"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop devbox",
	Run: func(*cobra.Command, []string) {
		// Get current working directory
		oswd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := devbox.Stop(oswd); err != nil {
			fn.PrintError(err)
		}
	},
}
