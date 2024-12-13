package shell

import (
	"os"

	"github.com/kloudlite/kl/cmd/v2/internal/lib"
	"github.com/spf13/cobra"
)

// INFO: read more at cmd/box/boxpkg/hashctrl/main.go:231 (generatePersistedEnv)

var Command = &cobra.Command{
	Use:   "shell",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		k, pk, err := lib.PreCommand()
		if err != nil {
			panic(err)
		}

		pkgs := make([]string, 0, len(k.Packages))
		for i := range k.Packages {
			pp, err := lib.ParsePackage(k.Packages[i])
			if err != nil {
				panic(err)
			}
			pkgs = append(pkgs, pp.NixpkgsHash)
		}

		libs := make([]string, 0, len(k.Libraries))
		for i := range k.Libraries {
			pp, err := lib.ParsePackage(k.Libraries[i])
			if err != nil {
				panic(err)
			}
			libs = append(libs, pp.NixpkgsHash)
		}

		if err := lib.NixShell(cmd.Context(), lib.ShellArgs{
			Shell:     os.Getenv("SHELL"),
			EnvVars:   append(pk.EnvVars, "KL_SHELL=true", "KL_HASH="+pk.Hash),
			Packages:  pkgs,
			Libraries: libs,
		}); err != nil {
			panic(err)
		}
	},
}
