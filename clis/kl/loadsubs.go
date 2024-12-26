package kl

import (

	// "github.com/kloudlite/kl/cmd/box"

	"os/exec"
	"runtime"

	"github.com/kloudlite/kl/cmd/auth"
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
	"github.com/kloudlite/kl/flags"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	if flags.IsDev() {
		rootCmd.AddCommand(DocsCmd)
	}

	rootCmd.AddCommand(shell.CheckCmd)
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(UpdateCmd)
	rootCmd.AddCommand(set_base_url.Cmd)
	rootCmd.AddCommand(initp.InitCommand)

	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(get.Cmd)
	rootCmd.AddCommand(vpn.Cmd)
	rootCmd.AddCommand(daemon_cmd.Cmd)

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

	// Not Required for now
	// rootCmd.AddCommand(env.Cmd)
	// rootCmd.AddCommand(clone.Cmd)

	if runtime.GOOS == constants.RuntimeDarwin {
		return
	}

	// rootCmd.AddCommand(box.BoxCmd)

	// rootCmd.AddCommand(runner.AttachCommand)
	//
	//
	// rootCmd.AddCommand(cluster.Cmd)
	//
	// rootCmd.AddCommand(connect.Command)
	// rootCmd.AddCommand(v2Shell.Command)
	// rootCmd.AddCommand(v2Add.Command)
	// rootCmd.AddCommand(v2Pkg.Command)
	// rootCmd.AddCommand(v2Lib.Command)

	if _, err := exec.LookPath("k9s"); err == nil {
		rootCmd.AddCommand(kubectl.K9sCmd)
	}

	if _, err := exec.LookPath("kubectl"); err == nil {
		rootCmd.AddCommand(kubectl.KubectlCmd)
	}
}
