package list

import (
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var mresCmd = &cobra.Command{
	Use:   "mreses",
	Short: "Get list of managed resources in selected environment",
	Run: func(cmd *cobra.Command, args []string) {
		fc := clients.File
		apic := clients.Api

		currentEnv, err := apic.EnsureEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		currentTeam, err := fc.GetDirTeam()
		if err != nil {
			fn.PrintError(err)
			return
		}

		mres, err := apic.ListMreses(currentTeam, currentEnv)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printMres(apic, cmd, mres, currentEnv); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printMres(apic apiclient.ApiClient, cmd *cobra.Command, mres []apiclient.Mres, currentEnv string) error {
	e, err := apic.EnsureEnv()
	if err != nil {
		return fn.NewE(err)
	}
	if len(mres) == 0 {
		return fn.Errorf("[#] no managed resources found in environemnt: %s", text.Blue(e))
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("Secret Ref Name"),
	}

	rows := make([]table.Row, 0)

	for _, a := range mres {
		rows = append(rows, table.Row{a.DisplayName, a.Name, a.SecretRefName.Name})
	}

	fn.Println(table.Table(&header, rows))
	table.KVOutput("managed resources of environment: ", currentEnv, true)
	table.TotalResults(len(mres), true)
	return nil
}

func init() {
	mresCmd.Aliases = append(mresCmd.Aliases, "mres", "managed-resources", "res")
	fn.WithOutputVariant(mresCmd)
}
