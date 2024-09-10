package apiclient

import (
	"fmt"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"os"
)

type Cluster struct {
	ClusterToken   string          `json:"clusterToken"`
	Name           string          `json:"name"`
	InstallCommand *InstallCommand `json:"installCommand"`
	Metadata       struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

type InstallCommand struct {
	ChartRepo    string `json:"chart-repo"`
	ChartVersion string `json:"chart-version"`
	CRDsURL      string `json:"crds-url"`
	HelmValues   struct {
		AccountName           string `json:"accountName"`
		ClusterName           string `json:"clusterName"`
		ClusterToken          string `json:"clusterToken"`
		KloudliteDNSSuffix    string `json:"kloudliteDNSSuffix"`
		MessageOfficeGRPCAddr string `json:"messageOfficeGRPCAddr"`
	} `json:"helm-values"`
}

func (apic *apiClient) GetClusterConfig(account string) (*fileclient.AccountClusterConfig, error) {
	clusterConfig, err := apic.fc.GetClusterConfig(account)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if clusterConfig == nil {
		forAccount, err := createClusterForAccount()
		if err != nil {
			return nil, fn.NewE(err)
		}
		config := fileclient.AccountClusterConfig{
			ClusterToken: forAccount.ClusterToken,
			ClusterName:  forAccount.Metadata.Name,
			InstallCommand: fileclient.InstallCommand{
				ChartRepo:    forAccount.InstallCommand.ChartRepo,
				ChartVersion: forAccount.InstallCommand.ChartVersion,
				CRDsURL:      forAccount.InstallCommand.CRDsURL,
				HelmValues: fileclient.InstallHelmValues{
					AccountName:           forAccount.InstallCommand.HelmValues.AccountName,
					ClusterName:           forAccount.InstallCommand.HelmValues.ClusterName,
					ClusterToken:          forAccount.InstallCommand.HelmValues.ClusterToken,
					KloudliteDNSSuffix:    forAccount.InstallCommand.HelmValues.KloudliteDNSSuffix,
					MessageOfficeGRPCAddr: forAccount.InstallCommand.HelmValues.MessageOfficeGRPCAddr,
				},
			},
		}
		err = apic.fc.SetClusterConfig(account, &config)
		if err != nil {
			return nil, fn.NewE(err)
		}
		clusterConfig = &config
	}
	return clusterConfig, nil
}

func getClusterName(clusterName string) (*CheckName, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_infraCheckNameAvailability", map[string]any{
		"resType": ClusterType,
		"name":    clusterName,
	}, &cookie)
	if err != nil {
		return nil, fn.NewE(err)
	}

	if fromResp, err := GetFromResp[CheckName](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}

func createCluster(clusterName string) (*Cluster, error) {
	cn, err := getClusterName(clusterName)
	if err != nil {
		return nil, fn.NewE(err)
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, fn.NewE(err)
	}

	dn := clusterName
	if !cn.Result {
		if len(cn.SuggestedNames) == 0 {
			return nil, fmt.Errorf("no suggested names for cluster %s", clusterName)
		}

		dn = cn.SuggestedNames[0]
	}

	fn.Logf("creating new cluster %s\n", dn)
	respData, err := klFetch("cli_createClusterReference", map[string]any{
		"cluster": map[string]any{
			"metadata":    map[string]string{"name": dn},
			"displayName": dn,
			"visibility":  map[string]string{"mode": "private"},
		},
	}, &cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to create vpn: %s", err.Error())
	}
	d, err := GetFromResp[Cluster](respData)
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err = klFetch("cli_clusterReferenceInstructions", map[string]any{
		"name": d.Metadata.Name,
	}, &cookie)

	if err != nil {
		return nil, fn.NewE(err)
	}

	instruction, err := GetFromResp[InstallCommand](respData)
	if err != nil {
		return nil, fn.NewE(err)
	}

	d.InstallCommand = instruction
	return d, nil
}

func createClusterForAccount() (*Cluster, error) {
	clusterName, err := os.Hostname()
	if err != nil {
		return nil, fn.NewE(err)
	}
	checkNames, err := getClusterName(clusterName)
	if err != nil {
		return nil, fn.NewE(err)
	}
	if !checkNames.Result {
		if len(checkNames.SuggestedNames) == 0 {
			return nil, fmt.Errorf("no suggested names for device %s", clusterName)
		}
		clusterName = checkNames.SuggestedNames[0]
	}
	cluster, err := createCluster(clusterName)
	if err != nil {
		return nil, fn.NewE(err)
	}
	return cluster, nil
}
