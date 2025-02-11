package cluster

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	humanize "github.com/dustin/go-humanize"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/ui/table"

	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all clusters",
	Long:  "list all clusters",
	Run: func(cmd *cobra.Command, _ []string) {
		if err := listK3sClusters(cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func listK3sClusters(cmd *cobra.Command) error {
	defer spinner.Client.UpdateMessage("fetching clusters")()
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	crlist, err := cli.ContainerList(cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", k3s.CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", k3s.K3S_MARK_KEY, "true")),
		),
		All: true,
	})
	if err != nil {
		return err
	}

	if len(crlist) == 0 {
		return nil
	}

	teamName, _ := fileclient.File.GetDataContext().GetTeam()

	header := table.Row{table.HeaderText("Name"), table.HeaderText("Created"), table.HeaderText("Status")}
	rows := make([]table.Row, 0)
	for _, c := range crlist {
		res := apiclient.Team{
			Metadata: apiclient.Metadata{
				Name: c.Labels[k3s.TEAM_NAME_KEY],
			},
		}
		rows = append(rows, table.Row{
			fn.GetPrintRow(res, teamName, c.Labels[k3s.TEAM_NAME_KEY]),
			fn.GetPrintRow(res, teamName, humanize.Time(time.Unix(c.Created, 0))),
			fn.GetPrintRow(res, teamName, c.Status),
		})
	}

	fn.Println(table.Table(&header, rows, cmd))
	return nil
}

func init() {
	listCmd.Aliases = append(listCmd.Aliases, "ls")
}
