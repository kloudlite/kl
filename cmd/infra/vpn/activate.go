package vpn

import (
	"errors"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "activate vpn in any environment",
	Long: `This command let you activate vpn in any environment.
Example:
  # activate vpn in any environment
  kl infra vpn activate -n <namespace>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		err := activateInfraVPN(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func activateInfraVPN(cmd *cobra.Command) error {

	ns := ""

	if cmd.Flags().Changed("name") {
		ns, _ = cmd.Flags().GetString("name")
	}
	if ns == "" {
		return errors.New("namespace is missing, please provide using kl infra vpn activate -n <namespace>")
	}
	if err := server.UpdateInfraDeviceNS(ns); err != nil {
		fn.PrintError(err)
		return err
	}

	fn.Log("namespace updated successfully")
	return nil
}

func init() {
	activateCmd.Aliases = append(listCmd.Aliases, "active", "act", "a")
	activateCmd.Flags().StringP("name", "n", "", "namespace")
}
