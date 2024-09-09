package apiclient

import "github.com/kloudlite/kl/domain/fileclient"

func (apic *apiClient) GetClusterConfig(account string) (*fileclient.AccountClusterConfig, error) {
	clusterConfig, err := apic.fc.GetClusterConfig(account)
}
