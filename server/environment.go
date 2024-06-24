package server

import (
	fn "github.com/kloudlite/kl2/pkg/functions"
)

type Env struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
	ClusterName string   `json:"clusterName"`
	Spec        struct {
		TargetNamespace string `json:"targetNamespace"`
	} `json:"spec"`
}

type EnvList struct {
	Edges Edges[Env] `json:"edges"`
}

func ListEnvs(options ...fn.Option) ([]Env, error) {
	var err error
	cookie, err := getCookieString(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listEnvironments", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[Env](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func GetEnvironment(envName string) (*Env, error) {
	cookie, err := getCookieString()
	if err != nil {
		return nil, err
	}

	// todo: add env name to the query
	respData, err := klFetch("cli_getEnvironment", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[Env](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}
