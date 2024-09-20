package clusterpkg

import (
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type clusterclient struct {
	fc      fileclient.FileClient
	apic    apiclient.ApiClient
	cli     *dockerclient.Client
	account string
}

type ClusterClient interface {
	StartClusterForAccount() error
	ExecuteClusterScript(conatinerId string) error
}

func New(fc fileclient.FileClient, apic apiclient.ApiClient) (ClusterClient, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fn.NewE(err)
	}

	account, err := fc.CurrentAccountName()
	if err != nil {
		return nil, fn.NewE(err)
	}

	return &clusterclient{
		fc:      fc,
		apic:    apic,
		cli:     cli,
		account: account,
	}, nil
}
