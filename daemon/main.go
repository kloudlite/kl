package daemon

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/daemon/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

func RunApp(binName string) error {
	wc := wg_vpn.NewWgClient()

	signalChannel := make(chan os.Signal, 1)

	// Notify the channel for SIGINT (Ctrl+C), SIGTERM, etc.
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	// done := make(chan bool)

	ctx, cf := context.WithCancel(context.Background())

	ch := make(chan error, 0)

	go func() {
		<-signalChannel
		if err := wc.StopService(constants.InterfaceName, true); err != nil {
			fn.Debug(err)
			ch <- err
		}

		ch <- nil
	}()

	go func() {
		s := server.New(binName)
		ch <- s.Start(ctx)
	}()

	select {
	case i := <-ch:
		cf()
		return i
	}
}
