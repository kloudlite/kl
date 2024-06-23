package intercept

import (
	"errors"
	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/server"
	"github.com/kloudlite/kl2/utils"
	"github.com/kloudlite/kl2/utils/klfile"
	"github.com/spf13/cobra"
	"os"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start tunneling the traffic to your device",
	Long: `start intercept app to tunnel trafic to your device
Examples:
	# intercept app with selected vpn device
  kl intercept start
`,
	Run: func(cmd *cobra.Command, _ []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		env, err := utils.EnvAtPath(cwd)
		if err != nil {
			fn.PrintError(err)
			return
		}
		klFile, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}
		if len(klFile.Ports) == 0 {
			fn.PrintError(errors.New("no ports exposed"))
			return
		}

		apps, err := server.ListApps([]fn.Option{
			fn.MakeOption("envName", string(env)),
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

		err = server.InterceptApp(true, klFile.Ports, app, []fn.Option{
			fn.MakeOption("envName", string(env)),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("intercept app started successfully\n")
		fn.Log("Please check if vpn is connected to your device, if not please connect it using sudo kl vpn start. Ignore this message if already connected.")
	},
}
