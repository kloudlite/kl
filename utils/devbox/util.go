package devbox

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/kloudlite/kl2/pkg/functions"
)

func GetUserHomeDir() (string, error) {
	if runtime.GOOS == "windows" {
		return xdg.Home, nil
	}

	if euid := os.Geteuid(); euid == 0 {
		username, ok := os.LookupEnv("SUDO_USER")
		if !ok {
			return "", errors.New("failed to get sudo user name")
		}

		oldPwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		sp := strings.Split(oldPwd, "/")

		for i := range sp {
			if sp[i] == username {
				return path.Join("/", path.Join(sp[:i+1]...)), nil
			}
		}

		return "", errors.New("failed to get home path of sudo user")
	}

	userHome, ok := os.LookupEnv("HOME")
	if !ok {
		return "", errors.New("failed to get home path of user")
	}

	return userHome, nil
}

func userOwn(fpath string) error {
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := functions.ExecCmd(
			fmt.Sprintf("chown -R %s %s", usr, filepath.Dir(fpath)), nil, false,
		); err != nil {
			return err
		}
	}

	return nil
}

func writeOnUserScope(fpath string, data []byte) error {
	if err := os.WriteFile(fpath, data, 0o644); err != nil {
		return err
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := functions.ExecCmd(
			fmt.Sprintf("chown -R %s %s", usr, filepath.Dir(fpath)), nil, false,
		); err != nil {
			return err
		}
	}

	return nil
}

func getConfigFolder() (configFolder string, err error) {
	homePath, err := GetUserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := path.Join(homePath, ".cache", ".kl")

	// ensuring the dir is present
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return "", err
	}

	// ensuring user permission on created dir
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = functions.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, configPath), nil, false,
		); err != nil {
			return "", err
		}
	}

	return configPath, nil
}
