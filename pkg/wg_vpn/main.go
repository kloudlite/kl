package wg_vpn

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"golang.zx2c4.com/wireguard/wgctrl"
)

type WgClient interface {
	StartServiceInBg(devName string, configFolder string) error

	StartService(devName string, verbose bool) error
	startService(devName string, verbose bool) error

	StopService(verbose bool) error
	stopService(verbose bool) error

	Configure(configuration []byte, interfaceName string, verbose bool) error

	SetDnsServers(dnsServers []net.IP, devName string, verbose bool) error
	ResetDnsServers(devName string, verbose bool) error

	SetSearchDomain(domain string) error
	setSearchDomain(domain string, devName string) error

	ResetSearchDomain() error
	resetSearchDomain(devName string) error

	setDnsServers(dnsServers []net.IP, devName string, verbose bool) error
	resetDnsServers(devName string, verbose bool) error

	setDeviceIp(ip net.IPNet, deviceName string, verbose bool) error
	ipRouteAdd(ip string, _ string, devName string, _ bool) error
}

type wgClientImpl struct{}

func NewWgClient() WgClient {
	return &wgClientImpl{}
}

// Deprecated: start service from daemon istead
func (c *wgClientImpl) StartServiceInBg(devName string, configFolder string) error {
	s, err := exec.LookPath(flags.CliName)
	if err != nil {
		fn.Warn(err)
		return err
	}

	command := exec.Command(s, "vpn", "start-fg", "-d", devName)
	if err := command.Start(); err != nil {
		return err
	}

	err = os.WriteFile(configFolder+"/wgpid", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0644)
	if err != nil {
		fn.PrintError(err)
		return err
	}

	if _, ok := os.LookupEnv("SUDO_USER"); ok {
		uid, gid, err := fn.GetUidNGid()
		if err != nil {
			return err
		}

		if err = os.Chown(configFolder+"/wgpid", uid, gid); err != nil {
			fn.PrintError(err)
			return err
		}
	}

	return nil
}

func (c *wgClientImpl) Configure(
	configuration []byte,
	interfaceName string,
	verbose bool,
) error {

	stopSpinner := spinner.Client.UpdateMessage("validating configuration")
	cfg := wgc.Config{}

	if verbose {
		fn.Log("[#] validating configuration")
	}

	if e := cfg.UnmarshalText(configuration); e != nil {
		return e
	}

	stopSpinner()

	if len(cfg.Address) == 0 {
		return errors.New("device ip not found")
	} else if e := c.setDeviceIp(cfg.Address[0], interfaceName, verbose); e != nil {
		return e
	}

	wg, err := wgctrl.New()
	if err != nil {
		return err
	}

	if verbose {
		fn.Log("[#] setting up connection")
	}

	// TODO: needs to managed separately
	// if err := c.setDnsServers(cfg.DNS, interfaceName, verbose); err != nil {
	// 	return err
	// }

	err = wg.ConfigureDevice(interfaceName, cfg.Config)
	if err != nil {
		return err
	}

	for _, pc := range cfg.Peers {
		for _, i2 := range pc.AllowedIPs {
			err = c.ipRouteAdd(i2.String(), cfg.Address[0].IP.String(), interfaceName, verbose)
			if err != nil {
				return err
			}
		}
	}

	if err != nil {
		return err
	}

	// TODO: needs to managed separately
	// if len(cfg.DNS) > 0 {
	// 	fc:= clients.File
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	ed, err := fc.GetExtraData()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	ed.SetBackupDns(func() []string {
	// 		resp := make([]string, len(cfg.DNS))
	// 		for i, v := range cfg.DNS {
	// 			resp[i] = v.To4().String()
	// 		}
	// 		return resp
	// 	}())
	// }

	return nil
}

func (c *wgClientImpl) SetDnsServers(dnsServers []net.IP, devName string, verbose bool) error {
	return c.setDnsServers(dnsServers, devName, verbose)
}

func (c *wgClientImpl) ResetDnsServers(devName string, verbose bool) error {
	return c.resetDnsServers(devName, verbose)
}

func (c *wgClientImpl) SetSearchDomain(domain string) error {
	return c.setSearchDomain(domain, ifName)
}

func (c *wgClientImpl) ResetSearchDomain() error {
	return c.resetSearchDomain(ifName)
}

func (c *wgClientImpl) StartService(devName string, verbose bool) error {
	return c.startService(devName, verbose)
}

func (c *wgClientImpl) StopService(verbose bool) error {
	return c.stopService(verbose)
}
