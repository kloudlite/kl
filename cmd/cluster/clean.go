package cluster

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/clone"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
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

	currentTeam, err := fc.CurrentTeamName()
	if err != nil {
		return fn.NewE(err)
	}

	_, err = clone.SelectCluster(apic, fc)
	if err != nil {
		return fn.NewE(err)
	}

	fn.Printf(text.Yellow("this will delete cluster and all its data and volumes. Do you want to continue? (y/N): "))
	if !fn.Confirm("Y", "N") {
		return nil
	}

	_, err = apic.GetClustersOfTeam(currentTeam)
	if err != nil {
		return fn.NewE(err)
	}

	// TODO: delete cluster api call

	// TODO: remove data from k3s-local

	// TODO: remove docker volumes

	return nil

}
