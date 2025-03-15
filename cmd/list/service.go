package list

import (
	"strconv"
	"strings"

	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/spf13/cobra"
)

var serviesCmd = &cobra.Command{
	Use:   "services",
	Short: "get list of services in current environment",
	Run: func(cmd *cobra.Command, args []string) {
		apic := clients.Api

		if err := listServices(apic, cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listServices(apic apiclient.ApiClient, cmd *cobra.Command, _ []string) error {
	fc := clients.File

	currentTeamName, err := fc.GetDirTeam()
	if err != nil {
		return functions.NewE(err)
	}

	currentEnvName, err := apic.EnsureEnv()
	if err != nil {
		return functions.NewE(err)
	}

	services, err := apic.ListServices(currentTeamName, currentEnvName)
	if err != nil {
		return functions.NewE(err)
	}

	if len(services) == 0 {
		return fn.Errorf("[#] no services found in environemnt: %s", text.Blue(currentEnvName))
	}

	header := table.Row{
		table.HeaderText("Service Name"),
		table.HeaderText("Ip"),
		table.HeaderText("Port"),
	}

	rows := make([]table.Row, 0)

	ports := make([]string, 0)
	for _, a := range services {
		ports = nil
		for _, v := range a.Spec.Ports {
			ports = append(ports, strconv.Itoa(v.Port))
		}

		rows = append(rows, table.Row{a.Spec.ServiceRef.Name, a.Metadata.Name, strings.Join(ports, ", ")})
	}

	fn.Println(table.Table(&header, rows, cmd))

	table.KVOutput("services of environment: ", currentEnvName, true)
	table.TotalResults(len(services), true)
	return nil
}

func init() {
	serviesCmd.Aliases = append(serviesCmd.Aliases, "service")
	serviesCmd.Aliases = append(serviesCmd.Aliases, "svc")
}
