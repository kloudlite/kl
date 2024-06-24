package get

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/fzf"
	"github.com/kloudlite/kl2/pkg/ui/table"
	"github.com/kloudlite/kl2/server"
	"github.com/kloudlite/kl2/utils/klfile"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var secretCmd = &cobra.Command{
	Use:   "secret [name]",
	Short: "list secrets entries",
	Long:  "use this command to list the entries of specific secret",
	Run: func(cmd *cobra.Command, args []string) {
		secName := ""

		if len(args) >= 1 {
			secName = args[0]
		}

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

		sec, err := EnsureSecret([]fn.Option{
			fn.MakeOption("secretName", secName),
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printSecret(sec, cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func SelectSecret(options ...fn.Option) (*server.Secret, error) {

	secrets, err := server.ListSecrets(options...)
	if err != nil {
		return nil, err
	}

	if len(secrets) == 0 {
		return nil, errors.New("no secret found")
	}

	secret, err := fzf.FindOne(
		secrets,
		func(sec server.Secret) string {
			return sec.DisplayName
		},
	)

	if err != nil {
		return nil, err
	}

	return secret, nil
}

func EnsureSecret(options ...fn.Option) (*server.Secret, error) {
	secName := fn.GetOption(options, "secretName")

	if secName != "" {
		return server.GetSecret(options...)
	}

	secret, err := SelectSecret(options...)

	if err != nil {
		return nil, err
	}

	return secret, nil
}

func printSecret(secret *server.Secret, cmd *cobra.Command) error {
	outputFormat := cmd.Flag("output").Value.String()

	switch outputFormat {
	case "json":
		configBytes, err := json.Marshal(secret.StringData)
		if err != nil {
			return err
		}
		fn.Println(string(configBytes))

	case "yaml", "yml":
		configBytes, err := yaml.Marshal(secret.StringData)
		if err != nil {
			return err
		}
		fn.Println(string(configBytes))

	default:
		header := table.Row{
			table.HeaderText("key"),
			table.HeaderText("value"),
		}
		rows := make([]table.Row, 0)

		for k, v := range secret.StringData {
			rows = append(rows, table.Row{
				k, v,
			})
		}

		fmt.Println(table.Table(&header, rows))
		table.KVOutput("Showing entries of secret:", secret.Metadata.Name, true)
		table.TotalResults(len(secret.StringData), true)
	}

	return nil
}

func init() {
	secretCmd.Flags().StringP("output", "o", "table", "output format (table|json|yaml)")
}
