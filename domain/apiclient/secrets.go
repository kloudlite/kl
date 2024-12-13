package apiclient

import (
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type Secret struct {
	DisplayName string            `yaml:"displayName" json:"displayName"`
	Metadata    Metadata          `yaml:"metadata" json:"metadata"`
	Status      Status            `yaml:"status" json:"status"`
	StringData  map[string]string `yaml:"stringData" json:"stringData"`
	IsReadyOnly bool              `yaml:"isReadyOnly" json:"isReadyOnly"`
}

func (apic *apiClient) ListSecrets(teamName string, envName string) ([]Secret, error) {

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_listSecrets", map[string]any{
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
	if fromResp, err := GetFromRespForEdge[Secret](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		var secrets []Secret
		for _, s := range fromResp {
			if !s.IsReadyOnly {
				secrets = append(secrets, s)
			}
		}
		return secrets, nil
	}
}

func (apic *apiClient) GetSecret(teamName string, secretName string) (*Secret, error) {

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	currentEnv, err := apic.EnsureEnv()
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_getSecret", map[string]any{
		"name":    secretName,
		"envName": strings.TrimSpace(currentEnv),
	}, &cookie)

	if err != nil {
		return nil, fn.NewE(err)
	}

	if fromResp, err := getFromResp[Secret](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}
