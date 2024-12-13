package env

import (
	"github.com/kloudlite/kl/domain/clients"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "suspend environment",
	Run: func(_ *cobra.Command, _ []string) {
		if err := envPause(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func envPause() error {

	fc := clients.File
	if err != nil {
		return err
	}

	apic := clients.Api
	if err != nil {
		return err
	}

	env, err := apic.EnsureEnv()
	if err != nil {
		return err
	}

	team, err := fc.CurrentTeamName()
	if err != nil {
		return err
	}

	e, err := apic.GetEnvironment(team, env)
	if err != nil {
		return err
	}

	if !e.Status.IsReady {
		fn.Warnf("environment is not ready, please wait for it to be ready")
		return nil
	}

	if err := apic.UpdateEnvironment(team, e, true); err != nil {
		return err
	}
	fn.Log(text.Bold(text.Green("\nEnvironment suspended successfully\n")))
	return nil
}
