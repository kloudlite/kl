package vpn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

const (
	ifName string = "utun2464"
)

func connect(verbose bool, options ...fn.Option) error {
	return fmt.Errorf("not supported for windows")

	if err := func() error {

		f, err := fileclient.GetConfigFolder()
		if err != nil {
			return err
		}

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

		pth := path.Join(f, fmt.Sprintf("%s.conf", ifName))

		if err := os.WriteFile(pth, configuration, os.ModePerm); err != nil {
			return err
		}

		if _, err := exec.LookPath("wireguard"); err != nil {
			return fmt.Errorf("can't find wireguard in path, please ensure it's installed. installation link %s", text.Blue("https://www.wireguard.com/install"))
		}

		if _, err := fn.WinSudoExec(fmt.Sprintf("%s /installtunnelservice %s", "wireguard", pth), nil); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func disconnect(verbose bool) error {

	if _, err := fn.WinSudoExec(fmt.Sprintf("%s /uninstalltunnelservice %s", "wireguard", ifName), map[string]string{"PATH": os.Getenv("PATH")}); err != nil {
		return err
	}

	return nil
}
