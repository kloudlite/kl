package kubectl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/k3s"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var K9sCmd = &cobra.Command{
	Use:                "k9s",
	Short:              "k9s is a terminal UI for Kubernetes",
	Hidden:             true,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {

		kconfPath, err := GetPath(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}

		c := exec.Command("k9s", args...)
		c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kconfPath))

		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		if err := c.Run(); err != nil {
			err = nil
			return
		}
	},
}

func GetPath(cmd *cobra.Command) ([]byte, error) {
	kc, err := k3s.NewClient(cmd)
	if err != nil {
		return nil, err
	}

	out, err := kc.Exec("cat /etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return nil, err
	}

	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return nil, err
	}

	kconfPath := path.Join(td, "k3s.yaml")

	k3sPort, err := fileclient.File.GetDataContext().GetK3sPort()
	if err != nil {
		return nil, err
	}

	out = bytes.Replace(out, []byte("127.0.0.1:6443"), []byte(fmt.Sprintf("127.0.0.1:%s", *k3sPort)), 1)

	if err := os.WriteFile(kconfPath, out, 0644); err != nil {
		return nil, err
	}

	return []byte(kconfPath), nil
}

func init() {
}
