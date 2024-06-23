package box

import (
	"os"

	"github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start new devbox",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		oswd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		err = devbox.Start(oswd)
		if err != nil {
			functions.PrintError(err)
		}
	},
}
