package server

import (
	"errors"
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

func ListConfigs(options ...fn.Option) ([]Config, error) {

	env := fn.GetOption(options, "envName")
	if env == "" {
		return nil, errors.New("no environment found")
	}

	cookie, err := getCookieString(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listConfigs", map[string]any{
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

	if fromResp, err := GetFromRespForEdge[Config](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func GetConfig(options ...fn.Option) (*Config, error) {
	env := fn.GetOption(options, "envName")
	configName := fn.GetOption(options, "configName")

	if env == "" {
		return nil, errors.New("no environment found")
	}

	cookie, err := getCookieString(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getConfig", map[string]any{
		"name":    configName,
		"envName": strings.TrimSpace(env),
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[Config](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}
