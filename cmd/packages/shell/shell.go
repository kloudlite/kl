package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/kloudlite/kl/domain/apiclient"
	daemon_server "github.com/kloudlite/kl/domain/daemon-server"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/nixpkghandler"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "shell",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Shell(cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func Shell(cmd *cobra.Command, args []string) error {
	_, err := exec.LookPath("nix")
	if err != nil {
		return fn.NewE(err, text.Red("nix is not installed. Please install it before using `kl shell`"))
	}

	apic, err := apiclient.New()
	if err != nil {
		return err
	}

	isOnlyPkgMode := func() bool {
		s := apic.GetFClient().GetDataContext()
		_, err := s.GetSession()
		if err != nil {
			return false
		}

		fn.Warn(text.Yellow("session not found, but you can use as pure package manager"))
		return true
	}()

	if !isOnlyPkgMode {
		dclient, err := daemon_server.NewProxyWithService(false)
		if err != nil {
			return err
		}

		if err = dclient.Start(); err != nil {
			return err
		}

		searchDomain, err := apic.GetFClient().GetDataContext().GetSearchDomain()
		if err != nil {
			return err
		}

		if _, err := dclient.SetSearchDomain(searchDomain); err != nil {
			return err
		}

	}

	pc, err := nixpkghandler.New(cmd)
	if err != nil {
		return err
	}

	var envMap, mountMap map[string]string

	ck, err := getCache()
	if err == nil {
		envMap = ck.EnvVars
		mountMap = ck.Mounts
	} else {
		fn.Warn(text.Yellow("cache not found, refetching"))
		envMap, mountMap, err = apic.GetLoadMaps()
		if err != nil {
			fn.Warn(err)
			fn.Warn("ignoring, loading environment variables and mounts.")
			envMap = make(map[string]string)
			mountMap = make(map[string]string)
		}

		ck = &fileclient.CacheKLConfig{
			Mounts:  mountMap,
			EnvVars: envMap,
		}

		if err := setCache(ck); err != nil {
			return err
		}
	}

	kpath, err := apic.GetFClient().GetConfigPath()
	if err != nil {
		return err
	}

	mountpath := path.Join(path.Dir(kpath), ".kl", "mounts")

	if !isOnlyPkgMode {
		if err := mount(mountMap, mountpath); err != nil {
			return err
		}
	}

	if err := pc.SyncLockfile(); err != nil {
		return err
	}

	lockFile, err := apic.GetFClient().GetLockfile()

	pkgs := make([]string, 0, len(lockFile.Packages))
	for _, v := range lockFile.Packages {
		pkgs = append(pkgs, fmt.Sprintf("nixpkgs/%s", v))
	}

	libs := make([]string, 0, len(lockFile.Libraries))
	for _, v := range lockFile.Libraries {
		libs = append(libs, fmt.Sprintf("nixpkgs/%s", v))
	}

	envs := make([]string, 0, len(envMap))
	for k, v := range envMap {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}

	if err := NixShell(cmd, ShellArgs{
		Shell:     os.Getenv("SHELL"),
		EnvVars:   append(envs, "KL_SHELL=true", fmt.Sprintf("kl_mounts=%s", mountpath)),
		Packages:  pkgs,
		Libraries: libs,
		ShellData: ck.ShellData,
	}); err != nil {
		return fn.NewE(err)
	}

	return nil
}
