package kubectl

import (
	"fmt"
	"os"
	"os/exec"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var HelmCmd = &cobra.Command{
	Use:                "helm",
	Short:              "Helm is a package manager for Kubernetes",
	Hidden:             true,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {

		kconfPath, err := GetPath(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}

		c := exec.Command("helm", args...)
		c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kconfPath))

		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		if err := c.Run(); err != nil {
			err = nil
			return
		}
	},
}

func init() {
}
