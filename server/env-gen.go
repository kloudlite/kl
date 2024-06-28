package server

import (
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/types"
	"github.com/kloudlite/kl/utils/klfile"
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
	Key      string `json:"key"`
	MresName string `json:"mresName"`
	Value    string `json:"value"`
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

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

type Kv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type CSResp map[string]map[string]*Kv
type MountMap map[string]string

func GetLoadMaps(envName string) (map[string]string, MountMap, error) {

	kt, err := klfile.GetKlFile("")
	if err != nil {
		return nil, nil, err
	}

	cookie, err := getCookieString(
		functions.MakeOption("envName", envName),
		functions.MakeOption("accountName", kt.AccountName),
	)
	if err != nil {
		return nil, nil, err
	}

	currMreses := kt.EnvVars.GetMreses()
	currSecs := kt.EnvVars.GetSecrets()
	currConfs := kt.EnvVars.GetConfigs()
	currMounts := kt.Mounts.GetMounts()

	respData, err := klFetch("cli_getConfigSecretMap", map[string]any{
		"envName": envName,
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
				if fe.Type == types.ConfigType {
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
						"mresName": rt.Name,
						"key":      v.RefKey,
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
				if fe.Type == types.SecretType {
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
		return nil, nil, err

	}

	fromResp, err := GetFromResp[EnvRsp](respData)

	if err != nil {
		return nil, nil, err
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
		ent := mmap[v.MresName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		if mmap[v.MresName][v.Key] != nil {
			mmap[v.MresName][v.Key].Value = v.Value
		}
	}

	// ************************[ handling mounts ]****************************
	mountMap := map[string]string{}

	for _, fe := range currMounts {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		if fe.Type == types.ConfigType {
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