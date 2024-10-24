package status

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/cmd/connect"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/k3s"
	"os"
	"time"

	"github.com/kloudlite/kl/domain/apiclient"
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
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if u, err := apic.GetCurrentUser(); err == nil {
			fn.Logf("\nLogged in as %s (%s)\n",
				text.Blue(u.Name),
				text.Blue(u.Email),
			)
		}

		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		k3sClient, err := k3s.NewClient()
		if err != nil {
			fn.PrintError(err)
			return
		}

		data, err := fileclient.GetExtraData()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Team: ")), data.SelectedTeam))
		}

		e, err := apic.EnsureEnv()
		if err == nil {
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), e.Name))
		} else if errors.Is(err, fileclient.NoEnvSelected) {
			filePath := fn.ParseKlFile(cmd)
			klFile, err := fc.GetKlFile(filePath)
			if err != nil {
				fn.PrintError(err)
				return
			}
			fn.Log(fmt.Sprint(text.Bold(text.Blue("Environment: ")), klFile.DefaultEnv))
		}
		fn.Log()
		fn.Log(text.Bold("Cluster Status"))

		config, err := fc.GetClusterConfig(data.SelectedTeam)
		if err != nil {
			if os.IsNotExist(err) {
				fn.PrintError(fn.Error("no cluster found"))
				return
			} else {
				fn.PrintError(err)
				return
			}
		}
		fn.Log("Name: ", text.Blue(config.ClusterName))

		k3sStatus, _ := k3sClient.CheckK3sRunningLocally()
		if k3sStatus {
			fn.Log("Running: ", text.Green("true"))
		} else {
			fn.Log("Running ", text.Yellow("false"))
		}

		k3sTracker, err := fc.GetK3sTracker()
		if err != nil {
			if flags.IsVerbose {
				fn.PrintError(err)
			}
			fn.Log("Local Cluster: ", text.Yellow("not ready"))
			fn.Log("Edge Connection:", text.Yellow("offline"))
		} else {
			err = getClusterK3sStatus(k3sTracker)
			if err != nil {
				if flags.IsVerbose {
					fn.PrintError(err)
				}
				fn.Log("Local Cluster: ", text.Yellow("not ready"))
				fn.Log("Edge Connection:", text.Yellow("offline"))
			}
		}

		if envclient.InsideBox() {
			fn.Log(text.Bold("\nWorkspace Status"))
			env, _ := fc.CurrentEnv()
			fn.Log("Current Environment: ", text.Blue(env.Name))

			if connect.ChekcWireguardConnection() {
				fn.Log("Edge Connection:", text.Green("online"))
			} else {
				fn.Log("Edge Connection:", text.Yellow("offline"))
			}
		}
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
