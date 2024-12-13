package list

import (
	"fmt"

	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	fn "github.com/kloudlite/kl/pkg/functions"

	// fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Get list of secrets in selected environment",
	Run: func(cmd *cobra.Command, args []string) {
		fc := clients.File
		apic := clients.Api

		currentTeam, err := fc.GetDataContext().GetWsTeam()
		if err != nil {
			fn.PrintError(err)
			return
		}
		currentEnv, err := apic.EnsureEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		sec, err := apic.ListSecrets(currentTeam, currentEnv)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printSecrets(apic, cmd, sec, currentEnv); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printSecrets(apic apiclient.ApiClient, cmd *cobra.Command, secrets []apiclient.Secret, currentEnv string) error {
	e, err := apic.EnsureEnv()
	if err != nil {
		return fn.NewE(err)
	}

	if len(secrets) == 0 {
		return fn.Errorf("[#] no secrets found in environemnt: %s", text.Blue(e))
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range secrets {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name, fmt.Sprintf("%d", len(a.StringData))})
	}

	fn.Println(table.Table(&header, rows))
	table.KVOutput("secrets of environment: ", currentEnv, true)
	table.TotalResults(len(secrets), true)
	return nil
}

func init() {
	secretsCmd.Aliases = append(secretsCmd.Aliases, "secret")
	secretsCmd.Aliases = append(secretsCmd.Aliases, "sec")
}
