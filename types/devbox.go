package types

import (
	"encoding/json"
	"fmt"
)

type PersistedEnv struct {
	Packages      []string          `yaml:"packages" json:"packages"`
	PackageHashes map[string]string `yaml:"packageHashes" json:"packageHashes"`
	Env           map[string]string `yaml:"env" json:"env"`
	Mounts        map[string]string `yaml:"mounts" json:"mounts"`
}

func (k *PersistedEnv) ToJson() ([]byte, error) {

	if k == nil {
		return nil, fmt.Errorf("kl file is nil")
	}

	return json.Marshal(*k)
}

func (k *PersistedEnv) ParseJson(b []byte) error {
	if k == nil {
		return fmt.Errorf("kl file is nil")
	}

	return json.Unmarshal(b, k)
}