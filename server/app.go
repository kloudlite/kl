package server

import (
	"errors"

	fn "github.com/kloudlite/kl2/pkg/functions"
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

func InterceptApp(status bool, ports []int, app *App, options ...fn.Option) error {

	// var err error

	// envName := fn.GetOption(options, "envName")

	// if envName == "" {
	// 	return errors.New("no environment found")
	// }

	// cookie, err := getCookie(options...)
	// if err != nil {
	// 	return err
	// }

	// if len(ports) == 0 {
	// 	if len(app.Spec.Intercept.PortMappings) != 0 {
	// 		ports = append(ports, app.Spec.Intercept.PortMappings...)
	// 	} else if len(app.Spec.Services) != 0 {
	// 		for _, v := range app.Spec.Services {
	// 			ports = append(ports, AppPort{
	// 				AppPort:    v.Port,
	// 				DevicePort: v.Port,
	// 			})
	// 		}
	// 	}
	// }

	// if err := func() error {
	// 	sshPort, ok := os.LookupEnv("SSH_PORT")
	// 	if ok {
	// 		var prs []sshclient.StartCh

	// 		for _, v := range ports {
	// 			prs = append(prs, sshclient.StartCh{
	// 				SshPort:    sshPort,
	// 				RemotePort: fmt.Sprint(v.DevicePort),
	// 				LocalPort:  fmt.Sprint(v.DevicePort),
	// 			})
	// 		}

	// 		p, err := proxy.NewProxy(false)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		if status {
	// 			if _, err := p.AddFwd(prs); err != nil {
	// 				fn.PrintError(err)
	// 				return err
	// 			}
	// 			return nil
	// 		}

	// 		if _, err := p.RemoveFwd(prs); err != nil {
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// }(); err != nil {
	// 	fn.PrintError(err)
	// }

	// if len(ports) == 0 {
	// 	return fmt.Errorf("no ports provided to intercept")
	// }

	// query := "cli_interceptApp"
	// if !app.IsMainApp {
	// 	query = "cli_intercepExternalApp"
	// }

	// respData, err := klFetch(query, map[string]any{
	// 	"appName":      app.Metadata.Name,
	// 	"envName":      envName,
	// 	"deviceName":   devName,
	// 	"intercept":    status,
	// 	"portMappings": ports,
	// }, &cookie)

	// if err != nil {
	// 	return err
	// }

	// if _, err := GetFromResp[bool](respData); err != nil {
	// 	return err
	// } else {
	// 	return nil
	// }
	return nil
}
