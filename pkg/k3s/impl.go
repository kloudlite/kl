package k3s

import (
	"bytes"
	"context"
	_ "embed"
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

//go:embed scripts/startup-script.sh.tmpl
var startupScript string

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
		Privileged:  true,
		NetworkMode: "kloudlite",
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

	p, err := t.Parse(startupScript)
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
