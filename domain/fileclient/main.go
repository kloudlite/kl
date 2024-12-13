package fileclient

import (
	fn "github.com/kloudlite/kl/pkg/functions"
)

type fclient struct {
	configPath string
	session    Session
}

type FileClient interface {
	SaveBaseURL(url string) error
	GetBaseURL() (string, error)

	GetExtraData() (Extra, error)

	GetHostWgConfig() (string, error)
	GetWGConfig() (*WGConfig, error)
	SetWGConfig(config string) error

	Logout() error

	GetK3sTracker() (*K3sTracker, error)
	GetClusterConfig(team string) (*TeamClusterConfig, error)
	SetClusterConfig(team string, accClusterConfig *TeamClusterConfig) error
	DeleteClusterData(team string) error
	GetDevice() (*DeviceData, error)
	SetDevice(device *DeviceData) error
	GetDataContext() Session

	GetKlFile() (*KLFileType, error)

	GetLockfile() (*Lockfile, error)

	SelectEnv(ev string) error
	CurrentEnv() (string, error)

	GetConfigPath() (string, error)
	GetWsContext() (WsContext, error)

	GetKlFileHash() ([]byte, error)
}

func (c *fclient) GetWsContext() (WsContext, error) {
	return getNewWsContext()
}

func New() (FileClient, error) {
	configPath, err := GetConfigFolder()
	if err != nil {
		return nil, fn.NewE(err)
	}

	sd, _ := getCtxData()

	return &fclient{
		configPath: configPath,
		session:    sd,
	}, nil
}
