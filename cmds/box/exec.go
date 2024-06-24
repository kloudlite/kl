package box

import (
	"os"

	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec in devbox",
	Run: func(_ *cobra.Command, args []string) {
		oswd, err := os.Getwd()
		if err != nil {
			functions.PrintError(err)
			return
		}

		if err := devbox.Start(oswd); err != nil {
			functions.PrintError(err)
			return
		}
		if _, err := devbox.Exec(oswd, args, nil); err != nil {
			functions.PrintError(err)
			return
		}
	},
}
