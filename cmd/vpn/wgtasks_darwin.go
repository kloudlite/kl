package vpn

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

func connect(verbose bool, options ...fn.Option) error {
	wc := wg_vpn.NewWgClient()
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
	if err = wc.StartServiceInBg(ifName, configFolder); err != nil {
		fn.Log(text.Yellow(fmt.Sprintf("[#] %s", err)))
	}

	if err := startConfiguration(verbose, options...); err != nil {
		return fn.NewE(err)
	}

	success = true
	return nil
}

func disconnect(verbose bool) error {
	wc := wg_vpn.NewWgClient()

	if err := wc.StopService(verbose); err != nil {
		return err
	}

	return nil
}
