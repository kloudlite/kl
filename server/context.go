package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"gopkg.in/yaml.v2"
)

const (
	SessionFileName   string = "kl-session.yaml"
	ExtraDataFileName string = "kl-extra-data.yaml"
	CompleteFileName  string = "kl-completion"
)

type LocalEnv struct {
	Name            string `json:"name"`
	ClusterName     string `json:"clusterName"`
	TargetNamespace string `json:"targetNamespace"`
}

type Session struct {
	Session string `json:"session"`
}

type MainContext struct {
	AccountName string `json:"accountName"`
}

type ExtraData struct {
	BaseUrl      string              `json:"baseUrl"`
	SelectedEnvs map[string]LocalEnv `json:"selectedEnvs"`
}

func WriteCompletionContext() (io.Writer, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(dir, CompleteFileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func GetCompletionContext() (string, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return "", err
	}

	filePath := path.Join(dir, CompleteFileName)
	return filePath, nil
}

func SaveBaseURL(url string) error {
	extraData, err := GetExtraData()
	if err != nil {
		return fn.Error(err)
	}

	extraData.BaseUrl = url
	file, err := yaml.Marshal(extraData)
	if err != nil {
		return fn.Error(err)
	}

	return writeOnUserScope(ExtraDataFileName, file)
}

func GetBaseURL() (string, error) {
	extraData, err := GetExtraData()
	if err != nil {
		return "", err
	}

	return extraData.BaseUrl, nil
}

func SaveExtraData(extraData *ExtraData) error {
	file, err := yaml.Marshal(extraData)
	if err != nil {
		return fn.Error(err)
	}

	return writeOnUserScope(ExtraDataFileName, file)
}

func GetExtraData() (*ExtraData, error) {
	file, err := ReadFile(ExtraDataFileName)
	extraData := ExtraData{}
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			b, err := yaml.Marshal(extraData)
			if err != nil {
				return nil, err
			}

			if err := writeOnUserScope(ExtraDataFileName, b); err != nil {
				return nil, err
			}
		}

		return &extraData, nil
	}

	if err = yaml.Unmarshal(file, &extraData); err != nil {
		return nil, err
	}
	if extraData.SelectedEnvs == nil {
		extraData.SelectedEnvs = make(map[string]LocalEnv)
	}

	return &extraData, nil
}

func getCookieString(options ...fn.Option) (string, error) {

	accName := fn.GetOption(options, "accountName")

	session, err := GetAuthSession()
	if err != nil {
		return "", err
	}

	if session == "" {
		return "", fmt.Errorf("no session found")
	}

	if accName != "" {
		return fmt.Sprintf("kloudlite-account=%s;hotspot-session=%s", accName, session), nil
	}

	return fmt.Sprintf("hotspot-session=%s", session), nil
}

func GetAuthSession() (string, error) {
	file, err := ReadFile(SessionFileName)

	session := Session{}

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			b, err := yaml.Marshal(session)
			if err != nil {
				return "", err
			}

			if err := writeOnUserScope(SessionFileName, b); err != nil {
				return "", err
			}
		}
	}

	if err = yaml.Unmarshal(file, &session); err != nil {
		return "", err
	}

	return session.Session, nil
}

func SaveAuthSession(session string) error {
	file, err := yaml.Marshal(Session{Session: session})
	if err != nil {
		return fn.Error(err)
	}

	return writeOnUserScope(SessionFileName, file)
}

func writeOnUserScope(name string, data []byte) error {
	dir, err := GetConfigFolder()
	if err != nil {
		return fn.Error(err)
	}

	if _, er := os.Stat(dir); errors.Is(er, os.ErrNotExist) {
		er := os.MkdirAll(dir, os.ModePerm)
		if er != nil {
			return er
		}
	}

	filePath := path.Join(dir, name)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fn.Error(err)
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, filePath), nil, false,
		); err != nil {
			return fn.Error(err)
		}
	}

	return nil
}

func ReadFile(name string) ([]byte, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, err
	}

	filePath := path.Join(dir, name)

	if _, er := os.Stat(filePath); errors.Is(er, os.ErrNotExist) {
		return nil, fmt.Errorf("file not found")
	}

	file, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func SetEnvAtPath(path string, env *LocalEnv) error {
	extradata, err := GetExtraData()
	if err != nil {
		return fn.Error(err)
	}
	extradata.SelectedEnvs[path] = *env
	return SaveExtraData(extradata)
}

func EnvAtPath(path string) (*LocalEnv, error) {
	extradata, err := GetExtraData()
	if err != nil {
		return nil, err
	}

	env, ok := extradata.SelectedEnvs[path]
	if !ok {
		return nil, fmt.Errorf("no env found for path %s", path)
	}

	if !ok {
		return nil, fmt.Errorf("no env found for path %s, please choose using 'kl use env'", path)

		// todo: fetch default env from kl.yml and server

		// klFile, err := klfile.GetKlFile(path + "/kl.yml")
		// if err != nil {
		// 	return nil, err
		// }
		//
		// // e, err := server.GetEnvironment(klFile.DefaultEnv)
		// // if err != nil {
		// // 	return nil, err
		// // }
		// //
		// // env = Env{
		// // 	Name:            e.Metadata.Name,
		// // 	ClusterName:     e.ClusterName,
		// // 	TargetNamespace: e.Spec.TargetNamespace,
		// // }
		//
		// env = Env(klFile.DefaultEnv)
	}
	return &env, nil
}

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

func GetConfigFolder() (configFolder string, err error) {
	if os.Getenv("IN_DEV_BOX") == "true" {
		return "/.cache/kl", nil
	}
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
		if err = fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, configPath), nil, false,
		); err != nil {
			return "", err
		}
	}

	return configPath, nil
}
