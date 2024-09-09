package apiclient

import "github.com/kloudlite/kl/domain/fileclient"

func (apic *apiClient) GetClusterConfig(account string) (*fileclient.AccountClusterConfig, error) {
	clusterConfig, err := apic.fc.GetClusterConfig(account)
	if err != nil {
		return nil, err
	}
	// TODO: check if cluster config is valid and if not, create it
	return clusterConfig, nil
}
