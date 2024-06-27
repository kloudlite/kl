package auth

import (
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(*cobra.Command, []string) {
		configFolder, err := client.GetConfigFolder()
		if err != nil {
			fn.Log(err)
			return
		}

		if err = client.Logout(configFolder); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

