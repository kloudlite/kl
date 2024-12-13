package pkg

import (
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list installed packages",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listPackages(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listPackages(cmd *cobra.Command, _ []string) error {
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
		table.HeaderText("packages"),
		table.HeaderText("nixpkgs"),
	}

	rows := make([]table.Row, 0)

	for _, v := range kt.Packages {
		rows = append(rows, table.Row{v, l.Packages[v]})
	}

	fn.Println(table.Table(&header, rows, cmd))

	table.TotalResults(len(kt.Packages), true)
	return nil
}

func init() {
	listCmd.Aliases = append(listCmd.Aliases, "ls")
	fn.WithOutputVariant(listCmd)
}
