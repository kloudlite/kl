package runner

import (
	"errors"
	"fmt"
	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var AttachCommand = &cobra.Command{
	Use:   "attach",
	Short: "attach team and environment to the kl-config file",
	Long:  `use this command to attach team and environment to the kl-config file`,
	Run: func(cmd *cobra.Command, args []string) {

		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}

		//if envclient.InsideBox() {
		//	fn.PrintError(fn.Error("cannot re-initialize workspace in dev box"))
		//	return
		//}

		//if _, err = fc.GetKlFile(""); err == nil {
		//	fn.Printf(text.Yellow("workspace is already initilized. Do you want to override? [y/N]: "))
		//	if !fn.Confirm("Y", "N") {
		//		return
		//	}
		//} else if !errors.Is(err, confighandler.ErrKlFileNotExists) {
		//	fn.PrintError(err)
		//	return
		//}

		filepath := ""
		if envclient.InsideBox() {
			filepath = "/home/kl/workspace/kl.yml"
		}
		klFile, err := fc.GetKlFile(filepath)
		if err != nil && errors.Is(err, confighandler.ErrKlFileNotExists) {
			klFile = &fileclient.KLFileType{
				Version: "v1",
			}
		}
		if err != nil && !errors.Is(err, confighandler.ErrKlFileNotExists) {
			fn.PrintError(err)
			return
		}

		selectedTeam, err := selectTeam(apic)
		if err != nil {
			fn.PrintError(err)
			return
		} else {
			if selectedEnv, err := selectEnv(apic, fc, *selectedTeam); err != nil {
				fn.PrintError(err)
			} else {
				//newKlFile := fileclient.KLFileType{
				//	TeamName:   *selectedTeam,
				//	DefaultEnv: *selectedEnv,
				//	Version:    "v1",
				//	Packages:   []string{"neovim", "git"},
				//}
				klFile.TeamName = *selectedTeam
				klFile.DefaultEnv = *selectedEnv
				if err := fc.WriteKLFile(*klFile); err != nil {
					fn.PrintError(err)
				} else {
					fn.Printf(text.Green("team name and environment updated.\n"))
				}
			}
		}

		dir, err := os.Getwd()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := hashctrl.SyncBoxHash(apic, fc, dir); err != nil {
			fn.PrintError(err)
			return
		}

		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.ConfirmBoxRestart(); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func selectTeam(apic apiclient.ApiClient) (*string, error) {
	if teams, err := apic.ListTeams(); err == nil {
		if selectedTeam, err := fzf.FindOne(
			teams,
			func(team apiclient.Team) string {
				return team.Metadata.Name + " #" + team.Metadata.Name
			},
			fzf.WithPrompt("select kloudlite team > "),
		); err != nil {
			return nil, fn.NewE(err)
		} else {
			return &selectedTeam.Metadata.Name, nil
		}
	} else {
		return nil, fn.NewE(err)
	}
}

func selectEnv(apic apiclient.ApiClient, fc fileclient.FileClient, teamName string) (*string, error) {
	if envs, err := apic.ListEnvs(teamName); err == nil {
		if selectedEnv, err := fzf.FindOne(
			envs,
			func(env apiclient.Env) string {
				if env.ClusterName == "" {
					return fmt.Sprintf("%s (%s) template-env", env.DisplayName, env.Metadata.Name)
				}
				return fmt.Sprintf("%s (%s) compute-env", env.DisplayName, env.Metadata.Name)
			},
			fzf.WithPrompt("select environment > "),
		); err != nil {
			return nil, fn.NewE(err)
		} else {
			cwd, err := os.Getwd()
			if envclient.InsideBox() {
				cwd = os.Getenv("KL_WORKSPACE")
			}
			env := &fileclient.Env{
				Name: selectedEnv.Metadata.Name,
			}
			err = fc.SelectEnvOnPath(cwd, *env)
			if err != nil {
				return nil, fn.NewE(err)
			}
			if err != nil {
				return nil, fn.NewE(err)
			}
			return &selectedEnv.Metadata.Name, nil
		}
	} else {
		return nil, fn.NewE(err)
	}
}
