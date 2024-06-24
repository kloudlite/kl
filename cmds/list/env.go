package list

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "envs",
	Short: "Get list of environments",
	Run: func(cmd *cobra.Command, args []string) {
		err := listEnvironments(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listEnvironments(cmd *cobra.Command, args []string) error {

	var err error

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

	envs, err := server.ListEnvs([]fn.Option{
		fn.MakeOption("accountName", klFile.AccountName),
	}...)
	if err != nil {
		return functions.Error(err)
	}

	if len(envs) == 0 {
		return functions.Error(err, "no environments found")
	}

	header := table.Row{table.HeaderText("DisplayName"), table.HeaderText("Name"), table.HeaderText("ready")}
	rows := make([]table.Row, 0)

	for _, a := range envs {
		rows = append(rows, table.Row{
			fn.GetPrintRow(a, env.Name, a.DisplayName, true),
			fn.GetPrintRow(a, env.Name, a.Metadata.Name),
			fn.GetPrintRow(a, env.Name, a.Status.IsReady),
		})
	}

	fmt.Println(table.Table(&header, rows))

	if s := fn.ParseStringFlag(cmd, "output"); s == "table" {
		table.TotalResults(len(envs), true)
	}
	table.TotalResults(len(envs), true)
	return nil
}

func init() {
	envCmd.Aliases = append(envCmd.Aliases, "env")
}
