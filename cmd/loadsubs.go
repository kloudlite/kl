package cmd

import (
	"github.com/kloudlite/kl/cmd/auth"
	"github.com/kloudlite/kl/cmd/infra"
	"github.com/kloudlite/kl/cmd/vpn"
	"github.com/spf13/cobra"

	"github.com/kloudlite/kl/cmd/list"
	"github.com/kloudlite/kl/cmd/runner"
	"github.com/kloudlite/kl/cmd/runner/add"
)

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	rootCmd.AddCommand(DocsCmd)
	rootCmd.AddCommand(UpdateCmd)

	rootCmd.AddCommand(list.Cmd)

	rootCmd.AddCommand(auth.Cmd)

	rootCmd.AddCommand(infra.Cmd)
	rootCmd.AddCommand(vpn.Cmd)

	rootCmd.AddCommand(runner.InitCommand)

	rootCmd.AddCommand(add.Command)
}
