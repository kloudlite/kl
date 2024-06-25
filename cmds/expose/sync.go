package expose

import (
	"os"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync ports",
	Long: `
This command will sync ports to your kl-config file.
`,
	Example: ` 
  kl expose sync
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
		return functions.Error(err)
	}
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return functions.Error(err, "please run 'kl init' if you are not initialized the file already")
	}
	containerWorkspacePath := cwd
	if val, ok := os.LookupEnv("KL_WORKSPACE"); ok {
		containerWorkspacePath = val
	}

	if err = devbox.SyncProxy(devbox.ProxyConfig{
		ExposedPorts:        klFile.Ports,
		TargetContainerPath: containerWorkspacePath,
	}); err != nil {
		return fn.Error(err)
	}
	return nil
}
