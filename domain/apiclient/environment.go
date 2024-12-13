package apiclient

import (
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Env struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
	ClusterName string   `json:"clusterName"`
	IsArchived  bool     `json:"isArchived"`
	Spec        struct {
		Suspend         bool   `json:"suspend"`
		TargetNamespace string `json:"targetNamespace"`
	} `json:"spec"`
}

type EnvList struct {
	Edges Edges[Env] `json:"edges"`
}

const (
	PublicEnvRoutingMode = "public"
	EnvironmentType      = "environment"
)

var NoDefaultEnvError = fn.Error("please initialize kl.yml by running `kl init` in current workspace")

func (apic *apiClient) ListEnvs(teamName string) ([]Env, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listEnvironments", map[string]any{
		"pq": map[string]any{
			"orderBy":       "updateTime",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	if fromResp, err := GetFromRespForEdge[Env](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, nil
	}
}

func (apic *apiClient) GetEnvironment(teamName, envName string) (*Env, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_getEnvironment", map[string]any{
		"name": envName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := getFromResp[Env](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func (apic *apiClient) EnsureEnv() (string, error) {
	CurrentEnv, err := apic.fc.CurrentEnv()
	if err != nil {
		return "", functions.NewE(err)
	} else if err == nil {
		return CurrentEnv, nil
	}
	kt, err := apic.fc.GetKlFile()
	if err != nil {
		return "", functions.NewE(err)
	}
	if kt.DefaultEnv == "" {
		return "", NoDefaultEnvError
	}
	e, err := apic.GetEnvironment(kt.TeamName, kt.DefaultEnv)
	if err != nil {
		return "", functions.NewE(err)
	}

	return e.Metadata.Name, nil
}

func (apic *apiClient) CloneEnv(teamName, envName, newEnvName, clusterName string) (*Env, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, functions.NewE(err)
	}
	respData, err := klFetch("cli_cloneEnvironment", map[string]any{
		"clusterName":            clusterName,
		"sourceEnvName":          envName,
		"destinationEnvName":     newEnvName,
		"displayName":            newEnvName,
		"environmentRoutingMode": PublicEnvRoutingMode,
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	if fromResp, err := getFromResp[Env](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, functions.NewE(err)
	}
}

func (apic *apiClient) CheckEnvName(teamName, envName string) (bool, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return false, functions.NewE(err)
	}
	respData, err := klFetch("cli_coreCheckNameAvailability", map[string]any{
		"resType": EnvironmentType,
		"name":    envName,
	}, &cookie)
	if err != nil {
		return false, functions.NewE(err)
	}

	if fromResp, err := getFromResp[CheckName](respData); err != nil {
		return false, functions.NewE(err)
	} else {
		return fromResp.Result, nil
	}
}

func (apic *apiClient) UpdateEnvironment(teamName string, env *Env, isSuspend bool) error {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return functions.NewE(err)
	}
	_, err = klFetch("cli_updateEnvironment", map[string]any{
		"env": map[string]any{
			"displayName": env.DisplayName,
			"clusterName": env.ClusterName,
			"metadata": map[string]any{
				"name":      env.Metadata.Name,
				"namespace": env.Metadata.Namespace,
			},
			"spec": map[string]any{
				"suspend":         isSuspend,
				"targetNamespace": env.Spec.TargetNamespace,
			},
		},
	}, &cookie)
	if err != nil {
		return functions.NewE(err)
	}
	return nil
}
