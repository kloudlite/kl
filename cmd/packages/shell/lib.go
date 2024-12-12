package shell

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

func NixShell(ctx context.Context, args ShellArgs) error {
	ev := append(os.Environ(), args.EnvVars...)

	// f := spinner.Client.UpdateMessage("setting up nix environment...")
	path, err := installPackage(args.Packages...)
	// f()
	if err != nil {
		return fn.NewE(err)
	}

	ev = append(ev, fmt.Sprintf("PATH=%s", strings.TrimSpace(path)))

	libPaths := make([]string, 0, len(args.Libraries))
	var includes []string

	for _, lib := range args.Libraries {
		c := exec.CommandContext(ctx, "nix", "eval", lib, "--raw")
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

		c2 := exec.CommandContext(ctx, "nix-store", "--query", "--references", string(b))
		b2, err := c2.CombinedOutput()
		if err != nil {
			return fn.NewE(err)
		}
		lines := strings.Split(string(b2), "\n")

		for _, line := range lines {
			if len(strings.TrimSpace(line)) > 0 && !strings.Contains(line, "-glibc-") {
				if pathExists(line + "/lib") {
					libPaths = append(libPaths, line+"/lib")
				}
				// if pathExists(line + "/lib64") {
				// 	libPaths = append(libPaths, line+"/lib64")
				// }

				// if pathExists(line + "/include") {
				// 	includes = append(includes, line+"/include")
				// }
			}
		}
	}

	libPaths = createSet(libPaths)
	includes = createSet(includes)

	ev = append(ev, fmt.Sprintf("LD_LIBRARY_PATH=%s:%s", strings.Join(libPaths, ":"), os.Getenv("LD_LIBRARY_PATH")))
	ev = append(ev, fmt.Sprintf("CPATH=%s:%s", strings.Join(includes, ":"), os.Getenv("CPATH")))

	ev = append(ev, "LD_LIBRARY_PATH=")

	shell := args.Shell
	if shell == "" {
		shell = "sh"
	}

	c := exec.Command(shell)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = ev

	if err := c.Run(); err != nil {
		return fn.NewE(err)
	}

	return nil
}
