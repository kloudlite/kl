package cluster

import (
	"github.com/kloudlite/kl/cmd/cluster/clusterpkg"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:     "connect",
	Short:   "connect to a cluster",
	Long:    `This command is used to connect to a cluster,`,
	Example: `kl cluster connect`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := connectCluster(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func connectCluster() error {
	defer spinner.Client.UpdateMessage("starting k3s cluster")()
	fc, err := fileclient.New()
	if err != nil {
		return fn.NewE(err)
	}
	apic, err := apiclient.New()
	if err != nil {
		return fn.NewE(err)
	}
	clusterClient, err := clusterpkg.New(fc, apic)
	if err != nil {
		return fn.NewE(err)
	}
	if err := clusterClient.StartClusterForAccount(); err != nil {
		return fn.NewE(err)
	}
	fn.Log("cluster connected")
	return nil
}

func init() {
	connectCmd.Aliases = []string{"conn", "start"}
}
