package server

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

var PaginationDefault = map[string]any{
	"orderBy":       "name",
	"sortDirection": "ASC",
	"first":         99999999,
}

type AppPort struct {
	AppPort    int `json:"appPort"`
	DevicePort int `json:"devicePort,omitempty"`
}

type AppSpec struct {
	Services []struct {
		Port int `json:"port"`
	} `json:"services"`
	Intercept struct {
		PortMappings []AppPort `json:"portMappings"`
	} `json:"intercept"`
}

type App struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Spec        AppSpec  `json:"spec"`
	Status      Status   `json:"status"`
	IsMainApp   bool     `json:"mapp"`
}

func ListApps(options ...fn.Option) ([]App, error) {

	envName := fn.GetOption(options, "envName")
	if envName == "" {
		return nil, errors.New("no environment found")
	}

	cookie, err := getCookieString(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listApps", map[string]any{
		"pq":      PaginationDefault,
		"envName": envName,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[App](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func InterceptApp(status bool, ports []AppPort, deviceName string, app *App, options ...fn.Option) error {

	var err error

	envName := fn.GetOption(options, "envName")

	if envName == "" {
		return functions.Error(err, "no environment found")
	}
	cookie, err := getCookieString(options...)
	if err != nil {
		return functions.Error(err)
	}

	if len(ports) == 0 {
		if len(app.Spec.Intercept.PortMappings) != 0 {
			ports = append(ports, app.Spec.Intercept.PortMappings...)
		} else if len(app.Spec.Services) != 0 {
			for _, v := range app.Spec.Services {
				ports = append(ports, AppPort{
					AppPort:    v.Port,
					DevicePort: v.Port,
				})
			}
		}
	}

	if len(ports) == 0 {
		return fmt.Errorf("no ports provided to intercept")
	}

	query := "cli_interceptApp"
	if !app.IsMainApp {
		query = "cli_intercepExternalApp"
	}

	respData, err := klFetch(query, map[string]any{
		"appName":      app.Metadata.Name,
		"envName":      envName,
		"deviceName":   deviceName,
		"intercept":    status,
		"portMappings": ports,
	}, &cookie)

	if err != nil {
		return functions.Error(err)
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return functions.Error(err)
	} else {
		return nil
	}
}
