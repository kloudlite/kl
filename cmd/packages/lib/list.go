package lib

import (
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list installed libraries",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listLibs(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listLibs(cmd *cobra.Command, _ []string) error {
	fc := clients.File

	l, err := fc.GetLockfile()
	if err != nil {
		return err
	}

	kt, err := fc.GetKlFile()
	if err != nil {
		return functions.NewE(err)
	}

	header := table.Row{
		table.HeaderText("libraries"),
		table.HeaderText("nixpkgs"),
	}

	rows := make([]table.Row, 0)

	for _, v := range kt.Libraries {
		rows = append(rows, table.Row{v, l.Libraries[v]})
	}

	fn.Println(table.Table(&header, rows, cmd))

	table.TotalResults(len(kt.Libraries), true)
	return nil
}

func init() {
	listCmd.Aliases = append(listCmd.Aliases, "ls")
	fn.WithOutputVariant(listCmd)
}
