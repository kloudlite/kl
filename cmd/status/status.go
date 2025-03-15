package status

import (
	"fmt"
	"time"

	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

const (
	K3sServerNotReady = "k3s server is not ready, please wait"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "get status of your current context (user, team, environment, vpn status)",
	Run: func(cmd *cobra.Command, _ []string) {
		apic := clients.Api
		fc := clients.File

		if u, err := apic.GetCurrentUser(); err == nil {
			fn.Logf("\nLogged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
		}

		var team string
		team, err := fc.GetDirTeam()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Team: ")), team))
		}

		func() {
			if team == "" {
				return
			}

			e, err := apic.EnsureEnv()
			selectedEnv := ""
			if err == nil {
				selectedEnv = e
			}

			ev, err := apic.GetEnvironment(team, selectedEnv)
			if err == nil {
				r := text.Yellow("offline")
				if ev.ClusterName != "" {
					if ev.IsArchived {
						r = text.Yellow("archived")
					} else {
						cluster, err := apic.GetCluster(team, ev.ClusterName)
						if err != nil {
							fn.PrintError(err)
							return
						}
						if time.Since(cluster.LastOnlineAt) < time.Minute {
							r = text.Green("online")
						}
						if ev.Spec.Suspend {
							r = text.Yellow("suspended")
						}
					}
					fn.Log(text.Bold(text.Blue("Environment: ")), selectedEnv, fmt.Sprintf("(%s)", r))
				}
			} else if selectedEnv != "" {
				fn.Log(text.Bold(text.Blue("Environment: ")), selectedEnv)
			}
		}()

		return

		// k3sClient, err := k3s.NewClient(cmd)
		// if err != nil {
		// 	fn.PrintError(err)
		// 	return
		// }

		// func() {
		// 	if team == "" {
		// 		return
		// 	}
		//
		// 	fn.Log(text.Bold("\nCluster Status"))
		// 	config, err := fc.GetClusterConfig(team)
		// 	if err == nil {
		// 		fn.Log("Name: ", text.Blue(config.ClusterName))
		//
		// 		k3sStatus, _ := k3sClient.CheckK3sRunningLocally()
		// 		if k3sStatus {
		// 			fn.Log("Running: ", text.Green("true"))
		// 		} else {
		// 			fn.Log("Running ", text.Yellow("false"))
		// 		}
		//
		// 		k3sTracker, err := fc.GetK3sTracker()
		// 		if err != nil {
		// 			if flags.IsVerbose {
		// 				fn.PrintError(err)
		// 			}
		// 			fn.Log("Local Cluster: ", text.Yellow("not ready"))
		// 			fn.Log("Edge Connection:", text.Yellow("offline"))
		// 		} else {
		// 			err = getClusterK3sStatus(k3sTracker)
		// 			if err != nil {
		// 				if flags.IsVerbose {
		// 					fn.PrintError(err)
		// 				}
		// 				fn.Log("Local Cluster: ", text.Yellow("not ready"))
		// 				fn.Log("Edge Connection:", text.Yellow("offline"))
		// 			}
		// 		}
		// 	}
		// 	if err != nil {
		// 		if os.IsNotExist(err) {
		// 			fn.Log(text.Yellow("cluster not found"))
		// 		} else {
		// 			fn.PrintError(err)
		// 		}
		// 	}
		// }()

	},
}

func getClusterK3sStatus(k3sTracker *fileclient.K3sTracker) error {

	lastCheckedAt, err := time.Parse(time.RFC3339, k3sTracker.LastCheckedAt)
	if err != nil {
		return err
	}

	if time.Since(lastCheckedAt) > 4*time.Second {
		return fn.Error(K3sServerNotReady)
	}

	if k3sTracker.Compute && k3sTracker.Gateway {
		fn.Log("Local Cluster: ", text.Green("ready"))
	} else {
		fn.Log("Local Cluster: ", text.Yellow("not ready"))
	}

	if k3sTracker.WgConnection {
		fn.Log("Edge Connection:", text.Green("online"))
		return nil
	}
	fn.Log("Edge Connection:", text.Yellow("offline"))

	return nil
}
