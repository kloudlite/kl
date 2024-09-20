package clusterpkg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"text/template"
)

const (
	CONT_MARK_KEY = "kl.container"
)

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

func (c *clusterclient) StartClusterForAccount() error {
	device, err := c.fc.GetDevice()
	if err != nil {
		return fn.NewE(err)
	}

	existingOtherContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-device", device.DeviceName)),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}

	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-account", c.account)),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-device", device.DeviceName)),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}

	if existingOtherContainers != nil && len(existingOtherContainers) > 0 && existingContainers[0].ID != existingOtherContainers[0].ID {
		fn.Logf(text.Yellow(fmt.Sprintf("[#] another cluster is running for another account. do you want to stop it and start cluster for account %s? [Y/n] ", c.account)))
		if !fn.Confirm("Y", "Y") {
			return nil
		}
		if err := c.cli.ContainerStop(context.Background(), existingOtherContainers[0].ID, container.StopOptions{}); err != nil {
			return fn.Error("failed to stop container")
		}
	}

	if existingContainers != nil && len(existingContainers) > 0 {
		if existingContainers[0].State == "running" {
			if err := c.cli.ContainerStop(context.Background(), existingContainers[0].ID, container.StopOptions{}); err != nil {
				return fn.Error("failed to stop container")
			}
		}
		if err := c.cli.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{}); err != nil {
			return fn.Error("failed to start container")
		}
		if err := c.ExecuteClusterScript(existingContainers[0].ID); err != nil {
			return fn.Error("failed to execute cluster script")
		}

		return nil
	}

	createdConatiner, err := c.cli.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			CONT_MARK_KEY: "true",
			"kl-k3s":      "true",
			"kl-account":  c.account,
			"kl-device":   device.DeviceName,
		},
		Image: constants.GetK3SImageName(),
		Cmd: []string{
			"server",
			"--tls-san",
			"0.0.0.0",
			"--tls-san",
			fmt.Sprintf("%s.kcluster.local.khost.dev", c.account),
		},
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: "host",
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds: []string{
			fmt.Sprintf("kl-k3s-%s-cache:/var/lib/rancher/k3s", c.account),
			//"kl-k3s-cache:/var/lib/rancher/k3s",
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return fn.Error("failed to create container")
	}
	if err := c.cli.ContainerStart(context.Background(), createdConatiner.ID, container.StartOptions{}); err != nil {
		return fn.Error("failed to start container")
	}
	if err := c.ExecuteClusterScript(createdConatiner.ID); err != nil {
		return fn.Error("failed to execute cluster script")
	}
	return nil
}

func (c *clusterclient) ExecuteClusterScript(conatinerId string) error {
	clusterConfig, err := c.apic.GetClusterConfig(c.account)
	if err != nil {
		return fn.NewE(err)
	}

	savedClusterConfig, err := c.fc.GetClusterConfig(c.account)
	if err != nil {
		return fn.NewE(err)
	}

	if clusterConfig.ClusterToken != savedClusterConfig.ClusterToken {
		if err := c.fc.SetClusterConfig(c.account, clusterConfig); err != nil {
			return fn.NewE(err)
		}
	}

	script, err := generateConnectionScript(clusterConfig)
	if err != nil {
		return fn.Error("failed to generate connection script")
	}
	execConfig := container.ExecOptions{
		Cmd: []string{"sh", "-c", script},
	}

	resp, err := c.cli.ContainerExecCreate(context.Background(), conatinerId, execConfig)
	if err != nil {
		return fn.Error("failed to create exec")
	}

	err = c.cli.ContainerExecStart(context.Background(), resp.ID, container.ExecStartOptions{})
	if err != nil {
		return fn.Error("failed to start exec")
	}
	return nil
}
