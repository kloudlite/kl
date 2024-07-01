package fileclient

import "github.com/kloudlite/kl/pkg/functions"

type fclient struct {
	configPath string
}

type FileClient interface {
	CurrentAccountName() (string, error)
	Logout() error
	GetVpnAccountConfig(account string) (*AccountVpnConfig, error)
	SetVpnAccountConfig(account string, config *AccountVpnConfig) error

	AddCluster(accName, clusterName string) error
	GetCluster(clusterName string) (*Cluster, error)
	DeleteCluster(clusterName string) error
	Clusters() ([]*Cluster, error)

	WriteKLFile(fileObj KLFileType) error
	GetKlFile(filePath string) (*KLFileType, error)
}

func New() (FileClient, error) {
	configPath, err := GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}

	return &fclient{
		configPath: configPath,
	}, nil
}
