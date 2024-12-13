package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/nixpkghandler"
	"github.com/spf13/cobra"
)

func resetEnvs(envMap map[string]string, env []string) map[string]string {
	if depth, ok := envMap["KL_DEPTH"]; ok {
		i, err := strconv.Atoi(depth)
		if err == nil {
			envMap["KL_DEPTH"] = strconv.Itoa(i + 1)
		} else {
			envMap["KL_DEPTH"] = "1"
		}
	} else {
		envMap["KL_DEPTH"] = "1"
	}

	for _, k := range env {
		if s, ok := envMap[fmt.Sprintf("KL_OLD_%s", k)]; ok {
			envMap[k] = s
		} else {
			envMap[fmt.Sprintf("KL_OLD_%s", k)] = envMap[k]
		}
	}

	return envMap
}

type ShellArgs struct {
	ShellData *fileclient.ShellData
	Shell     string
	EnvVars   []string
	Packages  []string
	Libraries []string
}

func NixShell(cmd *cobra.Command, args ShellArgs) error {
	pc, err := nixpkghandler.New(cmd)
	if err != nil {
		return err
	}

	envMap := fn.EnvSliceToMap(append(os.Environ(), args.EnvVars...))
	envMap = resetEnvs(envMap, []string{"PATH", "LD_LIBRARY_PATH", "CPATH"})

	if args.ShellData != nil {
		m := args.ShellData.Envs
		for k, v := range m {
			envMap[k] = v
		}
	}

	if args.ShellData == nil {
		m, err := pc.EvaluateShell(cmd.Context(), args.Packages, args.Libraries, envMap)
		if err != nil {
			return fn.NewE(err)
		}
		for k, v := range m {
			envMap[k] = v
		}

		fc := clients.File
		if err != nil {
			return fn.NewE(err)
		}

		wc, err := fc.GetWsContext()
		if err != nil {
			return fn.NewE(err)
		}

		if err := wc.SetShellData(&fileclient.ShellData{
			Envs: m,
		}); err != nil {
			return err
		}

	}

	shell := args.Shell
	if shell == "" {
		shell = "sh"
	}

	extraEnv, extraArgs, err := getShellOverrides(shell)
	if err != nil {
		return fn.NewE(err)
	}

	for k, v := range extraEnv {
		envMap[k] = v
	}

	c := exec.Command(shell, extraArgs...)
	if flags.IsVerbose {
		fn.Log(c.String())
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = fn.EnvMapToSlice(envMap)

	if err := c.Run(); err != nil {
		return fn.NewE(err)
	}

	return nil
}
