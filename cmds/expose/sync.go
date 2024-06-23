package expose

import (
	"os"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "port",
	Short: "expose ports",
	Long: `
This command will add ports to your kl-config file.
`,
	Example: ` 
  kl expose ports 8080 3000
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := sync(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func sync() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	devbox.Start(cwd)
	return nil
}
