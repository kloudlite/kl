package k3s

import (
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/domain/fileclient"
)

type Cluster struct {
	Name        string
	AccountName string
	Status      string
}

type K3SClient interface {
	CreateCluster(accName, name string) error
	StartCluster(name string) error
	StopCluster(name string) error
	RemoveCluster(name string) error
	ListClusters() ([]Cluster, error)
}

type K3sClientImpl struct {
	fc      fileclient.FileClient
	dClient *dockerclient.Client
}

func NewK3sClient() (K3SClient, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	fc, err := fileclient.New()
	if err != nil {
		return nil, err
	}
	return &K3sClientImpl{
		fc:      fc,
		dClient: cli,
	}, nil
}
