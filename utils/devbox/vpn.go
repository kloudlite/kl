package devbox

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/kloudlite/kl2/server"
	"os"
)

type AccountVpnConfig struct {
	WGconf string `json:"wg"`
	DeviceName string `json:"device"`
}

func createVpnForAccount(account string) (*server.Device, error) {
	devName, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	checkNames, err := server.GetDeviceName(devName, account)
	if err != nil {
		return nil, err
	}
	if !checkNames.Result {
		if len(checkNames.SuggestedNames) == 0 {
			return nil, fmt.Errorf("no suggested names for device %s", devName)
		}
		devName = checkNames.SuggestedNames[0]
	}
	device, err := server.CreateDevice(devName, account)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func vpnConfigForAccount(account string) (string, error) {
	cfgFolder, err := getConfigFolder()
	if err != nil {
		return "", err
	}
	cfgPath := cfgFolder + "/vpn/" + account + ".json"
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		dev, err := createVpnForAccount(account)
		if err != nil {
			return "", err
		}
		accountVpnConfig := AccountVpnConfig{
			WGconf: dev.WireguardConfig.Value,
			DeviceName: dev.Metadata.Name,
		}

	}
	c, err := os.ReadFile(cfgPath)
	if err != nil {
		return "", errors.New("failed to read vpn config")
	}
	return string(c), nil
}

func SyncVpn(wg string) error {
	err := ensureImage("ghcr.io/kloudlite/hub/wireguard:latest")
	if err != nil {
		return errors.New("failed to pull image")
	}
	cli, err := dockerClient()
	if err != nil {
		return errors.New("failed to create docker client")
	}
	existingVPN, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "wg=true"),
		),
	})
	if err != nil {
		return errors.New("failed to list containers")
	}
	md5sum := md5.Sum([]byte(wg))
	if len(existingVPN) > 0 {
		if existingVPN[0].Labels["wgsum"] == string(md5sum[:]) {
			if existingVPN[0].State != "running" {
				err := cli.ContainerStart(context.Background(), existingVPN[0].ID, container.StartOptions{})
				if err != nil {
					return errors.New("failed to start container")
				}
			}
			return nil
		}
		err := cli.ContainerStop(context.Background(), existingVPN[0].ID, container.StopOptions{
			Signal: "SIGKILL",
		})
		if err != nil {
			return errors.New("failed to stop container")
		}
		err = cli.ContainerRemove(context.Background(), existingVPN[0].ID, container.RemoveOptions{
			Force: true,
		})
		if err != nil {
			return errors.New("failed to remove container")
		}
	}

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			"kloudlite": "true",
			"wg":        "true",
			"wgsum":     string(md5sum[:]),
		},
		Image: "ghcr.io/kloudlite/hub/wireguard:latest",
		Cmd:   []string{"sh", "-c", "echo " + wg + " > /wireguard/wg0.conf && wg-quick up wg0 && tail -f /dev/null"},
	}, &container.HostConfig{}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return errors.New("failed to create container")
	}
	err = cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return errors.New("failed to start container")
	}
	return nil
}
