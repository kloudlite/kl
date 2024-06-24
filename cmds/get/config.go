package get

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var configCmd = &cobra.Command{
	Use:   "config [name]",
	Short: "list config entries",
	Long:  "use this command to list entries of specific config",
	Run: func(cmd *cobra.Command, args []string) {
		configName := ""

		if len(args) >= 1 {
			configName = args[0]
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

		config, err := EnsureConfig([]fn.Option{
			fn.MakeOption("configName", configName),
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := printConfig(config, cmd); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func SelectConfig(options ...fn.Option) (*server.Config, error) {

	configs, err := server.ListConfigs(options...)
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, errors.New("no configs found")
	}

	config, err := fzf.FindOne(
		configs,
		func(config server.Config) string {
			return config.DisplayName
		},
	)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func EnsureConfig(options ...fn.Option) (*server.Config, error) {
	configName := fn.GetOption(options, "configName")

	if configName != "" {
		return server.GetConfig(options...)
	}

	config, err := SelectConfig(options...)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func printConfig(config *server.Config, cmd *cobra.Command) error {
	outputFormat := cmd.Flag("output").Value.String()

	switch outputFormat {
	case "json":
		configBytes, err := json.Marshal(config.Data)
		if err != nil {
			return functions.Error(err)
		}
		fn.Println(string(configBytes))

	case "yaml", "yml":
		configBytes, err := yaml.Marshal(config.Data)
		if err != nil {
			return functions.Error(err)
		}
		fn.Println(string(configBytes))

	default:
		header := table.Row{
			table.HeaderText("key"),
			table.HeaderText("value"),
		}
		rows := make([]table.Row, 0)

		for k, v := range config.Data {
			rows = append(rows, table.Row{
				k, v,
			})
		}

		fmt.Println(table.Table(&header, rows))

		table.KVOutput("Showing entries of config:", config.Metadata.Name, true)

		table.TotalResults(len(config.Data), true)
	}

	return nil
}

func init() {
	configCmd.Flags().StringP("output", "o", "table", "json | yaml")
}
