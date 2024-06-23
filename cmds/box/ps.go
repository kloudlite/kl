package box

import (
	"fmt"

	"github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "list all running boxes",
	Run: func(cmd *cobra.Command, args []string) {
		cs, err := devbox.AllWorkspaceContainers()
		if err != nil {
			functions.PrintError(err)
		}
		// Print CS
		fmt.Println(cs)
	},
}
