package vpn

import (
	"encoding/base64"
	"errors"

	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

const (
	ifName string = "utun2464"
)

func startConfiguration(verbose bool, _ ...fn.Option) error {
	apic, err := apiclient.New()
	if err != nil {
		return err
	}

	device, err := apic.EnsureDevice()
	if err != nil {
		return err
	}

	if device.WGconf == "" {
		return errors.New("no wireguard config found, please try again in few seconds")
	}

	configuration, err := base64.StdEncoding.DecodeString(device.WGconf)
	if err != nil {
		return err
	}

	if err := wg_vpn.NewWgClient().Configure(configuration, ifName, verbose); err != nil {
		return err
	}

	// TODO: update search domain at the time of env switch and use
	// if wg_vpn.IsSystemdReslov() {
	// 	if err := wg_vpn.ExecCmd(fmt.Sprintf("resolvectl domain %s %s", device.DeviceName, func() string {
	// 		s, err := apic.GetFClient().GetDataContext().GetSearchDomain()
	// 		if err != nil {
	// 			return "~."
	// 		}
	//
	// 		return s
	// 	}()), false); err != nil {
	// 		return err
	// 	}
	// } else {
	// 	s, err := apic.GetFClient().GetDataContext().GetSearchDomain()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	return wg_vpn.SetSearchDomain(s)
	// }

	return nil
}
