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

		s, err := fc.GetDataContext().GetEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Printf("%s", s)
	},
}

func init() {
}
