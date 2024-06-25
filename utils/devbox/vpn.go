package devbox

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/server"
)

type AccountVpnConfig struct {
	WGconf     string `json:"wg"`
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

func GetAccVPNConfig(account string) (*AccountVpnConfig, error) {
	cfgFolder, err := getConfigFolder()
	if err != nil {
		return nil, fn.Error(err)
	}
	err = os.MkdirAll(path.Join(cfgFolder, "vpn"), 0755)
	if err != nil {
		return nil, fn.Error(err)
	}
	cfgPath := path.Join(cfgFolder, "vpn", fmt.Sprintf("%s.json", account))
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		dev, err := createVpnForAccount(account)
		if err != nil {
			return nil, fn.Error(err)
		}
		accountVpnConfig := AccountVpnConfig{
			WGconf:     dev.WireguardConfig.Value,
			DeviceName: dev.Metadata.Name,
		}
		marshal, err := json.Marshal(accountVpnConfig)
		if err != nil {
			return nil, fn.Error(err)
		}
		err = os.WriteFile(cfgPath, marshal, 0644)
		if err != nil {
			return nil, fn.Error(err)
		}
	}

	var accVPNConfig AccountVpnConfig
	c, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fn.NewError("failed to read vpn config")
	}
	err = json.Unmarshal(c, &accVPNConfig)
	if err != nil {
		return nil, fn.NewError("failed to parse vpn config")
	}

	if accVPNConfig.WGconf == "" {
		d, err := server.GetVPNDevice(accVPNConfig.DeviceName, fn.MakeOption("accountName", account))
		if err != nil {
			return nil, fn.Error(err)
		}

		accVPNConfig.WGconf = d.WireguardConfig.Value

		marshal, err := json.Marshal(accVPNConfig)
		if err != nil {
			return nil, fn.Error(err)
		}
		err = os.WriteFile(cfgPath, marshal, 0644)
		if err != nil {
			return nil, fn.Error(err)
		}
	}

	return &accVPNConfig, nil
}

func SyncVpn(wg string) error {
	err := ensureImage(constants.WireguardImage)
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
		if existingVPN[0].Labels["wgsum"] == fmt.Sprintf("%x", md5sum[:]) {
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
	script := fmt.Sprintf("echo %s | base64 -d > /etc/wireguard/wg0.conf && (wg-quick down wg0 || echo done) && wg-quick up wg0 && tail -f /dev/null", wg)

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Labels: map[string]string{
			"kloudlite": "true",
			"wg":        "true",
			"wgsum":     fmt.Sprintf("%x", md5sum[:]),
		},
		Image: constants.WireguardImage,
		Cmd:   []string{"sh", "-c", script},
	}, &container.HostConfig{
		CapAdd:      []string{"NET_ADMIN"},
		NetworkMode: "host",
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return errors.New("failed to create container")
	}
	err = cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return errors.New("failed to start container")
	}
	return nil
}
