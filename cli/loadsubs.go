package cli

import (
	"github.com/kloudlite/kl2/cmds/add"
	"github.com/kloudlite/kl2/cmds/auth"
	"github.com/kloudlite/kl2/cmds/box"
	"github.com/kloudlite/kl2/cmds/expose"
	"github.com/kloudlite/kl2/cmds/get"
	"github.com/kloudlite/kl2/cmds/list"
	"github.com/kloudlite/kl2/cmds/pkgs"
	"github.com/kloudlite/kl2/cmds/runner"
	set_base_url "github.com/kloudlite/kl2/cmds/set-base-url"
	"github.com/kloudlite/kl2/cmds/use"
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
}
