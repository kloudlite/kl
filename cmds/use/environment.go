package use

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "use env",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := selectEnvironment(); err != nil {
				functions.PrintError(err)
				return
			}

			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			return
		}

		e, err := server.GetEnvironment(args[0])
		if err != nil {
			functions.PrintError(err)
			return
		}

		err = utils.SetEnvAtPath(cwd, utils.Env{
			Name:            e.Metadata.Name,
			ClusterName:     e.ClusterName,
			TargetNamespace: e.Spec.TargetNamespace,
		})
		if err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func selectEnvironment() error {
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return functions.Error(err)
	}

	envs, err := server.ListEnvs(functions.Option{
		Key:   "accountName",
		Value: klFile.AccountName,
	})
	if err != nil {
		return functions.Error(err)
	}

	selectedEnv, err := fzf.FindOne(envs, func(item server.Env) string {
		return item.Metadata.Name
	}, fzf.WithPrompt("Select an environment: "))
	if err != nil {
		return functions.Error(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return functions.Error(err)
	}

	if err = utils.SetEnvAtPath(cwd, utils.Env{
		Name:            selectedEnv.Metadata.Name,
		ClusterName:     selectedEnv.ClusterName,
		TargetNamespace: selectedEnv.Spec.TargetNamespace,
	}); err != nil {
		return functions.Error(err)
	}

	functions.Log(fmt.Sprintf("switched to %s environment", selectedEnv.Metadata.Name))
	return nil
}
