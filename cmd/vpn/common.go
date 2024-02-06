package vpn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"
	wg_svc "github.com/kloudlite/kl/pkg/wg_vpn/wg_service"
)

const (
	ifName string = "utun2464"
)

func startConfiguration(verbose bool, options ...fn.Option) error {
	selectedDevice, err := client.GetDeviceContext()
	if err != nil {
		return err
	}

	switch flags.CliName {
	case constants.CoreCliName:
		envName := fn.GetOption(options, "envName")
		if envName != "" {
			en, err := client.CurrentEnv()
			if err == nil {
				if en.Name != "" && en.Name != envName {
					if err := server.UpdateDeviceEnv(options...); err != nil {
						return err
					}
				}
			}
		}

	case constants.InfraCliName:
		clusterName := fn.GetOption(options, "clusterName")
		if clusterName != "" {

			cn, err := client.CurrentClusterName()
			if err != nil {
				return err
			}
			if cn != "" && cn != clusterName {
				if err := server.UpdateDeviceClusterName(clusterName); err != nil {
					return err
				}
			}

			time.Sleep(2 * time.Second)
		}
	}

	devName := selectedDevice.DeviceName

	device, err := server.GetDevice(fn.MakeOption("deviceName", devName))
	if err != nil {
		switch flags.CliName {
		case constants.CoreCliName:
			return fmt.Errorf("error getting device vpn config, please ensure environment is selected and try again")
		case constants.InfraCliName:
			return fmt.Errorf("error getting device vpn config, please ensure cluster is selected and try again")
		default:
			return err
		}
	}

	switch flags.CliName {
	case constants.CoreCliName:
		envName := fn.GetOption(options, "envName")
		projectName := fn.GetOption(options, "projectName")

		if envName != "" {
			if envName != "" {
				en, err := client.CurrentEnv()
				if err == nil {
					envName = en.Name
				}
			}

			opt := []fn.Option{}
			if projectName != "" {
				opt = append(opt, fn.MakeOption("projectName", projectName))
			}
			opt = append(opt, fn.MakeOption("envName", envName))

			if device.EnvName == "" || device.EnvName != envName {
				if err := server.UpdateDeviceEnv(opt...); err != nil {
					return err
				}
				time.Sleep(2 * time.Second)
			}
		}

	case constants.InfraCliName:
		clusterName := fn.GetOption(options, "clusterName")

		if clusterName == "" {
			if s, err := client.CurrentClusterName(); err != nil {
				return err
			} else {
				clusterName = s
			}
		}

		if device.ClusterName == "" || (device.ClusterName != clusterName) {
			if err := server.UpdateDeviceClusterName(clusterName); err != nil {
				return err
			}

			time.Sleep(2 * time.Second)
		}
	}

	if len(device.Spec.Ports) == 0 {
		fn.Log(text.Yellow(fmt.Sprintf("[#] no ports found for device %s, you can export ports using %s vpn expose\n", devName, flags.CliName)))
	}

	if device.WireguardConfig.Value == "" {
		return errors.New("no wireguard config found, please try again in few seconds")
	}

	configuration, err := base64.StdEncoding.DecodeString(device.WireguardConfig.Value)
	if err != nil {
		return err
	}

	if runtime.GOOS != "linux" {

		if err := wg_svc.StartVpn(configuration); err != nil {
			return err
		}

		return nil
	}

	if err := wg_vpn.Configure(configuration, devName, func() string {
		if runtime.GOOS == "darwin" {
			return ifName
		}
		return devName
	}(), verbose); err != nil {
		return err
	}

	if wg_vpn.IsSystemdReslov() {
		if err := wg_vpn.ExecCmd(fmt.Sprintf("resolvectl domain %s %s", device.Metadata.Name, func() string {
			if device.Spec.ActiveNamespace != "" {
				return fmt.Sprintf("%s.svc.cluster.local", device.Spec.ActiveNamespace)
			}

			return "~."
		}()), false); err != nil {
			return err
		}
	}
	return nil
}
