package shell

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func createSet[T comparable](v []T) []T {
	m := make(map[T]struct{}, len(v))
	result := make([]T, 0, len(v))

	for i := range v {
		if _, ok := m[v[i]]; !ok {
			m[v[i]] = struct{}{}
			result = append(result, v[i])
		}
	}

	return result
}

func installPackage(pkgs ...string) (path string, err error) {

	c := exec.Command("sh", "-c", fmt.Sprintf("nix shell %s --command printenv PATH", strings.Join(pkgs, " ")))
	if flags.IsVerbose {
		fn.Log(c.String())
	}

	b := new(bytes.Buffer)
	c.Stdout = b
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		return "", err
	}

	return b.String(), nil
}

type ShellArgs struct {
	Shell     string
	EnvVars   []string
	Packages  []string
	Libraries []string
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func envMapToSlice(env map[string]string) []string {
	var result []string
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func envSliceToMap(env []string) map[string]string {
	result := make(map[string]string, len(env))
	for _, kv := range env {
		key, val, found := strings.Cut(kv, "=")
		if !found {
			return nil
		}
		result[key] = val
	}
	return result
}

func NixShell(ctx context.Context, args ShellArgs) error {
	envMap := envSliceToMap(append(os.Environ(), args.EnvVars...))

	path, err := installPackage(args.Packages...)
	if err != nil {
		return fn.NewE(err)
	}

	envMap["PATH"] = strings.TrimSpace(path)

	libPaths := make([]string, 0, len(args.Libraries))
	var includes []string

	for _, lib := range args.Libraries {
		c := exec.CommandContext(ctx, "nix", "eval", lib, "--raw")
		if flags.IsVerbose {
			fn.Log(c.String())
		}

		b, err := c.CombinedOutput()
		if err != nil {
			return fn.NewE(err)
		}

		if pathExists(string(b) + "/lib") {
			libPaths = append(libPaths, string(b)+"/lib")
		}

		if pathExists(string(b) + "/include") {
			includes = append(includes, string(b)+"/include")
		}

		cmd := exec.CommandContext(ctx, "nix-store", "--query", "--references", string(b))

		if flags.IsVerbose {
			fn.Log(cmd.String())
		}

		coutput, err := cmd.CombinedOutput()
		if err != nil {
			return fn.NewE(err)
		}
		lines := strings.Split(string(coutput), "\n")

		for _, line := range lines {
			if len(strings.TrimSpace(line)) > 0 && !strings.Contains(line, "-glibc-") {
				if pathExists(line + "/lib") {
					libPaths = append(libPaths, line+"/lib")
				}
			}
		}
	}

	libPaths = createSet(libPaths)
	includes = createSet(includes)

	envMap["LD_LIBRARY_PATH"] = fmt.Sprintf("%s:%s", strings.Join(libPaths, ":"), os.Getenv("LD_LIBRARY_PATH"))
	envMap["CPATH"] = fmt.Sprintf("%s:%s", strings.Join(includes, ":"), os.Getenv("CPATH"))

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
	c.Env = envMapToSlice(envMap)

	if err := c.Run(); err != nil {
		return fn.NewE(err)
	}

	return nil
}
