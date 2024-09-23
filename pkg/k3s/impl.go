package k3s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"os"
	"text/template"
)

const (
	CONT_MARK_KEY = "kl.container"
)

func (c *client) CreateClustersAccounts(accountName string) error {
	if err := c.ensureImage(constants.GetK3SImageName()); err != nil {
		return fn.NewE(err)
	}
	existingContainers, err := c.c.ContainerList(context.Background(), container.ListOptions{
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
		for _, ec := range existingContainers {
			if ec.Labels["kl-account"] != accountName {
				fn.Log(text.Yellow(fmt.Sprintf("[#] another cluster is running for another account. do you want to stop it and start cluster for account %s? [Y/n] ", accountName)))
				if !fn.Confirm("Y", "Y") {
					return nil
				}
				if err := c.c.ContainerStop(context.Background(), ec.ID, container.StopOptions{}); err != nil {
					return fn.Error("failed to stop container")
				}
				if err := c.c.ContainerRemove(context.Background(), ec.ID, container.RemoveOptions{}); err != nil {
					return fn.Error("failed to remove container")
				}
			}
		}
	}

	existingContainers, err = c.c.ContainerList(context.Background(), container.ListOptions{
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
			if err := c.c.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{}); err != nil {
				return fn.Error("failed to start container")
			}
		}
		return nil
	}

	clusterConfig, err := c.apic.GetClusterConfig(accountName)
	if err != nil {
		return fn.NewE(err)
	}
	// Expose UDP 31820 port
	createdConatiner, err := c.c.ContainerCreate(context.Background(), &container.Config{
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
		ExposedPorts: nat.PortSet{
			"51820/udp": struct{}{},
			"6443/tcp":  struct{}{},
		},
	}, &container.HostConfig{
		Privileged: true,
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds: []string{
			fmt.Sprintf("kl-k3s-%s-cache:/var/lib/rancher/k3s", clusterConfig.ClusterName),
		},
		PortBindings: map[nat.Port][]nat.PortBinding{
			"6443/tcp": {
				{
					HostPort: "6443",
				},
			},
			"51820/udp": {
				{
					HostPort: "51820",
				},
			},
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return fn.Error("failed to create container")
	}

	if err := c.c.ContainerStart(context.Background(), createdConatiner.ID, container.StartOptions{}); err != nil {
		return fn.Error("failed to start container")
	}

	script, err := generateConnectionScript(clusterConfig)
	if err != nil {
		return fn.Error("failed to generate connection script")
	}
	execConfig := container.ExecOptions{
		Cmd: []string{"sh", "-c", script},
	}

	resp, err := c.c.ContainerExecCreate(context.Background(), createdConatiner.ID, execConfig)
	if err != nil {
		return fn.Error("failed to create exec")
	}

	err = c.c.ContainerExecStart(context.Background(), resp.ID, container.ExecStartOptions{})
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

kubectl create ns kl-gateway
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wg-proxy
  namespace: kl-gateway
spec:
  selector:
    matchLabels:
      app: wg-proxy
  replicas: 1
  template:
    metadata:
      labels:
        app: wg-proxy
    spec:
      containers:
        - name: wg-proxy
          image: ghcr.io/kloudlite/kl/box/wireguard:v1.0.0-nightly
          securityContext:
            capabilities:
              add: ["NET_ADMIN"]
          env:
            - name: GATEWAY_ENDPOINT
              value: default-wg:31820
            - name: PRIVATE_KEY
              value: {{.WGConfig.Proxy.PrivateKey}}
            - name: GATEWAY_PUBLIC_KEY
              valueFrom:
                secretKeyRef:
                  name: kl-gateway
                  key: public_key 
            - name: WORKSPACE_PUBLIC_KEY
              value: {{.WGConfig.Workspace.PublicKey}}
            - name: HOST_PUBLIC_KEY
              value: {{.WGConfig.Host.PublicKey}}
---
apiVersion: v1
kind: Service
metadata:
  name: wg-proxy
  namespace: kl-gateway
spec:
  type: LoadBalancer
  selector:
    app: wg-proxy
  ports:
    - protocol: UDP
      port: 51820
      targetPort: 31820 
EOF

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-local-overrides
  namespace: kloudlite
data:
  peers: |+
    - allowedIPs:
      - 192.18.0.1/32
      publicKey: {{.WGConfig.Proxy.PublicKey}}
EOF

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
    agentOperator:
      image:
        repository: ghcr.io/kloudlite/operator/agent
        tag: v1.0.8-alpha
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

func (c *client) imageExists(imageName string) (bool, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", imageName)
	images, err := c.c.ImageList(context.Background(), image.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *client) ensureImage(i string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("checking image %s", i))()

	if imageExists, err := c.imageExists(i); err == nil && imageExists {
		return nil
	}

	out, err := c.c.ImagePull(context.Background(), i, image.PullOptions{})
	if err != nil {
		return fn.NewE(err, fmt.Sprintf("failed to pull image %s", i))
	}
	defer out.Close()

	jsonmessage.DisplayJSONMessagesStream(out, os.Stdout, os.Stdout.Fd(), true, nil)
	return nil
}
