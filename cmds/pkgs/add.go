package pkgs

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/envhash"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/kloudlite/kl/utils/packages"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add new package",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := klfile.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}
		if slices.Contains(config.Packages, args[0]) {
			return
		}
		for i, p := range config.Packages {
			splits := strings.Split(p, "@")
			inputSplits := strings.Split(args[0], "@")
			if splits[0] == inputSplits[0] {
				config.Packages = append(config.Packages[:i], config.Packages[i+1:]...)
				break
			}
		}
		config.Packages = append(config.Packages, args[0])
		realPkgs, err := packages.SyncLockfileWithNewConfig(*config)
		if err != nil {
			fn.PrintError(err)
			return
		}
		oswd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		installCommand := []string{"/home/kl/.nix-profile/bin/nix", "shell"}
		for _, pkg := range realPkgs {
			installCommand = append(installCommand, pkg)
		}
		installCommand = append(installCommand, "--command", "echo", fmt.Sprintf("Installed %s", args[0]))
		if !(os.Getenv("IN_DEV_BOX") == "true") {
			devbox.Start(oswd, config)
			exitCode, err := devbox.Exec(oswd, installCommand, nil)
			if err != nil {
				fn.PrintError(err)
				return
			} else if exitCode != 0 {
				fn.PrintError(errors.New("failed to install package"))
				return
			}
		} else {
			err = fn.ExecCmd(strings.Join(installCommand, " "), nil, false)
			if err != nil {
				fn.PrintError(err)
				return
			}
		}
		err = klfile.WriteKLFile(*config)
		if err != nil {
			fn.PrintError(err)
			return
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

		if err := envhash.SyncBoxHash(env.Name, cwd, config); err != nil {
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
				err = devbox.Stop(cwd)
				if err != nil {
					fn.PrintError(err)
					return
				}
				err = devbox.Start(cwd, config)
				if err != nil {
					fn.PrintError(err)
					return
				}
			}
		}
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "name of the package to install")
	addCmd.Flags().BoolP("verbose", "v", false, "name of the package to install")
}
