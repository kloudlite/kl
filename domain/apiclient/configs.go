package apiclient

import (
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type ConfigORSecret struct {
	Entries map[string]string `json:"entries"`
	Name    string            `json:"name"`
}

type Config struct {
	DisplayName string            `yaml:"displayName"`
	Metadata    Metadata          `yaml:"metadata"`
	Status      Status            `yaml:"status"`
	Data        map[string]string `yaml:"data"`
}

func (apic *apiClient) ListConfigs(teamName string, envName string) ([]Config, error) {

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_listConfigs", map[string]any{
		"pq": map[string]any{
			"orderBy":       "updateTime",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"envName": strings.TrimSpace(envName),
	}, &cookie)

	if err != nil {
		return nil, fn.NewE(err)
	}

	if fromResp, err := GetFromRespForEdge[Config](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}

func (apic *apiClient) GetConfig(teamName string, envName string, configName string) (*Config, error) {

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_getConfig", map[string]any{
		"name":    configName,
		"envName": envName,
	}, &cookie)

	if err != nil {
		return nil, fn.NewE(err)
	}

	if fromResp, err := getFromResp[Config](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}
