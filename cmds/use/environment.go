package use

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/envhash"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

func effectiveCWD() (*string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if os.Getenv("IN_DEV_BOX") == "true" {
		cwd = os.Getenv("KL_WORKSPACE")
	}
	return &cwd, nil
}

func getCurrentLocalEnv() (*server.LocalEnv, error) {
	cwd, err := effectiveCWD()
	if err != nil {
		return nil, err
	}
	return server.EnvAtPath(*cwd)
}

func setLocalEnv(localEnv *server.LocalEnv) error {
	cwd, err := effectiveCWD()
	if err != nil {
		return err
	}
	if err := server.SetEnvAtPath(*cwd, localEnv); err != nil {
		return err
	}
	return nil
}

func syncBoxHash(envName string, klFile *klfile.KLFileType) error {
	cwd, err := effectiveCWD()
	if err != nil {
		return err
	}
	if err := envhash.SyncBoxHash(envName, *cwd, klFile); err != nil {
		return err
	}
	return nil
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "use env",
	Run: func(_ *cobra.Command, args []string) {
		currentLocalEnv, err := getCurrentLocalEnv()
		if err != nil {
			fn.PrintError(err)
			return
		}

		envName := ""
		selectedEnv := &server.Env{}

		klFile, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}

		if len(args) > 0 {
			if currentLocalEnv.Name == args[0] {
				fn.Println(fmt.Sprintf("already using %s environment", args[0]))
				return
			}
			envName = args[0]
			selectedEnv, err = server.GetEnvironment(klFile.AccountName, envName)
			if err != nil {
				fn.PrintError(err)
				return
			}
		} else {
			selectedEnv, err = selectEnvironment()
			if err != nil {
				fn.PrintError(err)
				return
			}
			if currentLocalEnv.Name == selectedEnv.Metadata.Name {
				fn.Println(fmt.Sprintf("already using %s environment", selectedEnv.Metadata.Name))
				return
			}
		}

		err = setLocalEnv(&server.LocalEnv{
			Name:            selectedEnv.Metadata.Name,
			ClusterName:     selectedEnv.ClusterName,
			TargetNamespace: selectedEnv.Spec.TargetNamespace,
		})

		if err != nil {
			fn.PrintError(err)
			return
		}

		err = syncBoxHash(selectedEnv.Metadata.Name, klFile)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if !(os.Getenv("IN_DEV_BOX") == "true") {
			cwd, err := os.Getwd()
			if err != nil {
				fn.PrintError(err)
				return
			}
			_, err = devbox.ContainerAtPath(cwd)
			if err != nil && err.Error() == devbox.NO_RUNNING_CONTAINERS {
				return
			} else if err != nil {
				fn.PrintError(err)
				return
			}
			fn.Printf(text.Yellow("environments may have been updated. to reflect the changes, do you want to restart the container? [Y/n] "))
			if fn.Confirm("Y", "Y") {
				if err = devbox.Stop(cwd); err != nil {
					fn.PrintError(err)
					return
				}
				if err = devbox.Start(cwd, klFile); err != nil {
					fn.PrintError(err)
					return
				}
			}
		}
	},
}

func selectEnvironment() (*server.Env, error) {
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return nil, fn.Error(err)
	}
	envs, err := server.ListEnvs(fn.Option{
		Key:   "accountName",
		Value: klFile.AccountName,
	})
	if err != nil {
		return nil, fn.Error(err)
	}
	selectedEnv, err := fzf.FindOne(envs, func(item server.Env) string {
		return item.Metadata.Name
	}, fzf.WithPrompt("Select an environment: "))
	if err != nil {
		return nil, fn.Error(err)
	}
	return selectedEnv, nil
}
