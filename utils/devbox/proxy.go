package devbox

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/pkg/functions"
)

type ProxyConfig struct {
	TargetContainerId   string
	TargetContainerPath string
	ExposedPorts        []int
}

func SyncProxy(config ProxyConfig) error {
	if err := ensureImage(constants.SocatImage); err != nil {
		return functions.Error(err, "failed to pull image")
	}

	cli, err := dockerClient()
	if err != nil {
		return functions.Error(err, "failed to create docker client")
	}

	targetContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", fmt.Sprintf("working_dir=%s", config.TargetContainerPath)),
		),
	})
	if err != nil {
		return functions.Error(err, "failed to list containers")
	}

	existingProxies, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "proxy=true"),
		),
	})
	if err != nil {
		return functions.Error(err, "failed to list containers")
	}

	if len(existingProxies) > 0 {
		err := cli.ContainerStop(context.Background(), existingProxies[0].ID, container.StopOptions{
			Signal: "SIGKILL",
		})
		if err != nil {
			return functions.Error(err, "failed to stop container")
		}
		err = cli.ContainerRemove(context.Background(), existingProxies[0].ID, container.RemoveOptions{
			Force: true,
		})
		if err != nil {
			return functions.Error(err, "failed to remove container")
		}
	}
	if len(config.ExposedPorts) == 0 {
		return nil
	}

	targetContainer := targetContainers[0]
	targetIpAddress := targetContainer.NetworkSettings.Networks["kloudlite"].IPAddress
	socatCommand := ""
	for _, port := range config.ExposedPorts {
		socatCommand += fmt.Sprintf(`socat TCP-LISTEN:%d,fork TCP:%s:%d & `, port, targetIpAddress, port)
		socatCommand += fmt.Sprintf(`socat UDP-RECVFROM:%d,fork UDP-SENDTO:%s:%d & `, port, targetIpAddress, port)
	}
	socatCommand += "tail -f /dev/null"

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: constants.SocatImage,
		Labels: map[string]string{
			"kloudlite": "true",
			"proxy":     "true",
		},
		ExposedPorts: func() nat.PortSet {
			ports := nat.PortSet{}
			for _, port := range config.ExposedPorts {
				ports[nat.Port(fmt.Sprintf("%d/tcp", port))] = struct{}{}
				ports[nat.Port(fmt.Sprintf("%d/udp", port))] = struct{}{}
			}
			return ports
		}(),
		Entrypoint: []string{"sh", "-c", socatCommand},
	}, &container.HostConfig{
		PortBindings: func() nat.PortMap {
			portBindings := nat.PortMap{}
			for _, port := range config.ExposedPorts {
				portBindings[nat.Port(fmt.Sprintf("%d/tcp", port))] = []nat.PortBinding{
					{
						HostPort: fmt.Sprintf("%d", port),
					},
				}
				portBindings[nat.Port(fmt.Sprintf("%d/udp", port))] = []nat.PortBinding{
					{
						HostPort: fmt.Sprintf("%d", port),
					},
				}
			}
			return portBindings
		}(),
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"kloudlite": {},
		},
	}, nil, "")

	if err != nil {
		return functions.Error(err, "failed to create container")
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return functions.Error(err, "failed to start container")
	}
	return nil
}
