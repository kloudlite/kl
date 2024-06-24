package cli

import (
	"github.com/kloudlite/kl/cmds"
	"github.com/kloudlite/kl/cmds/add"
	"github.com/kloudlite/kl/cmds/auth"
	"github.com/kloudlite/kl/cmds/box"
	"github.com/kloudlite/kl/cmds/expose"
	"github.com/kloudlite/kl/cmds/get"
	"github.com/kloudlite/kl/cmds/intercept"
	"github.com/kloudlite/kl/cmds/list"
	"github.com/kloudlite/kl/cmds/pkgs"
	"github.com/kloudlite/kl/cmds/runner"
	set_base_url "github.com/kloudlite/kl/cmds/set-base-url"
	"github.com/kloudlite/kl/cmds/use"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	// if flags.IsDev() {
	// 	rootCmd.AddCommand(DocsCmd)
	// }
	rootCmd.AddCommand(add.Command)
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(runner.InitCommand)
	rootCmd.AddCommand(set_base_url.Cmd)
	rootCmd.AddCommand(expose.Command)
	rootCmd.AddCommand(get.Command)
	rootCmd.AddCommand(list.Cmd)
	rootCmd.AddCommand(box.Cmd)
	rootCmd.AddCommand(pkgs.Cmd)
	rootCmd.AddCommand(use.Cmd)
	rootCmd.AddCommand(cmds.StatusCmd)
	rootCmd.AddCommand(intercept.Cmd)
}
