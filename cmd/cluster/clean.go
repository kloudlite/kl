package cluster

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/clone"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
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
			fmt.Println(err)
			return
		}
	},
}

func cleanCluster(cmd *cobra.Command) error {
	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	apic, err := apiclient.New()
	if err != nil {
		return err
	}

	k3sClient, err := k3s.NewClient()
	if err != nil {
		return err
	}

	currentTeam, err := fc.CurrentTeamName()
	if err != nil {
		return fn.NewE(err)
	}

	selectedCluster, err := clone.SelectCluster(apic, fc)
	if err != nil {
		return fn.NewE(err)
	}

	fn.Printf(text.Yellow("this will delete cluster and all its data and volumes. Do you want to continue? (y/N): "))
	if !fn.Confirm("Y", "N") {
		return nil
	}

	//currentCluster, err := apic.GetClustersOfTeam(currentTeam)
	//if err != nil {
	//	return fn.NewE(err)
	//}

	fmt.Println(selectedCluster.Metadata.Name)
	// TODO: delete cluster api call

	if err = fc.DeleteClusterData(currentTeam); err != nil {
		return fn.NewE(err)
	}

	if err = k3sClient.RemoveClusterVolume(cmd, selectedCluster.Metadata.Name); err != nil {
		return fn.NewE(err)
	}

	fn.Log(fmt.Sprintf("cluster %s deleted ", selectedCluster.Metadata.Name))
	return nil

}
