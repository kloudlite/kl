package list

import (
	"errors"
	"fmt"
	"os"


	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Get list of secrets in selected environment",
	Run: func(cmd *cobra.Command, _ []string) {

		cwd, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}
		env, err := server.EnvAtPath(cwd)
		if err != nil {
			fn.PrintError(err)
			return
		}

		klFile, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			fn.PrintError(errors.New("please run 'kl init' if you are not initialized the file already"))
			return
		}

		sec, err := server.ListSecrets([]fn.Option{
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printSecrets(cmd, sec); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func printSecrets(_ *cobra.Command, secrets []server.Secret) error {
	if len(secrets) == 0 {
		return fn.NewError("no secrets found")
	}

	header := table.Row{
		table.HeaderText("Display Name"),
		table.HeaderText("Name"),
		table.HeaderText("entries"),
	}

	rows := make([]table.Row, 0)

	for _, a := range secrets {
		rows = append(rows, table.Row{a.DisplayName, a.Metadata.Name, fmt.Sprintf("%d", len(a.StringData))})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(secrets), true)
	return nil
}

func init() {
	secretsCmd.Aliases = append(secretsCmd.Aliases, "sec")
}
