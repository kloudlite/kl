package wg_vpn

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Fa1k3n/resolvconf"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/miekg/dns"
	"github.com/vishvananda/netlink"
)

const (
	bkpPath = "/etc/.resolv.conf.kl-bkp"
	resPath = "/etc/resolv.conf"
)

const (
	ifName = constants.InterfaceName
)

func (wc *wgClientImpl) resetSearchDomain(devName string) error {
	return wc.setSearchDomain("", devName)
}

func (wc *wgClientImpl) setSearchDomain(domain string, deviceName string) error {
	if IsSystemdReslov() {
		if err := ExecCmd(fmt.Sprintf("resolvectl domain %s %s", deviceName, func() string {
			if domain == "" {
				return "~."
			}

			return domain
		}()), false); err != nil {
			return err
		}

		return nil
	}

	fc := fileclient.File
	e, err := fc.GetExtraData()
	if err != nil {
		return err
	}

	hs, err := e.GetDnsHostSuffix()
	if err != nil {
		return err
	}

	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return err
	}

	currSearchDomains := config.Search
	newSearchDomains := make([]string, 0)
	for _, v := range currSearchDomains {
		if strings.HasSuffix(v, hs) {
			continue
		}
		newSearchDomains = append(newSearchDomains, v)
	}

	if _, err := os.Stat(bkpPath); os.IsNotExist(err) {
		if err := copyFile(resPath, bkpPath); err != nil {
			return err
		}
	}

	if domain != "" {
		newSearchDomains = append(newSearchDomains, domain)
	}

	config.Search = newSearchDomains

	s := clientConfigToString(config)

	return os.WriteFile(resPath, []byte(s), 0644)
}

func (wc *wgClientImpl) setDnsServers(dnsServers []net.IP, deviceName string, verbose bool) error {

	if IsSystemdReslov() {
		if len(dnsServers) == 0 {
			fn.Warnf("No DNS server configured for %s", deviceName)
			return nil
		}

		return ExecCmd(fmt.Sprintf("resolvectl dns %s %s", deviceName, dnsServers[0].String()), verbose)
	}

	if _, err := os.Stat(bkpPath); os.IsNotExist(err) {
		if err := copyFile(resPath, bkpPath); err != nil {
			return err
		}
	}

	conf := resolvconf.New()

	for _, v := range dnsServers {
		conf.Add(resolvconf.NewNameserver(v))
	}

	writer, err := os.Create("/etc/resolv.conf")
	if err != nil {
		return err
	}

	if err := conf.Write(writer); err != nil {
		return err
	}

	return nil
}

func (wc *wgClientImpl) resetDnsServers(deviceName string, verbose bool) error {
	if _, err := os.Stat(bkpPath); err == nil {
		if err := copyFile(bkpPath, resPath); err != nil {
			return err
		}
		return os.Remove(bkpPath)
	}

	return nil
}

func getCurrentDns(_ bool) ([]string, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return nil, err
	}

	return config.Servers, nil
}

func (wc *wgClientImpl) startService(devName string, _ bool) error {

	// Add Wireguard device
	wgLink := &netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{Name: devName},
		LinkType:  "wireguard",
	}
	if err := netlink.LinkAdd(wgLink); err != nil {
		return fmt.Errorf("failed to add WireGuard interface: %v", err)
	}

	// Set MTU and bring up the device
	link, err := netlink.LinkByName(devName)
	if err != nil {
		return fmt.Errorf("failed to find the interface %s: %v", devName, err)
	}
	if err := netlink.LinkSetMTU(link, 1420); err != nil {
		return fmt.Errorf("failed to set MTU for %s: %v", devName, err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring up the interface %s: %v", devName, err)
	}

	return nil
}

func (wc *wgClientImpl) ipRouteAdd(ip string, _ string, devName string, _ bool) error {
	_, dst, err := net.ParseCIDR(ip)
	if err != nil {
		return fmt.Errorf("failed to parse CIDR: %v", err)
	}

	link, err := netlink.LinkByName(devName)
	if err != nil {
		return fmt.Errorf("failed to find the interface %s: %v", devName, err)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Scope:     netlink.SCOPE_UNIVERSE,
	}
	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("failed to add the route: %v", err)
	}

	return nil
}

func (wc *wgClientImpl) stopService(devName string, verbose bool) error {
	wgInterface, err := wgc.Show(&wgc.WgShowOptions{
		Interface: "interfaces",
	})
	if err != nil {
		return err
	}

	if len(wgInterface) == 0 {
		return nil
	}
	for _, v := range wgInterface {

		if strings.TrimSpace(v) == "" || v != devName {
			continue
		}

		if verbose {
			fn.Log("[#] disconnecting from ", v)
		}

		link, err := netlink.LinkByName(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("failed to find the interface %s: %v", wgInterface, err)
		}

		if err := netlink.LinkDel(link); err != nil {
			return fmt.Errorf("failed to delete the interface %s: %v", wgInterface, err)
		}
	}

	return nil
}

func (wc *wgClientImpl) setDeviceIp(ip net.IPNet, deviceName string, _ bool) error {
	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		return fmt.Errorf("failed to find the interface %s: %v", deviceName, err)
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip.IP,
			Mask: ip.Mask,
		},
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("failed to set IP address for %s: %v", deviceName, err)
	}

	return nil
}
