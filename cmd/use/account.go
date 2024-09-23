package use

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"text/template"
)

const (
	CONT_MARK_KEY = "kl-container"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "use account",
	Run: func(_ *cobra.Command, _ []string) {
		if err := useAccount(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func useAccount() error {
	apic, err := apiclient.New()
	if err != nil {
		return fn.NewE(err)
	}
	accounts, err := apic.ListAccounts()
	if err != nil {
		return fn.NewE(err)
	}
	selectedAccount, err := fzf.FindOne(accounts, func(item apiclient.Account) string {
		return item.Metadata.Name
	}, fzf.WithPrompt("Select account to use >"))
	if err != nil {
		return err
	}

	if err = CreateClustersAccounts(apic, selectedAccount.Metadata.Name); err != nil {
		return fn.NewE(err)
	}
	return nil
}

func CreateClustersAccounts(apic apiclient.ApiClient, accountName string) error {

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return fn.NewE(err)
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}

	if (existingContainers != nil) && (len(existingContainers) > 0) {
		for _, c := range existingContainers {
			if c.Labels["kl-account"] != accountName {
				fn.Log(text.Yellow(fmt.Sprintf("[#] another cluster is running for another account. do you want to stop it and start cluster for account %s? [Y/n] ", accountName)))
				if !fn.Confirm("Y", "Y") {
					return nil
				}
				if err := cli.ContainerStop(context.Background(), c.ID, container.StopOptions{}); err != nil {
					return fn.Error("failed to stop container")
				}
				if err := cli.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{}); err != nil {
					return fn.Error("failed to remove container")
				}
			}
		}
	}

	existingContainers, err = cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}

	if existingContainers != nil && len(existingContainers) > 0 {
		if existingContainers[0].State != "running" {
			if err := cli.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{}); err != nil {
				return fn.Error("failed to start container")
			}
		}
		return nil
	}

	clusterConfig, err := apic.GetClusterConfig(accountName)
	if err != nil {
		return fn.NewE(err)
	}

	createdConatiner, err := cli.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			CONT_MARK_KEY: "true",
			"kl-k3s":      "true",
			"kl-account":  accountName,
		},
		Image: constants.GetK3SImageName(),
		Cmd: []string{
			"server",
			"--disable", "traefik",
			"--node-name", clusterConfig.ClusterName,
			//fmt.Sprintf("%s.kcluster.local.khost.dev", account.Metadata.Name),
		},
	}, &container.HostConfig{
		Privileged: true,
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds: []string{
			fmt.Sprintf("kl-k3s-%s-cache:/var/lib/rancher/k3s", clusterConfig.ClusterName),
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return fn.Error("failed to create container")
	}

	if err := cli.ContainerStart(context.Background(), createdConatiner.ID, container.StartOptions{}); err != nil {
		return fn.Error("failed to start container")
	}

	script, err := generateConnectionScript(clusterConfig)
	if err != nil {
		return fn.Error("failed to generate connection script")
	}
	execConfig := container.ExecOptions{
		Cmd: []string{"sh", "-c", script},
	}

	resp, err := cli.ContainerExecCreate(context.Background(), createdConatiner.ID, execConfig)
	if err != nil {
		return fn.Error("failed to create exec")
	}

	err = cli.ContainerExecStart(context.Background(), resp.ID, container.ExecStartOptions{})
	if err != nil {
		return fn.Error("failed to start exec")
	}
	return nil
}

func generateConnectionScript(clusterConfig *fileclient.AccountClusterConfig) (string, error) {
	t := template.New("connectionScript")
	p, err := t.Parse(`
echo "checking whether k3s server is accepting connections"
while true; do
  lines=$(kubectl get nodes | wc -l)
  if [ "$lines" -lt 2 ]; then
	echo "k3s server is not accepting connections yet, retrying in 1s ..."
	sleep 1
	continue
  fi
  echo "successful, k3s server is now accepting connections"
  break
done
kubectl apply -f {{.InstallCommand.CRDsURL}} --server-side
kubectl create ns kloudlite
cat <<EOF | kubectl apply -f -
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: kloudlite
  namespace: kube-system
spec:
  repo: {{.InstallCommand.ChartRepo}}
  chart: kloudlite-agent
  version: {{.InstallCommand.ChartVersion}}
  targetNamespace: kloudlite
  valuesContent: |-
    accountName: {{.InstallCommand.HelmValues.AccountName}}
    clusterName: {{.InstallCommand.HelmValues.ClusterName}}
    clusterToken: {{.InstallCommand.HelmValues.ClusterToken}}
    kloudliteDNSSuffix: {{.InstallCommand.HelmValues.KloudliteDNSSuffix}}
    messageOfficeGRPCAddr: {{.InstallCommand.HelmValues.MessageOfficeGRPCAddr}}
EOF
`)
	if err != nil {
		return "", fn.NewE(err)
	}
	b := new(bytes.Buffer)
	err = p.Execute(b, clusterConfig)
	if err != nil {
		return "", fn.NewE(err)
	}
	return b.String(), nil
}
