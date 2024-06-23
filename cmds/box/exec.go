package box

import (
	"os"

	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec in devbox",
	Run: func(cmd *cobra.Command, args []string) {
		oswd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		devbox.Start(oswd)
		devbox.Exec(oswd, args, nil)
	},
}
