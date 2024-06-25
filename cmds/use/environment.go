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

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "use env",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := selectEnvironment(); err != nil {
				fn.PrintError(err)
				return
			}
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			return
		}
		klFile, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}

		e, err := server.GetEnvironment(klFile.AccountName, args[0])
		if err != nil {
			fn.PrintError(err)
			return
		}

		if os.Getenv("IN_DEV_BOX") == "true" {
			cwd = os.Getenv("KL_WORKSPACE")
		}

		err = server.SetEnvAtPath(cwd, &server.LocalEnv{
			Name:            e.Metadata.Name,
			ClusterName:     e.ClusterName,
			TargetNamespace: e.Spec.TargetNamespace,
		})
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err := envhash.SyncBoxHash(e.Metadata.Name, cwd); err != nil {
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
				if err = devbox.Start(cwd); err != nil {
					fn.PrintError(err)
					return
				}
			}
		}
	},
}

func selectEnvironment() error {
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return fn.Error(err)
	}

	envs, err := server.ListEnvs(fn.Option{
		Key:   "accountName",
		Value: klFile.AccountName,
	})
	if err != nil {
		return fn.Error(err)
	}

	selectedEnv, err := fzf.FindOne(envs, func(item server.Env) string {
		return item.Metadata.Name
	}, fzf.WithPrompt("Select an environment: "))
	if err != nil {
		return fn.Error(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fn.Error(err)
	}
	if os.Getenv("IN_DEV_BOX") == "true" {
		cwd = os.Getenv("KL_WORKSPACE")
	}
	if err = server.SetEnvAtPath(cwd, &server.LocalEnv{
		Name:            selectedEnv.Metadata.Name,
		ClusterName:     selectedEnv.ClusterName,
		TargetNamespace: selectedEnv.Spec.TargetNamespace,
	}); err != nil {
		return fn.Error(err)
	}

	fn.Log(fmt.Sprintf("switched to %s environment", selectedEnv.Metadata.Name))
	if err := envhash.SyncBoxHash(selectedEnv.Metadata.Name, cwd); err != nil {
		return err
	}
	if !(os.Getenv("IN_DEV_BOX") == "true") {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		_, err = devbox.ContainerAtPath(cwd)
		if err != nil && err.Error() == devbox.NO_RUNNING_CONTAINERS {
			return nil
		} else if err != nil {
			return err
		}
		fn.Printf(text.Yellow("environments may have been updated. to reflect the changes, do you want to restart the container? [Y/n] "))
		if fn.Confirm("Y", "Y") {
			if err = devbox.Stop(cwd); err != nil {
				return err
			}
			if err = devbox.Start(cwd); err != nil {
				return err
			}
		}
	}
	return nil
}
