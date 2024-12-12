package daemon

import (
	"context"

	"github.com/kloudlite/kl/daemon/server"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func RunApp(binName string) error {
	fn.Log("kl vpn and proxy controller")

	ctx, cf := context.WithCancel(context.Background())

	ch := make(chan error, 0)

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
