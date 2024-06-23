package server

import (
	"errors"
	fn "github.com/kloudlite/kl2/pkg/functions"
)

type Mres struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
}

func ListMreses(options ...fn.Option) ([]Mres, error) {

	env := fn.GetOption(options, "envName")

	if env == "" {
		return nil, errors.New("no environment found")
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listMreses", map[string]any{
		"envName": env,
		"search": map[string]any{
			"envName": map[string]any{
				"matchType": "exact",
				"exact":     env,
			},
		},
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)
	if err != nil {
		return nil, err
	}

	fromResp, err := GetFromRespForEdge[Mres](respData)
	if err != nil {
		return nil, err
	}

	return fromResp, nil
}

func ListMresKeys(options ...fn.Option) ([]string, error) {
	mresName := fn.GetOption(options, "mresName")
	env := fn.GetOption(options, "envName")

	if env == "" {
		return nil, errors.New("no environment found")
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getMresKeys", map[string]any{
		"envName": env,
		"name":    mresName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	s, err := GetFromResp[[]string](respData)
	if err != nil {
		return nil, err
	}

	return *s, nil
}
