package fileclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/adrg/xdg"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"

	"sigs.k8s.io/yaml"
)

const (
	SessionFileName                  string = "kl-session.yaml"
	ExtraDataFileName                string = "kl-extra-data.yaml"
	CompleteFileName                 string = "kl-completion"
	DeviceFileName                   string = "kl-device.yaml"
	WGConfigFileName                 string = "kl-wg.yaml"
	WorkspaceWireguardConfigFileName string = "kl-workspace-wg.conf"
	K3sTrackerFileName                      = "k3s-status.json"

	KLWGProxyIp   = "198.18.0.1"
	KLHostIp      = "198.18.0.2"
	KLWorkspaceIp = "198.18.0.3"
	KLWGAllowedIp = "100.64.0.0/10"
	LocalHostIP   = "127.0.0.1"
)

type Keys struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type WGConfig struct {
	UUID      string `json:"uuid"`
	Host      Keys   `json:"host"`
	Workspace Keys   `json:"workspace"`
	Proxy     Keys   `json:"wg-proxy"`
}

type Port struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	TargetPort int    `json:"targetPort"`
}

type K3sTracker struct {
	Compute      bool `json:"compute"`
	Gateway      bool `json:"gateway"`
	WgConnection bool `json:"wgConnection"`
	DeviceRouter struct {
		IP      string `json:"ip"`
		Service struct {
			Spec struct {
				Ports []Port `json:"ports"`
			} `json:"spec"`
		} `json:"service"`
	} `json:"deviceRouter"`
	LastCheckedAt string `json:"lastCheckedAt"`
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
			return "", functions.NewE(err, "failed to get current working directory")
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
	homePath, err := GetUserHomeDir()
	if err != nil {
		return "", functions.NewE(err)
	}

	configPath := path.Join(homePath, ".kl")

	// ensuring the dir is present
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return "", functions.NewE(err, "failed to create config folder")
	}

	// ensuring user permission on created dir
	if _, ok := os.LookupEnv("SUDO_USER"); ok {
		uid, gid, err := fn.GetUidNGid()
		if err != nil {
			return "", err
		}

		if err := os.Chown(configPath, uid, gid); err != nil {
			return "", functions.NewE(err, "failed to change user permission on config folder")
		}
	}

	return configPath, nil
}

func (fc *fclient) SaveBaseURL(url string) error {
	extraData, err := fc.getExtraData()
	if err != nil {
		return functions.NewE(err)
	}

	return extraData.SetBaseUrl(url)
}

func (fc *fclient) GetBaseURL() (string, error) {
	extraData, err := fc.getExtraData()
	if err != nil {
		return "", functions.NewE(err)
	}

	return extraData.GetBaseUrl(), nil
}

func (fc *fclient) SetDevice(device *DeviceData) error {
	file, err := yaml.Marshal(device)
	if err != nil {
		return functions.NewE(err, "failed to marshal device context")
	}

	return writeOnUserScope(DeviceFileName, file)
}

func (fc *fclient) GetDevice() (*DeviceData, error) {
	dData, err := fc.GetDataContext().GetDevice()
	if err != nil {
		return nil, err
	}

	return dData, nil
}

func (c *fclient) GetHostWgConfig() (string, error) {
	config, err := c.GetWGConfig()
	if err != nil {
		return "", fn.NewE(err, "failed to get wg config")
	}

	wgConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32

[Peer]
PublicKey = %s
AllowedIPs = 198.18.0.0/16, %s/32, %s
PersistentKeepalive = 25
Endpoint = %s:33820
`, config.Host.PrivateKey, KLHostIp, config.Proxy.PublicKey, KLWGProxyIp, KLWGAllowedIp, LocalHostIP)
	return wgConfig, nil
}

func (fc *fclient) SetWGConfig(config string) error {
	if err := writeOnUserScope("kl-host-wg.conf", []byte(config)); err != nil {
		return fn.NewE(err, "failed to write wg config")
	}

	return nil
}

func (fc *fclient) generateWGConfig(config *WGConfig) string {
	return fmt.Sprintf(`[Interface]
Address = %s/32
ListenPort = 31820
PrivateKey = %s

[Peer]
PublicKey = %s
AllowedIPs = 198.18.0.0/16, #CLUSTER_GATEWAY_IP/32, #CLUSTER_IP_RANGE
Endpoint = k3s-cluster.local:33820
PersistentKeepalive = 25
`, KLWorkspaceIp, config.Workspace.PrivateKey, config.Proxy.PublicKey)
}

func (fc *fclient) GetWGConfig() (*WGConfig, error) {
	file, err := readFile(WGConfigFileName)
	if err != nil {
		u, err := uuid.NewV4()
		if err != nil {
			return nil, fn.NewE(err, "failed to generate uuid")
		}
		wgProxyPrivateKey, wgProxyPublicKey, err := GenerateWireGuardKeys()
		if err != nil {
			return nil, fn.NewE(err, "failed to generate wg keys")
		}
		hostPrivateKey, hostPublicKey, err := GenerateWireGuardKeys()
		if err != nil {
			return nil, fn.NewE(err, "failed to generate wg keys")
		}
		workSpacePrivateKey, workSpacePublicKey, err := GenerateWireGuardKeys()
		if err != nil {
			return nil, fn.NewE(err, "failed to generate wg keys")
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
			Workspace: Keys{
				PrivateKey: workSpacePrivateKey.String(),
				PublicKey:  workSpacePublicKey.String(),
			},
		}
		file, err := yaml.Marshal(wgConfig)
		if err != nil {
			return nil, fn.NewE(err, "failed to marshal wg config")
		}
		if err := writeOnUserScope(WGConfigFileName, file); err != nil {
			return nil, fn.NewE(err, "failed to write wg config")
		}
		config := fc.generateWGConfig(&wgConfig)
		if err := writeOnUserScope(WorkspaceWireguardConfigFileName, []byte(config)); err != nil {
			return nil, fn.NewE(err, "failed to write wg config")
		}
		return &wgConfig, nil
	}

	wgConfig := WGConfig{}

	if err = yaml.Unmarshal(file, &wgConfig); err != nil {
		return nil, fn.NewE(err, "failed to unmarshal wg config")
	}

	return &wgConfig, nil
}

func (fc *fclient) GetK3sTracker() (*K3sTracker, error) {
	file, err := readFile(K3sTrackerFileName)
	if err != nil {
		return nil, fn.NewE(err, "failed to read k3s tracker")
	}

	tracker := K3sTracker{}

	if err = json.Unmarshal(bytes.Trim(file, "\x00"), &tracker); err != nil {
		return nil, fn.NewE(err, "failed to unmarshal k3s tracker")
	}

	return &tracker, nil
}

func GetCookieString(options ...fn.Option) (string, error) {
	teamName := fn.GetOption(options, "teamName")

	sd, err := getCtxData()
	if err != nil {
		return "", err
	}

	session, err := sd.GetSession()
	if err != nil {
		return "", fn.NewE(err, "unauthorized")
	}

	if teamName != "" {
		return fmt.Sprintf("kloudlite-account=%s;hotspot-session=%s", teamName, session), nil
	}

	return fmt.Sprintf("hotspot-session=%s", session), nil
}

func writeOnUserScope(name string, data []byte) error {
	dir, err := GetConfigFolder()
	if err != nil {
		return functions.NewE(err, "failed to get config folder")
	}

	if _, er := os.Stat(dir); errors.Is(er, os.ErrNotExist) {
		er := os.MkdirAll(dir, os.ModePerm)
		if er != nil {
			return er
		}
	}

	filePath := path.Join(dir, name)

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return functions.NewE(err, "failed to write file")
	}

	if _, ok := os.LookupEnv("SUDO_USER"); ok {
		uid, gid, err := fn.GetUidNGid()
		if err != nil {
			return err
		}

		if err := os.Chown(filePath, uid, gid); err != nil {
			return functions.NewE(err, "failed to change user permission on file")
		}
	}

	return nil
}

func readFile(name string) ([]byte, error) {
	dir, err := GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err, "failed to get config folder")
	}

	filePath := path.Join(dir, name)

	if _, er := os.Stat(filePath); errors.Is(er, os.ErrNotExist) {
		return nil, err
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, functions.NewE(err, "failed to read file")
	}

	return file, nil
}

func GenerateWireGuardKeys() (wgtypes.Key, wgtypes.Key, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, wgtypes.Key{}, fn.Errorf("failed to generate private key: %w", err)
	}
	publicKey := privateKey.PublicKey()

	return privateKey, publicKey, nil
}
