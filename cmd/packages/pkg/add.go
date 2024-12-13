package pkg

import (
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/nixpkghandler"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add new package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := addPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func addPackages(cmd *cobra.Command, args []string) error {
	name := fn.ParseStringFlag(cmd, "name")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return fn.Error("name is required")
	}

	pc, err := nixpkghandler.New(cmd)
	if err != nil {
		return fn.NewE(err)
	}

	name, hashpkg, err := pc.Find(name)
	if err != nil {
		return fn.NewE(err)
	}

	// download and update lockfile
	if err := pc.AddPackage(name, hashpkg); err != nil {
		return fn.NewE(err)
	}

	fn.Println(fmt.Sprintf("Package %s is added successfully", name))
	return nil
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "name of the package to install")
}
