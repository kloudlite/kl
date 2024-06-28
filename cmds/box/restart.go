package box

import (
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart running container of the current directory",
	Run: func(_ *cobra.Command, _ []string) {
		oswd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		klfile, err := klfile.GetKlFile(oswd)
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err = devbox.Restart(oswd, klfile); err != nil {
			fn.PrintError(err)
		}
	},
}