package fileclient

import (
	"encoding/json"
	"fmt"
	fn "github.com/kloudlite/kl/pkg/functions"
	"os"
	"path"
)

type AccountClusterConfig struct {
	ClusterToken          string `json:"clusterToken"`
	ClusterName           string `json:"cluster"`
	MessageOfficeGRPCAddr string `json:"MessageOfficeGRPCAddr"`
	KloudliteDNSSuffix    string `json:"kloudliteDNSSuffix"`
}

func (a *AccountClusterConfig) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

func (a *AccountClusterConfig) Unmarshal(b []byte) error {
	return json.Unmarshal(b, a)
}

func (c *fclient) GetClusterConfig(account string) (*AccountClusterConfig, error) {

	if account == "" {
		return nil, fn.Error("account is required")
	}

	cfgFolder := c.configPath

	if err := os.MkdirAll(path.Join(cfgFolder, "k3s-local"), 0755); err != nil {
		return nil, fn.NewE(err)
	}

	cfgPath := path.Join(cfgFolder, "k3s-local", fmt.Sprintf("%s.json", account))
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, err
	}

	var accClusterConfig AccountClusterConfig
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fn.NewE(err, "failed to read k3s-local config")
	}

	if err := accClusterConfig.Unmarshal(b); err != nil {
		return nil, fn.NewE(err, "failed to parse k3s-local config")
	}

	return &accClusterConfig, nil
}

func (c *fclient) SetClusterConfig(account string, accClusterConfig *AccountClusterConfig) error {
	if account == "" {
		return fn.Error("account is required")
	}

	cfgFolder := c.configPath

	if err := os.MkdirAll(path.Join(cfgFolder, "k3s-local"), 0755); err != nil {
		return fn.NewE(err)
	}

	cfgPath := path.Join(cfgFolder, "k3s-local", fmt.Sprintf("%s.json", account))

	marshal, err := accClusterConfig.Marshal()
	if err != nil {
		return fn.NewE(err)
	}
	err = os.WriteFile(cfgPath, marshal, 0644)
	if err != nil {
		return fn.NewE(err)
	}

	return nil
}
