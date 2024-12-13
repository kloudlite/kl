package list

import (
	"fmt"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var accCmd = &cobra.Command{
	Use:   "teams",
	Short: "Get list of teams accessible to you",
	Run: func(cmd *cobra.Command, _ []string) {
		apic := clients.Api
		if err := listTeams(apic, cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listTeams(apic apiclient.ApiClient, cmd *cobra.Command) error {
	teams, err := apic.ListTeams()

	if err != nil {
		return functions.NewE(err)
	}

	if len(teams) == 0 {
		return fn.Errorf("[#] no teams found")
	}

	// this erro ignore is intentional
	teamName, _ := apic.GetFClient().GetDataContext().GetTeam()

	header := table.Row{table.HeaderText("name"), table.HeaderText("id")}
	rows := make([]table.Row, 0)

	for _, a := range teams {
		rows = append(rows, table.Row{
			func() string {
				if a.Metadata.Name == teamName {
					return text.Colored(fmt.Sprint("*", a.DisplayName), 2)
				}
				return a.DisplayName
			}(),

			func() string {
				if a.Metadata.Name == teamName {
					return text.Colored(a.Metadata.Name, 2)
				}
				return a.Metadata.Name
			}(),
		})
	}

	fn.Println(table.Table(&header, rows, cmd))

	table.TotalResults(len(rows), true)
	return nil
}

func init() {
	accCmd.Aliases = append(accCmd.Aliases, "team")
}
