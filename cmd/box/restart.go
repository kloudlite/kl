package box

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"os"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart running container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := restartBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
		return
	},
}

func restartBox(cmd *cobra.Command, args []string) error {
	cont, err := getRunningContainer()
	if err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	if dir == cont.Path && cont.Path != "" {
		if err := stopBox(cmd, args); err != nil {
			return err
		}
	}
	if err := startBox(cmd, args); err != nil {
		return err
	}
	return nil
}

func init() {
	restartCmd.Aliases = append(restartCmd.Aliases, "rs")
}
