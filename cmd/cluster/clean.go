package cluster

import (
	"fmt"

	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "clean the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cleanCluster(cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func cleanCluster(cmd *cobra.Command) error {
	fc := clients.File
	k3sClient, err := k3s.NewClient(cmd)
	if err != nil {
		return err
	}

	team, err := fc.GetDirTeam()
	if err != nil {
		return fn.NewE(err)
	}

	fn.Printf(text.Yellow(fmt.Sprintf("this will delete k3s cluster for team %s and all its data and volumes. Do you want to continue? (y/N): ", team)))
	if !fn.Confirm("Y", "N") {
		return nil
	}

	if err = k3sClient.RemoveClusterVolume(team); err != nil {
		return fn.NewE(err)
	}
	return nil
}
