package list

import (
	"os"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Get list of apps in selected environment",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listapps(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listapps(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return functions.Error(err)
	}
	env, err := server.EnvAtPath(cwd)
	if err != nil {
		return functions.Error(err)
	}

	klFile, err := klfile.GetKlFile("")
	if err != nil {
		fn.PrintError(err)
		return functions.Error(err, "please run 'kl init' if you are not initialized the file already")
	}

	apps, err := server.ListApps([]fn.Option{
		fn.MakeOption("accountName", klFile.AccountName),
		fn.MakeOption("envName", env.Name),
	}...)
	if err != nil {
		return functions.Error(err)
	}

	if len(apps) == 0 {
		return functions.Error(err, "no apps found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
	}

	rows := make([]table.Row, 0)

	for _, a := range apps {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name})
	}

	fn.Println(table.Table(&header, rows, cmd))

	table.KVOutput("apps of", env.Name, true)
	table.TotalResults(len(apps), true)
	return nil
}

func init() {
	appsCmd.Aliases = append(appsCmd.Aliases, "app")
}
