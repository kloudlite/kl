package kl

import (

	// "github.com/kloudlite/kl/cmd/box"

	"os/exec"
	"runtime"

	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/cluster"
	daemon_cmd "github.com/kloudlite/kl/cmd/daemon"
	"github.com/kloudlite/kl/cmd/get"
	"github.com/kloudlite/kl/cmd/initp"
	"github.com/kloudlite/kl/cmd/intercept"
	"github.com/kloudlite/kl/cmd/kubectl"
	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/cmd/packages/lib"
	"github.com/kloudlite/kl/cmd/packages/pkg"
	"github.com/kloudlite/kl/cmd/packages/shell"
	"github.com/kloudlite/kl/cmd/runner/add"
	set_base_url "github.com/kloudlite/kl/cmd/set-base-url"
	"github.com/kloudlite/kl/cmd/status"
	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/cmd/vpn"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/flags"
	"github.com/spf13/cobra"
)

func init() {
	isBoxMode := fileclient.IsBoxMode()

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	if flags.IsDev() {
		rootCmd.AddCommand(DocsCmd)
	}

	rootCmd.AddCommand(shell.CheckCmd)

	if !isBoxMode {
		rootCmd.AddCommand(auth.Cmd)
		rootCmd.AddCommand(UpdateCmd)

		rootCmd.AddCommand(set_base_url.Cmd)
	}

	rootCmd.AddCommand(initp.InitCommand)
	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(get.Cmd)
	rootCmd.AddCommand(vpn.Cmd)

	if !isBoxMode {
		rootCmd.AddCommand(daemon_cmd.Cmd)
	}

	if runtime.GOOS == constants.RuntimeWindows {
		return
	}

	rootCmd.AddCommand(pkg.Cmd)
	rootCmd.AddCommand(lib.Cmd)
	rootCmd.AddCommand(shell.Cmd)

	rootCmd.AddCommand(add.Command)

	rootCmd.AddCommand(intercept.Cmd)
	// rootCmd.AddCommand(expose.Cmd)

	rootCmd.AddCommand(status.Cmd)

	if runtime.GOOS == constants.RuntimeDarwin {
		return
	}

	if !fileclient.IsBoxMode() {
		rootCmd.AddCommand(cluster.Cmd)

		if _, err := exec.LookPath("helm"); err == nil {
			rootCmd.AddCommand(kubectl.HelmCmd)
		}

		if _, err := exec.LookPath("k9s"); err == nil {
			rootCmd.AddCommand(kubectl.K9sCmd)
		}

		if _, err := exec.LookPath("kubectl"); err == nil {
			rootCmd.AddCommand(kubectl.KubectlCmd)
		}
	}
}
