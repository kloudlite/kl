package auth

import (
	"os"

	fn "github.com/kloudlite/kl2/pkg/functions"
	utils "github.com/kloudlite/kl2/utils"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(*cobra.Command, []string) {
		configFolder, err := utils.GetConfigFolder()
		if err != nil {
			fn.Log(err)
			return
		}

		if err := os.RemoveAll(configFolder); err != nil {
			fn.Log(err)
			return
		}
	},
}
