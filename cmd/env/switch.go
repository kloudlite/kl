package env

import (
	"fmt"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch to a different environment",
	Long: `Switch Environment
Examples:
  # switch to a different environment
  kl env switch

  # switch to a different environment with environment name
  kl env switch 
	`,

	Run: func(cmd *cobra.Command, _ []string) {
		err := switchAccounts(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func switchAccounts(cmd *cobra.Command) error {

	envName := fn.ParseStringFlag(cmd, "envname")
	projectName := fn.ParseStringFlag(cmd, "projectname")

	env, err := server.SelectEnv(envName)
	if err != nil {
		fn.PrintError(err)
		return err
	}

	proj, err := server.SelectProject(projectName)
	if err != nil {
		fn.PrintError(err)
		return err
	}

	if err := server.UpdateDeviceEnv([]fn.Option{
		fn.MakeOption("envName", env.Metadata.Name),
		fn.MakeOption("projectName", proj.Metadata.Name),
	}...); err != nil {
		return err
	}

	fn.Log(text.Bold(text.Green("\nSelected Environment and Project:")),
		text.Blue(fmt.Sprintf("\n%s (%s) and %s (%s)", env.DisplayName, env.Metadata.Name, proj.DisplayName, proj.Metadata.Name)),
	)
	return nil
}

func init() {
	switchCmd.Aliases = append(switchCmd.Aliases, "sw")

	switchCmd.Flags().StringP("envname", "e", "", "environment name")
	switchCmd.Flags().StringP("projectname", "p", "", "project name")
}
