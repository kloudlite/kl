package box

import (
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "info of a container",
	Run: func(_ *cobra.Command, _ []string) {
		// Get current working directory
		oswd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err = devbox.ContainerInfo(oswd); err != nil {
			fn.PrintError(err)
		}
	},
}
