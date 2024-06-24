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
	// TODO: match with api (envname)
	EnvironmentName string `json:"environmentName"`
	Metadata        struct {
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

func CreateDevice(devName string, accountName string) (*Device, error) {

	cookie, err := getCookieString(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, err
	}

	fn.Logf("creating new device %s", devName)
	respData, err := klFetch("cli_createGlobalVPNDevice", map[string]any{
		"gvpnDevice": map[string]any{
			"metadata":       map[string]string{"name": devName},
			"globalVPNName":  Default_GVPN,
			"displayName":    devName,
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

	return d, nil
}

type CheckName struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

const (
	VPNDeviceType = "global_vpn_device"
)

func GetDeviceName(devName string, accountName string) (*CheckName, error) {
	cookie, err := getCookieString(fn.MakeOption("accountName", accountName))
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
