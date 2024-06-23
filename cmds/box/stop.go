package box

import (
	"os"

	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop devbox",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		oswd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		devbox.Stop(oswd)
	},
}
