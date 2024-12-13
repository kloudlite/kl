package apiclient

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
)

type SecretEnv struct {
	Key        string `json:"key"`
	SecretName string `json:"secretName"`
	Value      string `json:"value"`
}

type ConfigEnv struct {
	Key        string `json:"key"`
	ConfigName string `json:"configName"`
	Value      string `json:"value"`
}

type MresEnv struct {
	Key        string `json:"key"`
	SecretName string `json:"secretName"`
	Value      string `json:"value"`
}

type EnvRsp struct {
	Secrets []SecretEnv `json:"secrets"`
	Configs []ConfigEnv `json:"configs"`
	Mreses  []MresEnv   `json:"mreses"`
}

type GeneratedEnvs struct {
	EnvVars    map[string]string `json:"envVars"`
	MountFiles map[string]string `json:"mountFiles"`
}

type Kv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type (
	CSResp   map[string]map[string]*Kv
	MountMap map[string]string
)

func (apic *apiClient) GetLoadMaps() (map[string]string, MountMap, error) {
	fc := apic.fc

	teamName, err := fc.GetDataContext().GetWsTeam()
	if err != nil {
		return nil, nil, err
	}

	kt, err := fc.GetKlFile()
	if err != nil {
		return nil, nil, functions.NewE(err)
	}

	env, err := apic.EnsureEnv()
	if err != nil {
		return nil, nil, functions.NewE(err)
	}

	cookie, err := getCookie([]functions.Option{
		functions.MakeOption("teamName", teamName),
	}...)

	if err != nil {
		return nil, nil, functions.NewE(err)
	}

	currMreses := kt.EnvVars.GetMreses()
	currSecs := kt.EnvVars.GetSecrets()
	currConfs := kt.EnvVars.GetConfigs()

	currMounts := kt.Mounts.GetMounts()

	respData, err := klFetch("cli_getConfigSecretMap", map[string]any{
		"envName": env,
		"configQueries": func() []any {
			var queries []any
			for _, v := range currConfs {
				for _, vv := range v.Env {
					queries = append(queries, map[string]any{
						"configName": v.Name,
						"key":        vv.RefKey,
					})
				}
			}

			for _, fe := range currMounts {
				if fe.Type == fileclient.ConfigType {
					queries = append(queries, map[string]any{
						"configName": fe.Name,
						"key":        fe.Key,
					})
				}
			}

			return queries
		}(),

		"mresQueries": func() []any {
			var queries []any
			for _, rt := range currMreses {
				for _, v := range rt.Env {
					queries = append(queries, map[string]any{
						"secretName": rt.Name,
						"key":        v.RefKey,
					})
				}
			}

			return queries
		}(),

		"secretQueries": func() []any {
			var queries []any
			for _, v := range currSecs {
				for _, vv := range v.Env {
					queries = append(queries, map[string]any{
						"secretName": v.Name,
						"key":        vv.RefKey,
					})
				}
			}

			for _, fe := range currMounts {
				if fe.Type == fileclient.SecretType {
					queries = append(queries, map[string]any{
						"secretName": fe.Name,
						"key":        fe.Key,
					})
				}
			}
			return queries
		}(),
	}, &cookie)
	if err != nil {
		return nil, nil, functions.NewE(err)
	}

	fromResp, err := getFromResp[EnvRsp](respData)
	if err != nil {
		return nil, nil, functions.NewE(err)
	}

	result := map[string]string{}

	cmap := CSResp{}

	for _, rt := range currConfs {
		cmap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			cmap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	smap := CSResp{}

	for _, rt := range currSecs {
		smap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			smap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	mmap := CSResp{}
	for _, rt := range currMreses {
		mmap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			mmap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	// ************************[ adding to result|env ]***************************
	for _, v := range fromResp.Configs {
		ent := cmap[v.ConfigName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		if cmap[v.ConfigName][v.Key] != nil {
			cmap[v.ConfigName][v.Key].Value = v.Value
		}

	}

	for _, v := range fromResp.Secrets {
		ent := smap[v.SecretName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		if smap[v.SecretName][v.Key] != nil {
			smap[v.SecretName][v.Key].Value = v.Value
		}
	}

	for _, v := range fromResp.Mreses {
		ent := mmap[v.SecretName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		if mmap[v.SecretName][v.Key] != nil {
			mmap[v.SecretName][v.Key].Value = v.Value
		}
	}
	// ************************[ handling mounts ]****************************
	mountMap := map[string]string{}

	for _, fe := range currMounts {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		if fe.Type == fileclient.ConfigType {
			mountMap[pth] = func() string {
				for _, ce := range fromResp.Configs {
					if ce.ConfigName == fe.Name && ce.Key == fe.Key {
						return ce.Value
					}
				}
				return ""
			}()
		} else {
			mountMap[pth] = func() string {
				for _, ce := range fromResp.Secrets {
					if ce.SecretName == fe.Name && ce.Key == fe.Key {
						return ce.Value
					}
				}
				return ""
			}()
		}
	}

	return result, mountMap, nil
}
