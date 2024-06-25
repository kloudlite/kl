package intercept

import (
	"errors"
	"os"
	"slices"
	"strconv"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
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
		maps, err := cmd.Flags().GetStringArray("port")
		if err != nil {
			fn.PrintError(err)
			return
		}

		ports := make([]server.AppPort, 0)

		for _, v := range maps {
			mp := strings.Split(v, ":")
			if len(mp) != 2 {
				fn.PrintError(
					errors.New("wrong map format use <server_port>:<local_port> eg: 80:3000"),
				)
				return
			}

			pp, err := strconv.ParseInt(mp[0], 10, 32)
			if err != nil {
				fn.PrintError(err)
				return
			}

			tp, err := strconv.ParseInt(mp[1], 10, 32)
			if err != nil {
				fn.PrintError(err)
				return
			}

			ports = append(ports, server.AppPort{
				AppPort:    int(pp),
				DevicePort: int(tp),
			})
		}

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

		app, err := EnsureApp(apps)
		if err != nil {
			fn.PrintError(err)
			return
		}

		vpnConfig, err := devbox.GetAccVPNConfig(klFile.AccountName)
		if err != nil {
			fn.PrintError(err)
			return
		}

		err = server.InterceptApp(true, ports, vpnConfig.DeviceName, app, []fn.Option{
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		containerWorkspacePath := cwd
		if val, ok := os.LookupEnv("KL_WORKSPACE"); ok {
			containerWorkspacePath = val
		}

		for _, ap := range ports {
			if !slices.Contains(klFile.Ports, ap.DevicePort) {
				klFile.Ports = append(klFile.Ports, ap.DevicePort)
			}
		}

		if err = devbox.SyncProxy(devbox.ProxyConfig{
			ExposedPorts:        klFile.Ports,
			TargetContainerPath: containerWorkspacePath,
		}); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("intercept app started successfully\n")
		fn.Log("Please check if vpn is connected to your device, if not please connect it using sudo kl vpn start. Ignore this message if already connected.")
	},
}

func init() {
	startCmd.Flags().StringArrayP(
		"port", "p", []string{},
		"expose port <server_port>:<local_port> while intercepting app",
	)

	startCmd.Aliases = append(startCmd.Aliases, "add", "begin", "connect")
}
