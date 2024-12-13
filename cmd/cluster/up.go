package cluster

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/k3s"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Starts the k3s server",
	Long:  `Starts the k3s server`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := startK3sServer(cmd); err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func startK3sServer(cmd *cobra.Command) error {
	defer spinner.Client.UpdateMessage("starting k3s server")()

	fc := clients.File
	if err != nil {
		return functions.NewE(err)
	}

	apic := clients.Api
	if err != nil {
		return functions.NewE(err)
	}

	k, err := k3s.NewClient(cmd)
	if err != nil {
		return functions.NewE(err)
	}

	teamName, err := k.CheckK3sServerRunning()
	if err != nil {
		return functions.NewE(err)
	}

	sd, err := fc.GetSessionData()
	if err != nil {
		return err
	}

	if (err != nil && os.IsNotExist(err)) || sd.Team == "" {
		currentTeam, err := fc.CurrentTeamName()
		if err != nil {
			return functions.NewE(err)
		}
		sd.Team = currentTeam
		if err := sd.Save(); err != nil {
			return functions.NewE(err)
		}
	} else if err != nil {
		return functions.NewE(err)
	}

	if sd.Team == "" {

		teams, err := apic.ListTeams()
		if err != nil {
			return functions.NewE(err)
		}

		var selectedTeam *apiclient.Team

		if len(teams) == 0 {
			return functions.Error("no teams found")
		} else if len(teams) == 1 {
			selectedTeam = &teams[0]
		} else {
			selectedTeam, err = fzf.FindOne(teams, func(item apiclient.Team) string {
				return item.Metadata.Name
			}, fzf.WithPrompt("Select team to use >"))
			if err != nil {
				return err
			}
		}

		//if selectedTeam.Metadata.Name != extraData.SelectedTeam && extraData.SelectedTeam != "" {
		//	if err := StopK3sServer(cmd); err != nil {
		//		return functions.NewE(err)
		//	}
		//}

		sd.Team = selectedTeam.Metadata.Name

		if err := sd.Save(); err != nil {
			return functions.NewE(err)
		}

	}

	if sd.Team != teamName && teamName != "" && sd.Team != "" {
		functions.Logf(text.Yellow(fmt.Sprintf("[#] local cluster is already running for team %s, do you want to stop it and start a new cluster for team %s? [y/N] ", teamName, sd.Team)))
		if !functions.Confirm("Y", "N") {
			return nil
		}
		if err := StopK3sServer(cmd); err != nil {
			return functions.NewE(err)
		}
	}

	_, err = apic.GetClusterConfig(sd.Team)
	if err != nil {
		return err
	}

	if err = k.CreateClustersTeams(sd.Team); err != nil {
		return functions.NewE(err)
	}
	functions.Log("k3s server started. It will usually take a minute to come online")
	return nil
}
