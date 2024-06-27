package auth

import (
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(*cobra.Command, []string) {
		configFolder, err := fileclient.GetConfigFolder()
		if err != nil {
			fn.Log(err)
			return
		}

		if err = fileclient.Logout(configFolder); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

