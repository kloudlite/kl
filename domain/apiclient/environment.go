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
	Spec        struct {
		TargetNamespace string `json:"targetNamespace"`
	} `json:"spec"`
}

type EnvList struct {
	Edges Edges[Env] `json:"edges"`
}

// func GetEnvironment(envName string) (*Env, error) {
// 	var err error
// 	projectName, err := EnsureProject()
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}
//
// 	respData, err := klFetch("cli_getEnvironment", map[string]any{
// 		"projectName": strings.TrimSpace(projectName),
// 		"pq": map[string]any{
// 			"orderBy":       "name",
// 			"sortDirection": "ASC",
// 			"first":         99999999,
// 		},
// 	}, &cookie)
//
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}
//
// 	if fromResp, err := GetFromResp[Env](respData); err != nil {
// 		return nil, functions.NewE(err)
// 	} else {
// 		return fromResp, nil
// 	}
// }

func (apic *apiClient) ListEnvs(options ...fn.Option) ([]Env, error) {
	var err error
	// _, err = EnsureAccount(options...)
	// if err != nil {
	// 	return nil, functions.NewE(err)
	// }

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listEnvironments", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
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

func (apic *apiClient) GetEnvironment(accountName, envName string) (*Env, error) {
	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_getEnvironment", map[string]any{
		"name": envName,
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

// func _EnsureEnv(env *fileclient.Env, options ...fn.Option) (*fileclient.Env, error) {
// 	fc, err := fileclient.New()
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	accountName := fn.GetOption(options, "accountName")
// 	if _, err := EnsureAccount(
// 		fn.MakeOption("accountName", accountName),
// 	); err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if env != nil && env.Name != "" {
// 		return env, nil
// 	}

// 	env, _ = fc.CurrentEnv()

// 	if env != nil {
// 		return env, nil
// 	}

// 	kl, err := fc.GetKlFile("")
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if kl.DefaultEnv == "" {
// 		return nil, functions.Error("please select an environment using 'kl use env'")
// 	}
// 	selectedEnv, err := SelectEnv(kl.DefaultEnv, options...)
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}
// 	return &fileclient.Env{
// 		Name:     selectedEnv.DisplayName,
// 		TargetNs: selectedEnv.Metadata.Namespace,
// 	}, nil
// }
