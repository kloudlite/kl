package list

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List [accounts | envs | configs | secrets | apps | mres]",
	Long: `Use this command to list resources like,
  account, environments, configs, secrets and apps`,
}

func init() {
	Cmd.AddCommand(accCmd)
	Cmd.AddCommand(envCmd)
	Cmd.AddCommand(appsCmd)
	Cmd.AddCommand(configsCmd)
	Cmd.AddCommand(secretsCmd)
	Cmd.AddCommand(mresCmd)
	Cmd.Aliases = append(Cmd.Aliases, "ls")
}
