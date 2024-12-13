package apiclient

import (
	"github.com/kloudlite/kl/pkg/functions"
)

type Team struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
}

func (apic *apiClient) ListTeams() ([]Team, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listAccounts", map[string]any{}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	type AccList []Team
	if fromResp, err := getFromResp[AccList](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return *fromResp, nil
	}
}
