package use

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "env",
	Short: "Switch to a different environment",
	Run: func(cmd *cobra.Command, args []string) {
		if err := switchEnv(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func switchEnv(cmd *cobra.Command, args []string) error {

	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	apic, err := apiclient.New()
	if err != nil {
		return err
	}

	//TODO: add changes to the klbox-hash file
	// envName := fn.ParseStringFlag(cmd, "envname")

	klFile, err := fc.GetKlFile("")
	if err != nil {
		return err
	}

	env, err := selectEnv(apic, fc)
	if err != nil {
		return err
	}

	if klFile.DefaultEnv == "" {
		klFile.DefaultEnv = env.Metadata.Name
		if err := fc.WriteKLFile(*klFile); err != nil {
			return err
		}
	}
	fn.Log(text.Bold(text.Green("\nSelected Environment:")),
		text.Blue(fmt.Sprintf("\n%s (%s)", env.DisplayName, env.Metadata.Name)),
	)

	wpath, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := hashctrl.SyncBoxHash(apic, fc, wpath); err != nil {
		return err
	}

	c, err := boxpkg.NewClient(cmd, nil)
	if err != nil {
		return err
	}

	if err := c.ConfirmBoxRestart(); err != nil {
		return err
	}

	return nil
}

func init() {
	switchCmd.Aliases = append(switchCmd.Aliases, "switch")

	switchCmd.Flags().StringP("envname", "e", "", "environment name")
	switchCmd.Flags().StringP("team", "a", "", "team name")
}

func selectEnv(apic apiclient.ApiClient, fc fileclient.FileClient) (*apiclient.Env, error) {

	err := apic.RemoveAllIntercepts()
	if err != nil {
		return nil, functions.NewE(err)
	}

	//k3sClient, err := k3s.NewClient()
	//if err != nil {
	//	return nil, functions.NewE(err)
	//}
	//if err = k3sClient.RemoveAllIntercepts(); err != nil {
	//	return nil, functions.NewE(err)
	//}

	persistSelectedEnv := func(env fileclient.Env) error {
		err := fc.SelectEnv(env)
		if err != nil {
			return functions.NewE(err)
		}
		return nil
	}

	currentTeam, err := fc.CurrentTeamName()
	if err != nil {
		return nil, functions.NewE(err)
	}

	envs, err := apic.ListEnvs(currentTeam)
	if err != nil {
		return nil, functions.NewE(err)
	}

	oldEnv, _ := apic.EnsureEnv()

	//env, err := fzf.FindOne(
	//	envs,
	//	func(env apiclient.Env) string {
	//		if env.ClusterName == "" {
	//			return fmt.Sprintf("%s -> %s (template)", env.DisplayName, env.Metadata.Name)
	//		}
	//		return fmt.Sprintf("%s -> %s ", env.DisplayName, env.Metadata.Name)
	//	},
	//	fzf.WithPrompt("Select Environment > "),
	//)

	env, err := fzf.FindOne(
		envs,
		func(env apiclient.Env) string {
			displayName := fmt.Sprintf("%-40s", env.DisplayName)
			name := fmt.Sprintf("%-30s", env.Metadata.Name)

			if env.ClusterName == "" {
				name := fmt.Sprintf("%-30s", fmt.Sprintf("%s (template)", env.Metadata.Name))
				return fmt.Sprintf("%s %s", name, displayName)
			}
			return fmt.Sprintf("%s %s", name, displayName)
		},
		fzf.WithPrompt("Select Environment > "),
	)

	if err != nil {
		return nil, functions.NewE(err)
	}

	if err := persistSelectedEnv(fileclient.Env{
		Name: env.Metadata.Name,
		SSHPort: func() int {
			if oldEnv == nil {
				return 0
			}
			return oldEnv.SSHPort
		}(),
	}); err != nil {
		return nil, functions.NewE(err)
	}

	return env, nil
}
