package shell

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"al.essio.dev/pkg/shellescape"
	"github.com/adrg/xdg"
	fn "github.com/kloudlite/kl/pkg/functions"
)

// TODO: enhance structure of this file

type name string

const (
	shUnknown name = ""
	shBash    name = "bash"
	shZsh     name = "zsh"
	shKsh     name = "ksh"
	shFish    name = "fish"
	shPosix   name = "posix"
)

//go:embed shellrc.tmpl
var shellrcText string
var shellrcTmpl = template.Must(template.New("shellrc").Parse(shellrcText))

//go:embed shellrc_fish.tmpl
var fishrcText string
var fishrcTmpl = template.Must(template.New("shellrc_fish").Parse(fishrcText))

func getShellOverrides(shell string) (map[string]string, []string, error) {

	shellName := getShellName(shell)
	shellrcpath, err := initShellBinaryFields(shellName)
	shellrc, err := writeKlShellrc(shellName, shellrcpath)
	if err != nil {
		return nil, nil, fn.NewE(err)
	}

	extraEnv, extraArgs := shellRCOverrides(shellrc, name(shellName))

	return extraEnv, extraArgs, nil
}

func initShellBinaryFields(path string) (string, error) {
	base := filepath.Base(path)
	// Login shell
	if base[0] == '-' {
		base = base[1:]
	}
	switch base {
	case "bash":
		return rcfilePath(".bashrc"), nil
	case "zsh":
		if zdotdir := os.Getenv("ZDOTDIR"); zdotdir != "" {
			return filepath.Join(os.ExpandEnv(zdotdir), ".zshrc"), nil
		} else {
			return rcfilePath(".zshrc"), nil
		}
	case "ksh":
		return rcfilePath(".kshrc"), nil
	case "fish":
		return fishConfig(), nil
	case "dash", "ash", "shell":
		resp := os.Getenv("ENV")
		if resp != "" {
			return ".shinit", nil
		}

	}
	return "", fn.Errorf("unknown shell: %s", base)
}

func writeKlShellrc(shell, shellRcPath string) (path string, err error) {
	tmp, err := os.MkdirTemp("", "kl")
	if err != nil {
		return "", fmt.Errorf("create temp dir for shell init file: %v", err)
	}

	userShellrc := []byte{}
	if shellRcPath != "" {
		userShellrc, _ = os.ReadFile(shellRcPath)
	}

	shellrcName := "shellrc"
	if shellRcPath != "" {
		shellrcName = filepath.Base(shellRcPath)
	}
	path = filepath.Join(tmp, shellrcName)
	shellrcf, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("write to shell init file: %v", err)
	}
	defer func() {
		cerr := shellrcf.Close()
		if err == nil {
			err = cerr
		}
	}()

	tmpl := shellrcTmpl
	if shell == string(shFish) {
		tmpl = fishrcTmpl
	}

	wd, _ := os.Getwd()

	err = tmpl.Execute(shellrcf, struct {
		ProjectDir       string
		OriginalInit     string
		OriginalInitPath string
	}{
		ProjectDir:       wd,
		OriginalInit:     string(bytes.TrimSpace(userShellrc)),
		OriginalInitPath: shellRcPath,
	})
	if err != nil {
		return "", fmt.Errorf("execute shellrc template: %v", err)
	}

	fn.Debug("wrote kl shellrc", "path", path)
	return path, nil
}

func shellRCOverrides(shellrc string, shell name) (extraEnv map[string]string, extraArgs []string) {
	switch shell {
	case "sh":
		extraArgs = []string{"--rcfile", shellescape.Quote(shellrc)}
	case shBash:
		extraArgs = []string{"--rcfile", shellescape.Quote(shellrc)}
	case shZsh:
		extraEnv = map[string]string{"ZDOTDIR": shellescape.Quote(filepath.Dir(shellrc))}
	case shKsh, shPosix:
		extraEnv = map[string]string{"ENV": shellescape.Quote(shellrc)}
	case shFish:
		extraArgs = []string{"-C", ". " + shellrc}
	}
	return extraEnv, extraArgs
}

func fishConfig() string {
	return path.Join(xdg.ConfigHome, "fish", "config.fish")
}

func rcfilePath(basename string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, basename)
}

func getShellName(path string) string {
	strings := strings.Split(path, "/")
	if len(strings) == 0 {
		return ""
	}
	return strings[len(strings)-1]
}
