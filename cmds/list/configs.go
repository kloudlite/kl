package list

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl2/utils"
	"os"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/table"
	"github.com/kloudlite/kl2/pkg/ui/text"
	"github.com/kloudlite/kl2/server"
	"github.com/kloudlite/kl2/utils/klfile"
	"github.com/spf13/cobra"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "Get list of configs in selected environment",
	Run: func(cmd *cobra.Command, args []string) {

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
			fn.PrintError(errors.New("please run 'kl init' if you are not initialized the file already"))
			return
		}

		config, err := server.ListConfigs([]fn.Option{
			fn.MakeOption("envName", string(env)),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)

		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printConfigs(cmd, config, string(env)); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printConfigs(cmd *cobra.Command, configs []server.Config, envName string) error {

	if len(configs) == 0 {
		return fmt.Errorf("[#] no configs found in environemnt: %s", text.Blue(envName))
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range configs {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name, fmt.Sprintf("%d", len(a.Data))})
	}

	fmt.Println(table.Table(&header, rows, cmd))

	table.TotalResults(len(configs), true)
	return nil
}
func init() {
	configsCmd.Aliases = append(configsCmd.Aliases, "config")
}
