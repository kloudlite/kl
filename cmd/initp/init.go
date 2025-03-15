package initp

import (
	"os"
	"path"
	"path/filepath"

	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/domain/utils"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/pkg/ui/text"

	"github.com/spf13/cobra"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "initialize a kl-config file",
	Long:  `use this command to initialize a kl-config file`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := handleInit(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func handleInit() error {

	isOnlyPkgMode := utils.IsOnlyPkgMode()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	files := []string{"kl.yml", "kl.yaml"}

	configPath := ""
	for _, name := range files {
		fi, err := os.Stat(filepath.Join(cwd, name))
		if err != nil {
			continue
		}
		if fi.IsDir() {
			continue
		}
		configPath = filepath.Join(cwd, name)
		break
	}

	if configPath != "" {
		fn.Printf(text.Yellow("workspace is already initilized. Do you want to override? [y/N]: "))
		if !fn.Confirm("Y", "N") {
			return fn.Errorf("file initialization aborted")
		}
	}

	team := ""

	if !isOnlyPkgMode {
		fc := clients.File

		team, err = fc.GetTeam()
		if err != nil {
			return err
		}
	}

	newKlFile := fileclient.KLFileType{
		Version:  "v1",
		TeamName: team,
		Packages: []string{"neovim", "git"},
	}

	wd, _ := os.Getwd()
	configFolder, err := fileclient.GetKlPath()
	if err == nil && path.Dir(configFolder) != wd {
		return fn.Errorf("current working directory is not same as config folder, please change your working directory to %s", path.Dir(configFolder))
	}

	if err := newKlFile.Save(); err != nil {
		fn.PrintError(err)
	} else {
		fn.Printf(text.Green("workspace initialized successfully.\n"))
	}

	return nil
}

func init() {
}
