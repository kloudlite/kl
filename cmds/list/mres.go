package list

import (
	"errors"
	"fmt"
	"os"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/table"
	"github.com/kloudlite/kl2/server"
	"github.com/kloudlite/kl2/utils/klfile"
	"github.com/spf13/cobra"
)

var mresCmd = &cobra.Command{
	Use:   "mreses",
	Short: "Get list of managed resources in selected environment",
	Run: func(cmd *cobra.Command, args []string) {

		cwd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		env, err := server.EnvAtPath(cwd)
		if err != nil {
			return
		}

		klFile, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			fn.PrintError(errors.New("please run 'kl init' if you are not initialized the file already"))
			return
		}

		sec, err := server.ListMreses([]fn.Option{
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printMres(cmd, sec); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printMres(_ *cobra.Command, secrets []server.Mres) error {
	if len(secrets) == 0 {
		return errors.New("no secrets found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		// table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range secrets {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(secrets), true)
	return nil
}

func init() {
	mresCmd.Aliases = append(mresCmd.Aliases, "mres")

}
