package lib

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type queryConfig struct {
	ConfigName string `json:"configName"`
	Key        string `json:"key"`
}

type querySecret struct {
	SecretName string `json:"secretName"`
	Key        string `json:"key"`
}

func ParseEnvVarsAndMounts(cfg *KLConfig) (envVars []string, mounts map[string][]byte, err error) {
	// TODO: validate whether Team with name: cfg.TeamName exists
	cookie, err := fileclient.GetCookieString(fn.MakeOption("teamName", cfg.TeamName))
	if err != nil {
		return nil, nil, err
	}

	configQueries := make([]queryConfig, 0, len(cfg.EnvVars))
	secretQueries := make([]querySecret, 0, len(cfg.EnvVars))
	managedResourceQueries := make([]querySecret, 0, len(cfg.EnvVars))

	for _, v := range cfg.EnvVars {
		switch {
		case v.ConfigRef != nil:
			{
				sp := strings.SplitN(*v.ConfigRef, "/", 2)
				if len(sp) != 2 {
					return nil, nil, fmt.Errorf("invalid value for env-var (key: %s) (value: %s), value must be in format <config-name>/<data-key>", v.Key, *v.ConfigRef)
				}
				configQueries = append(configQueries, queryConfig{
					ConfigName: sp[0],
					Key:        sp[1],
				})
			}
		case v.SecretRef != nil:
			{
				sp := strings.SplitN(*v.SecretRef, "/", 2)
				if len(sp) != 2 {
					return nil, nil, fmt.Errorf("invalid value for env-var (key: %s) (value: %s), value must be in format <secret-name>/<data-key>", v.Key, *v.ConfigRef)
				}
				secretQueries = append(secretQueries, querySecret{
					SecretName: sp[0],
					Key:        sp[1],
				})
			}
		case v.MresRef != nil:
			{
				sp := strings.SplitN(*v.MresRef, "/", 2)
				if len(sp) != 2 {
					return nil, nil, fmt.Errorf("invalid value for env-var (key: %s) (value: %s), value must be in format <secret-name>/<data-key>", v.Key, *v.ConfigRef)
				}
				managedResourceQueries = append(managedResourceQueries, querySecret{
					SecretName: sp[0],
					Key:        sp[1],
				})
			}
		}
	}

	for _, v := range cfg.Mounts {
		switch {
		case v.ConfigRef != nil:
			{
				sp := strings.SplitN(*v.ConfigRef, "/", 2)
				if len(sp) != 2 {
					return nil, nil, fmt.Errorf("invalid value for mount (path: %s) (value: %s), value must be in format <config-name>/<data-key>", v.Path, *v.ConfigRef)
				}
				configQueries = append(configQueries, queryConfig{
					ConfigName: sp[0],
					Key:        sp[1],
				})
			}
		case v.SecretRef != nil:
			{
				sp := strings.SplitN(*v.SecretRef, "/", 2)
				if len(sp) != 2 {
					return nil, nil, fmt.Errorf("invalid value for mount (path: %s) (value: %s), value must be in format <secret-name>/<data-key>", v.Path, *v.SecretRef)
				}
				secretQueries = append(secretQueries, querySecret{
					SecretName: sp[0],
					Key:        sp[1],
				})
			}
		}
	}

	m := map[string]any{
		"envName":       cfg.DefaultEnv,
		"configQueries": configQueries,
		"mresQueries":   managedResourceQueries,
		"secretQueries": secretQueries,
	}

	respData, err := klFetch("cli_getConfigSecretMap", m, &cookie)
	if err != nil {
		return nil, nil, fn.NewE(err)
	}

	var resp apiclient.EnvRsp
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, nil, err
	}

	rm := make(map[string]string)

	for _, v := range resp.Configs {
		rm[fmt.Sprintf("config/%s/%s", v.ConfigName, v.Key)] = v.Value
	}

	for _, v := range resp.Secrets {
		rm[fmt.Sprintf("secret/%s/%s", v.SecretName, v.Key)] = v.Value
	}

	for _, v := range resp.Mreses {
		rm[fmt.Sprintf("mres/%s/%s", v.SecretName, v.Key)] = v.Value
	}

	result := make([]string, 0, len(cfg.EnvVars))

	for _, v := range cfg.EnvVars {
		switch {
		case v.ConfigRef != nil:
			result = append(result, fmt.Sprintf("%s=%s", v.Key, rm["config/"+*v.ConfigRef]))
		case v.SecretRef != nil:
			result = append(result, fmt.Sprintf("%s=%s", v.Key, rm["secret/"+*v.SecretRef]))
		case v.MresRef != nil:
			result = append(result, fmt.Sprintf("%s=%s", v.Key, rm["mres/"+*v.MresRef]))
		case v.Value != nil:
			result = append(result, fmt.Sprintf("%s=%s", v.Key, *v.Value))
		}
	}

	mounts = make(map[string][]byte, len(cfg.Mounts))

	for _, v := range cfg.Mounts {
		switch {
		case v.ConfigRef != nil:
			mounts[v.Path] = []byte(rm["config/"+*v.ConfigRef])
		case v.SecretRef != nil:
			mounts[v.Path] = []byte(rm["secret/"+*v.SecretRef])
		}
	}

	slog.Debug("API (cli_getConfigSecretMap), got", "resp", respData)
	return result, mounts, nil
}
