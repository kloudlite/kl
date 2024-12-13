package pkg

import (
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/nixpkghandler"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove installed package",
	Run: func(cmd *cobra.Command, args []string) {

		if err := func() error {
			name := fn.ParseStringFlag(cmd, "name")
			if name == "" && len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				return fmt.Errorf("name is required")
			}

			pc, err := nixpkghandler.New(cmd)
			if err != nil {
				return err
			}

			if err := pc.RemovePackage(name); err != nil {
				return err
			}

			fn.Logf("package %s removed successfully\n", name)
			return nil
		}(); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func init() {
	rmCmd.Flags().StringP("name", "n", "", "name of the package to remove")
}
