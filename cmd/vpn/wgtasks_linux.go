package vpn

import (
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

const (
	ifName string = "kl"
)

func connect(verbose bool, options ...fn.Option) error {
	wc := wg_vpn.NewWgClient()
	success := false

	defer func() {
		if !success {
			_ = wc.StopService(verbose)
		}
	}()

	// TODO: handle this error later
	if err := wc.StartService(ifName, verbose); err != nil {
		fn.Log(text.Yellow(fmt.Sprintf("[#] %s", err)))
	}

	if err := startConfiguration(verbose, options...); err != nil {
		return err
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
