package server

import (
	"fmt"

	fn "github.com/kloudlite/kl2/pkg/functions"
)

type Device struct {
	AccountName       string `json:"accountName"`
	CreationTime      string `json:"creationTime"`
	CreatedBy         User   `json:"createdBy"`
	DisplayName       string `json:"displayName"`
	GlobalVPNName     string `json:"globalVPNName"`
	ID                string `json:"id"`
	IPAddress         string `json:"ipAddr"`
	LastUpdatedBy     User   `json:"lastUpdatedBy"`
	MarkedForDeletion bool   `json:"markedForDeletion"`
	EnvironmentName   string `json:"environmentName"`
	Metadata          struct {
		Annotations       map[string]string `json:"annotations"`
		CreationTimestamp string            `json:"creationTimestamp"`
		DeletionTimestamp string            `json:"deletionTimestamp"`
		Labels            map[string]string `json:"labels"`
		Name              string            `json:"name"`
	} `json:"metadata"`
	PrivateKey      string `json:"privateKey"`
	PublicEndpoint  string `json:"publicEndpoint"`
	PublicKey       string `json:"publicKey"`
	UpdateTime      string `json:"updateTime"`
	WireguardConfig struct {
		Value    string `json:"value"`
		Encoding string `json:"encoding"`
	} `json:"wireguardConfig,omitempty"`
}

const (
	Default_GVPN = "default"
)

type DeviceList struct {
	Edges Edges[Env] `json:"edges"`
}

func createDevice(accName, devName string) (*Device, error) {
	cn, err := getDeviceName(accName, devName)
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	dn := devName
	if !cn.Result {
		if len(cn.SuggestedNames) == 0 {
			return nil, fmt.Errorf("no suggested names for device %s", devName)
		}

		dn = cn.SuggestedNames[0]
	}

	respData, err := klFetch("cli_createGlobalVPNDevice", map[string]any{
		"gvpnDevice": map[string]any{
			"metadata":       map[string]string{"name": dn},
			"globalVPNName":  Default_GVPN,
			"displayName":    dn,
			"creationMethod": "kl",
		},
	}, &cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to create vpn: %s", err.Error())
	}

	d, err := GetFromResp[Device](respData)
	if err != nil {
		return nil, err
	}

	if err := client.SelectDevice(d.Metadata.Name); err != nil {
		return nil, err
	}

	return d, nil
}

func DeviceForAccount(accountName string) (*Device, error) {
	d, err := getVPNDevice(devName, accountName)
	if err != nil {
		return nil, err
	}

	if d != nil {
		return d, nil
	}

	return createDevice(accountName, devName)
}

type CheckName struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

const (
	VPNDeviceType = "global_vpn_device"
)

func getDeviceName(accName, devName string) (*CheckName, error) {

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_infraCheckNameAvailability", map[string]any{
		"resType": VPNDeviceType,
		"name":    devName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[CheckName](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func getVPNDevice(devName, accountName string, options ...fn.Option) (*Device, error) {

	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getGlobalVpnDevice", map[string]any{
		"gvpn":       Default_GVPN,
		"deviceName": devName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	return GetFromResp[Device](respData)
}

func CheckDeviceStatus() bool {
	verbose := false

	logF := func(format string, v ...interface{}) {
		if verbose {
			if len(v) > 0 {
				fn.Log(format, v)
			} else {
				fn.Log(format)
			}
		}
	}

	s, err := client.GetDeviceContext()
	if err != nil {
		logF(err.Error())
		return false
	}

	return true
}
