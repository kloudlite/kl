package apiclient

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

var paginationDefault = map[string]any{
	"orderBy":       "updateTime",
	"sortDirection": "ASC",
	"first":         99999999,
}

type ServiceSpec struct {
	GlobalIP string `json:"globalIP"`
	Hostname string `json:"hostname"`
	Ports    []struct {
		AppProtocol string `json:"appProtocol"`
		Name        string `json:"name"`
		Port        int    `json:"port"`
		NodePort    int    `json:"nodePort"`
		Protocol    string `json:"protocol"`
		TargetPort  struct {
			IntVal int    `json:"intVal"`
			StrVal string `json:"strVal"`
			Type   int    `json:"type"`
		} `json:"targetPort"`
	}
	ServiceIp  string   `json:"serviceIp"`
	ServiceRef Metadata `json:"serviceRef"`
}

type InterceptStatus struct {
	Intercepted  bool          `json:"intercepted"`
	PortMappings []ServicePort `json:"portMappings"`
}

type Service struct {
	//DisplayName string      `json:"displayName"`
	Metadata        Metadata        `json:"metadata"`
	InterceptStatus InterceptStatus `json:"interceptStatus"`
	Spec            ServiceSpec     `json:"spec"`
	Status          Status          `json:"status"`
	//IsMainService bool        `json:"mservice"`
}

type ServicePort struct {
	ServicePort int `json:"servicePort"`
	DevicePort  int `json:"devicePort,omitempty"`
}

func (apic *apiClient) ListServices(teamName string, envName string) ([]Service, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listServices", map[string]any{
		"pq":      paginationDefault,
		"envName": envName,
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}
	if fromResp, err := GetFromRespForEdge[Service](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, nil
	}
}

func (apic *apiClient) InterceptService(service *Service, status bool, ports []ServicePort, envName string, options ...fn.Option) error {
	devName := fn.GetOption(options, "deviceName")
	fc := apic.GetFClient()

	teamName, err := fc.GetDataContext().GetWsTeam()
	if err != nil {
		return functions.NewE(err)
	}

	options = append(options, fn.MakeOption("teamName", teamName))

	if devName == "" {
		avc, err := fc.GetDataContext().GetDevice()
		if err != nil {
			return functions.NewE(err)
		}

		if avc.DeviceName == "" {
			return fmt.Errorf("device name is required")
		}

		devName = avc.DeviceName
	}

	cookie, err := getCookie([]fn.Option{
		fn.MakeOption("teamName", teamName),
	}...)
	if err != nil {
		return functions.NewE(err)
	}

	if len(ports) == 0 {
		if service.InterceptStatus.Intercepted {
			ports = append(ports, service.InterceptStatus.PortMappings...)
		} else if len(service.Spec.Ports) != 0 {
			for _, v := range service.Spec.Ports {
				ports = append(ports, ServicePort{
					ServicePort: v.Port,
					DevicePort:  v.Port,
				})
			}
		}
	}

	if len(ports) == 0 {
		return fn.Errorf("no ports provided to intercept")
	}

	query := "cli_createServiceIntercept"
	if !status {
		query = "cli_deleteServiceIntercept"
	}

	respData, err := klFetch(query, map[string]any{
		"serviceName":  service.Spec.ServiceRef.Name,
		"envName":      envName,
		"interceptTo":  devName,
		"portMappings": ports,
	}, &cookie)
	if err != nil {
		return functions.NewE(err)
	}

	if _, err := getFromResp[bool](respData); err != nil {
		return functions.NewE(err)
	} else {
		return nil
	}
}

func (apic *apiClient) RemoveAllIntercepts(options ...fn.Option) error {
	defer spinner.Client.UpdateMessage("Cleaning up intercepts...")()
	devName := fn.GetOption(options, "deviceName")
	teamName := fn.GetOption(options, "teamName")
	currentEnv, err := apic.EnsureEnv()
	if err != nil {
		return functions.NewE(err)
	}

	fc := fileclient.File
	if err != nil {
		return functions.NewE(err)
	}

	if teamName == "" {
		kt, err := fc.GetKlFile()
		if err != nil {
			return functions.NewE(err)
		}

		if kt.TeamName == "" {
			return fn.Errorf("team name is required")
		}

		teamName = kt.TeamName
		options = append(options, fn.MakeOption("teamName", teamName))
	}

	if devName == "" {
		avc, err := fc.GetDataContext().GetDevice()
		if err != nil {
			return functions.NewE(err)
		}

		if avc.DeviceName == "" {
			return fn.Errorf("device name is required")
		}

		devName = avc.DeviceName
	}

	cookie, err := getCookie([]fn.Option{
		fn.MakeOption("teamName", teamName),
	}...)
	if err != nil {
		return functions.NewE(err)
	}
	query := "cli_removeDeviceIntercepts"

	respData, err := klFetch(query, map[string]any{
		"envName":    currentEnv,
		"deviceName": devName,
		//"deviceName": config.ClusterName,
	}, &cookie)
	if err != nil {
		return functions.NewE(err)
	}

	if _, err := getFromResp[bool](respData); err != nil {
		return functions.NewE(err)
	} else {
		return nil
	}
}
