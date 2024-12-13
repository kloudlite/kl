package add

import (
	"fmt"

	"github.com/kloudlite/kl/cmd/v2/internal/lib"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "v2:add",
	Short: "",
}

func init() {
	Command.AddCommand(
		&cobra.Command{
			Use: "config [key] [config-name/config-key]",
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) != 0 && len(args) != 2 {
					return fmt.Errorf("must be either no arguments or 2 arguments")
				}
				return nil
			},
			Run: func(cmd *cobra.Command, args []string) {
				if len(args) == 0 {
					// TODO: use fzf to ask user
					panic("not implemented")
				}

				klCfg, _, err := lib.PreCommand()
				if err != nil {
					panic(err)
				}

				klCfg.AddEnvVar(lib.EnvType{
					Key:       args[0],
					ConfigRef: &args[1],
				})
			},
		},
		&cobra.Command{
			Use: "secret",
			Run: func(cmd *cobra.Command, args []string) {
				if len(args) == 0 {
					// TODO: use fzf to ask user
					panic("not implemented")
				}

				klCfg, _, err := lib.PreCommand()
				if err != nil {
					panic(err)
				}

				klCfg.AddEnvVar(lib.EnvType{
					Key:       args[0],
					SecretRef: &args[1],
				})
			},
		},
		&cobra.Command{
			Use: "managed-resource",
			Run: func(cmd *cobra.Command, args []string) {
				if len(args) == 0 {
					// TODO: use fzf to ask user
					panic("not implemented")
				}

				klCfg, _, err := lib.PreCommand()
				if err != nil {
					panic(err)
				}

				klCfg.AddEnvVar(lib.EnvType{
					Key:     args[0],
					MresRef: &args[1],
				})
			},
		},
	)
}
