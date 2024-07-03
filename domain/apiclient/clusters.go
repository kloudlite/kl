package apiclient

import (
	"fmt"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Cluster struct {
	ID       string   `json:"id"`
	Metadata Metadata `json:"metadata"`
}

type ClusterSetupInstructions struct {
	ChartRepo    string         `json:"chart-repo"`
	ChartVersion string         `json:"chart-version"`
	CRDSUrl      string         `json:"crds-url"`
	HelmValues   map[string]any `json:"helm-values" yaml:"helm-values"`
}

func getClusterName(clusterName string, options ...fn.Option) (*CheckName, error) {
	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_infraCheckNameAvailability", map[string]any{
		"resType": ClusterType,
		"name":    clusterName,
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	if fromResp, err := GetFromResp[CheckName](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, nil
	}
}

func DeleteClusterReference(cn string, options ...fn.Option) error {
	cookie, err := getCookie(options...)
	if err != nil {
		return functions.NewE(err)
	}
	_, err = klFetch("cli_deleteClusterReference", map[string]any{
		"name": cn,
	}, &cookie)
	if err != nil {
		return functions.NewE(err)
	}

	return nil
}

func GetClusterConnectionParams(clusterName string, options ...fn.Option) (*ClusterSetupInstructions, error) {
	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}
	respData, err := klFetch("cli_clusterReferenceInstructions", map[string]any{
		"name": clusterName,
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	c, err := GetFromResp[ClusterSetupInstructions](respData)
	if err != nil {
		return nil, functions.NewE(err)
	}
	return c, nil
}

func CreateClusterReference(displayName, clusterName string, options ...fn.Option) (*Cluster, error) {

	suggested, err := getClusterName(clusterName, options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	cn := clusterName
	if !suggested.Result {
		if len(suggested.SuggestedNames) == 0 {
			return nil, fmt.Errorf("no suggested names for cluster %s", clusterName)
		}
		cn = suggested.SuggestedNames[0]
	}

	respData, err := klFetch("cli_createClusterReference", map[string]any{
		"cluster": map[string]any{
			"displayName": displayName,
			"metadata": map[string]any{
				"name": cn,
			},
			"visibility": map[string]any{
				"mode": "private",
			},
		},
	}, &cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %s", err.Error())
	}

	c, err := GetFromResp[Cluster](respData)
	if err != nil {
		return nil, functions.NewE(err)
	}

	return c, nil
}
