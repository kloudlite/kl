package k3s

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func indent(text []byte, spaces int) string {
	pad := strings.Repeat(" ", spaces)
	lines := strings.Split(string(text), "\n")
	for i, line := range lines {
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}

func dockerLabelFilter(key, value string) filters.KeyValuePair {
	return filters.Arg("label", fmt.Sprintf("%s=%s", key, value))
}

func (k *K3sClientImpl) generateRemoteClusterName(userName, clusterName string) (string, error) {
	return fmt.Sprintf("%s-local-%s", strings.ReplaceAll(strings.ToLower(userName), " ", "-"), clusterName), nil
}

func (k *K3sClientImpl) CreateCluster(accName, name string) error {
	u, err := apiclient.GetCurrentUser()
	if err != nil {
		return err
	}
	remoteClusterName, err := k.generateRemoteClusterName(u.Name, name)
	if err != nil {
		return err
	}
	hostName, err := os.Hostname()
	if err != nil {
		hostName = fmt.Sprintf("%s's local", strings.Split(u.Name, " ")[0])
	}

	r, err := apiclient.CreateClusterReference(fmt.Sprintf("%s (%s)", name, hostName), remoteClusterName, functions.MakeOption("accountName", accName))
	if err != nil {
		return err
	}

	instructions, err := apiclient.GetClusterConnectionParams(r.Metadata.Name, functions.MakeOption("accountName", accName))
	if err != nil {
		return err
	}
	fmt.Println(instructions)

	err = k.fc.AddCluster(accName, name, r.Metadata.Name)
	if err != nil {
		return err
	}
	err = k.ensureImage(constants.K3SImage)
	if err != nil {
		return err
	}
	err = k.ensureVolume(name + "_var_lib_cni")
	if err != nil {
		return err
	}
	err = k.ensureVolume(name + "_var_lib_kubelet")
	if err != nil {
		return err
	}
	err = k.ensureVolume(name + "_var_lib_rancher_k3s")
	if err != nil {
		return err
	}
	err = k.ensureVolume(name + "_var_log")
	if err != nil {
		return err
	}
	err = k.start(accName, name)
	if err != nil {
		return err
	}
	err = k.setup(name, *instructions)
	if err != nil {
		return err
	}
	return nil
}

func (k *K3sClientImpl) stop(name string) error {
	spinner.Client.UpdateMessage(fmt.Sprintf("stopping cluster %s", name))
	data, err := k.dClient.ContainerInspect(context.Background(), "kl-cluster-"+name)
	if err != nil {
		if !errdefs.IsInvalidParameter(err) {
			return nil
		}
		return err
	}
	if !data.State.Running {
		return nil
	}
	err = k.dClient.ContainerKill(context.Background(), "kl-cluster-"+name, "SIGKILL")
	if err != nil {
		return err
	}
	err = k.dClient.ContainerRemove(context.Background(), "kl-cluster-"+name, container.RemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (k *K3sClientImpl) setup(name string, instructions apiclient.ClusterSetupInstructions) error {
	spinner.Client.UpdateMessage(fmt.Sprintf("setting up cluster %s", name))

	helmValuesYaml, err := yaml.Marshal(instructions.HelmValues)
	if err != nil {
		return fmt.Errorf("failed to marshal Helm values to YAML: %w", err)
	}

	indentedHelmValues := indent(helmValuesYaml, 4)

	formattedCommand := fmt.Sprintf(
		`
			while true; do
				if kubectl get nodes 2>/dev/null | grep -q ' Ready'; then
					break
				else
					echo "Waiting for cluster to be ready..."
					sleep 5
				fi
			done
			kubectl apply -f %s --server-side

		    cat <<EOF | kubectl apply -f -\
			apiVersion: v1
			kind: Namespace
			metadata:
			  name: kloudlite
       		---
			apiVersion: helm.cattle.io/v1
			kind: HelmChart
			metadata:
			  namespace: kloudlite
			  name: kloudlite
			spec:
			  targetNamespace: kloudlite
			  createNamespace: true
			  version: %s
			  chart: kloudlite
			  repo: %s
			  valuesContent: |-
%s
			EOF
		`, instructions.CRDSUrl, instructions.ChartVersion, instructions.ChartRepo, indentedHelmValues,
	)
	fmt.Println(formattedCommand)
	return nil

	exec, err := k.dClient.ContainerExecCreate(context.Background(), "kl-cluster-"+name, container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{"sh", "-c", fmt.Sprintf(
			`
			while true; do
				if kubectl get nodes 2>/dev/null | grep -q ' Ready'; then
					break
				else
					echo "Waiting for cluster to be ready..."
					sleep 5
				fi
			done
			kubectl apply -f %s --server-side

		    cat  <<EOF | kubectl apply -f -\
			apiVersion: v1
			kind: Namespace
			metadata:
			  name: kloudlite
       		---
			apiVersion: helm.cattle.io/v1
			kind: HelmChart
			metadata:
			  namespace: kloudlite
			  name: kloudlite
			spec:
			  targetNamespace:  kloudlite
			  createNamespace: true
			  version: %s
			  chart: kloudlite
			  repo: %s
			  valuesContent: |-
				%s
			EOF
		`, instructions.CRDSUrl, instructions.ChartVersion, instructions.ChartRepo, indentedHelmValues,
		)},
	})
	if err != nil {
		return err
	}
	resp, err := k.dClient.ContainerExecAttach(context.Background(), exec.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer resp.Close()
	if k.verbose {
		spinner.Client.Pause()
		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader)
		if err != nil {
			return err
		}
		spinner.Client.Resume()
	}
	return nil
}

func (k *K3sClientImpl) start(accName, name string) error {
	spinner.Client.UpdateMessage(fmt.Sprintf("starting cluster %s", name))
	data, err := k.dClient.ContainerInspect(context.Background(), "kl-cluster-"+name)
	if err == nil && data.State.Running {
		return nil
	}
	resp, err := k.dClient.ContainerCreate(context.Background(), &container.Config{
		Image: constants.K3SImage,
		Labels: map[string]string{
			"kloudlite.io/cluster":      "true",
			"kloudlite.io/cluster/name": name,
			"kloudlite.io/account/name": accName,
		},
		Cmd: []string{"server", "--tls-san=0.0.0.0"},
	}, &container.HostConfig{
		Privileged: true,
		Binds: []string{
			fmt.Sprintf("%s:/var/lib/cni", name+"_var_lib_cni"),
			fmt.Sprintf("%s:/var/lib/kubelet", name+"_var_lib_kubelet"),
			fmt.Sprintf("%s:/var/lib/rancher/k3s", name+"_var_lib_rancher_k3s"),
			fmt.Sprintf("%s:/var/log", name+"_var_log"),
		},
	}, nil, nil, "kl-cluster-"+name)
	if err != nil {
		return err
	}
	err = k.dClient.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}
