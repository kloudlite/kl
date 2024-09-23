package fileclient

import (
	"errors"
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"

	"sigs.k8s.io/yaml"
)

const (
	SessionFileName   string = "kl-session.yaml"
	ExtraDataFileName string = "kl-extra-data.yaml"
	CompleteFileName  string = "kl-completion"
	DeviceFileName    string = "kl-device.yaml"
	WGConfigFileName  string = "kl-wg.yaml"
)

type Keys struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type WGConfig struct {
	UUID      string `json:"uuid"`
	Host      Keys   `json:"host"`
	WorkSpace Keys   `json:"workspace"`
	Proxy     Keys   `json:"wg-proxy"`
}

type Env struct {
	Name    string `json:"name"`
	SSHPort int    `json:"sshPort"`
}

type Session struct {
	Session string `json:"session"`
}

type MainContext struct {
	AccountName string `json:"accountName"`
}

type DeviceContext struct {
	DisplayName string `json:"display_name"`
	DeviceName  string `json:"device_name"`
}

type InfraContext struct {
	Name        string `json:"name"`
	AccountName string `json:"accountName"`
	ClusterName string `json:"ClusterName"`
	DeviceName  string `json:"deviceName"`
}

type InfraContexts struct {
	InfraContexts map[string]*InfraContext `json:"infraContexts"`
	ActiveContext string                   `json:"activeContext"`
}

type ExtraData struct {
	BaseUrl      string          `json:"baseUrl"`
	SelectedEnvs map[string]*Env `json:"selectedEnvs"`
}

func GetUserHomeDir() (string, error) {
	if runtime.GOOS == "windows" {
		return xdg.Home, nil
	}

	if euid := os.Geteuid(); euid == 0 {
		username, ok := os.LookupEnv("SUDO_USER")
		if !ok {
			return "", functions.Error("failed to get sudo user name")
		}

		oldPwd, err := os.Getwd()
		if err != nil {
			return "", functions.NewE(err)
		}

		sp := strings.Split(oldPwd, "/")

		for i := range sp {
			if sp[i] == username {
				return path.Join("/", path.Join(sp[:i+1]...)), nil
			}
		}

		return "", functions.Error("failed to get home path of sudo user")
	}

	userHome, ok := os.LookupEnv("HOME")
	if !ok {
		return "", functions.Error("failed to get home path of user")
	}

	return userHome, nil
}

func GetConfigFolder() (configFolder string, err error) {
	if envclient.InsideBox() {
		return path.Join("/.cache", "/kl"), nil
	}

	homePath, err := GetUserHomeDir()
	if err != nil {
		return "", functions.NewE(err)
	}

	configPath := path.Join(homePath, ".cache", ".kl")

	// ensuring the dir is present
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return "", functions.NewE(err)
	}

	// ensuring user permission on created dir
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, configPath), nil, false,
		); err != nil {
			return "", functions.NewE(err)
		}
	}

	return configPath, nil
}

func SaveBaseURL(url string) error {
	extraData, err := GetExtraData()
	if err != nil {
		return functions.NewE(err)
	}

	extraData.BaseUrl = url
	file, err := yaml.Marshal(extraData)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(ExtraDataFileName, file)
}

func GetBaseURL() (string, error) {
	extraData, err := GetExtraData()
	if err != nil {
		return "", functions.NewE(err)
	}

	return extraData.BaseUrl, nil
}

func SaveExtraData(extraData *ExtraData) error {
	file, err := yaml.Marshal(extraData)
	if err != nil {
		return functions.NewE(err)
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
				return nil, functions.NewE(err)
			}

			if err := writeOnUserScope(ExtraDataFileName, b); err != nil {
				return nil, functions.NewE(err)
			}
		}

		return &extraData, nil
	}

	if err = yaml.Unmarshal(file, &extraData); err != nil {
		return nil, functions.NewE(err)
	}

	return &extraData, nil
}

func (fc *fclient) SetDevice(device *DeviceContext) error {
	file, err := yaml.Marshal(device)
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(DeviceFileName, file)
}

func (fc *fclient) GetDevice() (*DeviceContext, error) {
	file, err := ReadFile(DeviceFileName)
	device := DeviceContext{}

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			b, err := yaml.Marshal(device)

			if err != nil {
				return nil, functions.NewE(err)
			}

			if err := writeOnUserScope(DeviceFileName, b); err != nil {
				return nil, functions.NewE(err)
			}
		}

		return &device, nil
	}

	if err = yaml.Unmarshal(file, &device); err != nil {
		return nil, functions.NewE(err)
	}

	return &device, nil
}

func GenerateWireGuardKeys() (wgtypes.Key, wgtypes.Key, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, wgtypes.Key{}, fmt.Errorf("failed to generate private key: %w", err)
	}
	publicKey := privateKey.PublicKey()

	return privateKey, publicKey, nil
}

func (fc *fclient) GetWGConfig() (*WGConfig, error) {
	file, err := ReadFile(WGConfigFileName)
	if err != nil {
		u, err := uuid.NewV4()
		if err != nil {
			return nil, fn.NewE(err)
		}
		wgProxyPrivateKey, wgProxyPublicKey, err := GenerateWireGuardKeys()
		if err != nil {
			return nil, fn.NewE(err)
		}
		hostPrivateKey, hostPublicKey, err := GenerateWireGuardKeys()
		if err != nil {
			return nil, fn.NewE(err)
		}
		workSpacePrivateKey, workSpacePublicKey, err := GenerateWireGuardKeys()
		if err != nil {
			return nil, fn.NewE(err)
		}
		wgConfig := WGConfig{
			UUID: u.String(),
			Proxy: Keys{
				PrivateKey: wgProxyPrivateKey.String(),
				PublicKey:  wgProxyPublicKey.String(),
			},
			Host: Keys{
				PrivateKey: hostPrivateKey.String(),
				PublicKey:  hostPublicKey.String(),
			},
			WorkSpace: Keys{
				PrivateKey: workSpacePrivateKey.String(),
				PublicKey:  workSpacePublicKey.String(),
			},
		}
		file, err := yaml.Marshal(wgConfig)
		if err != nil {
			return nil, fn.NewE(err)
		}
		if err := writeOnUserScope(WGConfigFileName, file); err != nil {
			return nil, fn.NewE(err)
		}
	}

	wgConfig := WGConfig{}

	if err = yaml.Unmarshal(file, &wgConfig); err != nil {
		return nil, fn.NewE(err)
	}

	return &wgConfig, nil
}

func GetCookieString(options ...fn.Option) (string, error) {

	accName := fn.GetOption(options, "accountName")

	session, err := GetAuthSession()
	if err != nil {
		return "", functions.NewE(err)
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
				return "", functions.NewE(err)
			}

			if err := writeOnUserScope(SessionFileName, b); err != nil {
				return "", functions.NewE(err)
			}
		}
	}

	if err = yaml.Unmarshal(file, &session); err != nil {
		return "", functions.NewE(err)
	}

	return session.Session, nil
}

func SaveAuthSession(session string) error {
	file, err := yaml.Marshal(Session{Session: session})
	if err != nil {
		return functions.NewE(err)
	}

	return writeOnUserScope(SessionFileName, file)
}

func writeOnUserScope(name string, data []byte) error {
	dir, err := GetConfigFolder()
	if err != nil {
		return functions.NewE(err)
	}

	if _, er := os.Stat(dir); errors.Is(er, os.ErrNotExist) {
		er := os.MkdirAll(dir, os.ModePerm)
		if er != nil {
			return er
		}
	}

	filePath := path.Join(dir, name)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return functions.NewE(err)
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown %s %s", usr, filePath), nil, false,
		); err != nil {
			return functions.NewE(err)
		}
	}

	return nil
}

func ReadFile(name string) ([]byte, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}

	filePath := path.Join(dir, name)

	if _, er := os.Stat(filePath); errors.Is(er, os.ErrNotExist) {
		return nil, fmt.Errorf("file not found")
	}

	file, err := os.ReadFile(filePath)

	if err != nil {
		return nil, functions.NewE(err)
	}

	return file, nil
}
