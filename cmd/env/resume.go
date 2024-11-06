package env

import (
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "resume suspended env",
	Run: func(cmd *cobra.Command, args []string) {
		if err := envResume(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func envResume() error {
	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	apic, err := apiclient.New()
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

	e, err := apic.GetEnvironment(team, env.Name)
	if err != nil {
		return err
	}

	if !e.Status.IsReady {
		fn.Warnf("environment is not ready, please wait for it to be ready")
		return nil
	}

	if err := apic.UpdateEnvironment(team, e, false); err != nil {
		return err
	}
	fn.Log(text.Bold(text.Green("\nEnvironment resumed successfully\n")))
	return nil
}
