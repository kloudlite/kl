package intercept

import (
	"os"

<<<<<<< HEAD
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils"
	"github.com/kloudlite/kl/utils/klfile"
=======
	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/server"
	"github.com/kloudlite/kl2/utils/klfile"
>>>>>>> exeperimental
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [app_name]",
	Short: "stop tunneling the traffic to your device",
	Long: `stop intercept app to stop tunnel traffic to your device
Examples:
	# close intercept app
  kl intercept stop [app_name]
	`,

	Run: func(cmd *cobra.Command, _ []string) {

		cwd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		env, err := server.EnvAtPath(cwd)
		if err != nil {
			fn.PrintError(err)
			return
		}

		klFile, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}

		apps, err := server.ListApps([]fn.Option{
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		app, err := EnsuseApp(apps)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := server.InterceptApp(false, nil, app, []fn.Option{
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("intercepted app stopped successfully")
	},
}