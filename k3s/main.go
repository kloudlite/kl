package k3s

import (
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

type K3sClient interface {
	CreateClustersTeams(name string) error
	EnsureKloudliteNetwork() error
	StartAppInterceptService(ports []apiclient.ServicePort, toStart bool) error
	EnsureImage(i string) error
	RestartWgProxyContainer() error
	RemoveAllIntercepts() error
	DeletePods() error
	CheckK3sRunningLocally() (bool, error)
	RemoveClusterVolume(clusterName string) error
	CheckK3sServerRunning() (string, error)
	Exec(script string) ([]byte, error)
}

type client struct {
	c    *dockerclient.Client
	apic apiclient.ApiClient
	fc   fileclient.FileClient
	cmd  *cobra.Command
}

func NewClient(cmd *cobra.Command) (K3sClient, error) {
	ac := clients.Api

	c, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &client{
		c:    c,
		apic: ac,
		fc:   ac.GetFClient(),
		cmd:  cmd,
	}, nil
}
