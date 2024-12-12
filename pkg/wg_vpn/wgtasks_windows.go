package wg_vpn

import (
	"net"

	fn "github.com/kloudlite/kl/pkg/functions"
)

var notSupportedInWindows = fn.Errorf("not supported on windows")

func (wc *wgClientImpl) resetSearchDomain() error {
	return notSupportedInWindows
}

func (wc *wgClientImpl) setSearchDomain(domain string) error {
	return notSupportedInWindows
}

func (wc *wgClientImpl) setDnsServers(dnsServers []net.IP, deviceName string, verbose bool) error {
	return notSupportedInWindows
}

func (wc *wgClientImpl) resetDnsServers(deviceName string, verbose bool) error {
	return notSupportedInWindows
}

func (wc *wgClientImpl) setDeviceIp(ip net.IPNet, deviceName string, verbose bool) error {
	return notSupportedInWindows
}
func (wc *wgClientImpl) ipRouteAdd(ip string, _ string, devName string, _ bool) error {
	return notSupportedInWindows
}

func (wc *wgClientImpl) startService(devName string, _ bool) error {
	return notSupportedInWindows
}

func (wc *wgClientImpl) stopService(verbose bool) error {
	return notSupportedInWindows
}
