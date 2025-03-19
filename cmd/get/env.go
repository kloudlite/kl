package get

import (
	"fmt"

	"github.com/kloudlite/kl/domain/clients"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:    "env",
	Short:  "get current environment",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fc := clients.File

		ev, err := fc.DirEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Print(ev)
	},
}

func init() {
}
