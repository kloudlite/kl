package server

import (
	"errors"
	"strings"

	fn "github.com/kloudlite/kl2/pkg/functions"
)

type Secret struct {
	DisplayName string            `yaml:"displayName"`
	Metadata    Metadata          `yaml:"metadata"`
	Status      Status            `yaml:"status"`
	StringData  map[string]string `yaml:"stringData"`
}

func ListSecrets(options ...fn.Option) ([]Secret, error) {

	env := fn.GetOption(options, "envName")
	if env == "" {
		return nil, errors.New("no environment found")
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listSecrets", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"envName": strings.TrimSpace(env),
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[Secret](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func GetSecret(options ...fn.Option) (*Secret, error) {
	env := fn.GetOption(options, "envName")
	secName := fn.GetOption(options, "secretName")

	if env == "" {
		return nil, errors.New("no environment found")
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getSecret", map[string]any{
		"name":    secName,
		"envName": strings.TrimSpace(env),
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[Secret](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}
