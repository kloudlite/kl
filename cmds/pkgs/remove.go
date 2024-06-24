package pkgs

import (
	"fmt"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove existing package",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}
		for i, p := range config.Packages {
			splits := strings.Split(p, "@")
			if splits[0] == args[0] {
				config.Packages = append(config.Packages[:i], config.Packages[i+1:]...)
				if err := klfile.WriteKLFile(*config); err != nil {
					fn.PrintError(err)
					return
				}
				fn.Println(fmt.Sprintf("Removed %s", args[0]))
				return
			}
		}
		fn.PrintError(fmt.Errorf("package %s not found", args[0]))
	},
}
