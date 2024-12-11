package vpn

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

func connect(verbose bool, options ...fn.Option) error {

	success := false
	defer func() {
		if !success {
			_ = wg_vpn.StopService(verbose)
		}
	}()

	configFolder, err := fileclient.GetConfigFolder()
	if err != nil {
		return err
	}

	// TODO: handle this error later
	if err = wg_vpn.StartServiceInBg(ifName, configFolder); err != nil {
		fn.Log(text.Yellow(fmt.Sprintf("[#] %s", err)))
	}

	if err := startConfiguration(verbose, options...); err != nil {
		_ = wg_vpn.ResetDnsServers(ifName, verbose)
		return fn.NewE(err)
	}
	success = true

	return nil
}

func disconnect(verbose bool) error {

	if err := wg_vpn.StopService(verbose); err != nil {
		return err
	}

	if err := wg_vpn.ResetDnsServers(ifName, verbose); err != nil {
		return err
	}

	return nil
}
