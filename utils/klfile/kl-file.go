package klfile

import (
	"encoding/json"
	"fmt"
	"os"

	confighandler "github.com/kloudlite/kl2/pkg/config-handler"
	"github.com/kloudlite/kl2/types"
	"github.com/kloudlite/kl2/utils/envvars"
)

var ErrorKlFileNotExists = fmt.Errorf("kl file does not exist")

type KLFileType struct {
	Version    string   `json:"version" yaml:"version"`
	DefaultEnv string   `json:"defaultEnv" yaml:"defaultEnv"`
	Packages   []string `json:"packages" yaml:"packages"`
	Ports      []int    `json:"ports" yaml:"ports"`

	EnvVars envvars.EnvVars `json:"envVars" yaml:"envVars"`
	Mounts  types.Mounts    `json:"mounts" yaml:"mounts"`

	InitScripts []string `json:"initScripts" yaml:"initScripts"`
	AccountName string   `json:"accountName" yaml:"accountName"`
}

func (k *KLFileType) ToJson() ([]byte, error) {
	if k == nil {
		return nil, fmt.Errorf("kl file is nil")
	}

	return json.Marshal(*k)
}

func (k *KLFileType) ParseJson(b []byte) error {
	if k == nil {
		return fmt.Errorf("kl file is nil")
	}

	return json.Unmarshal(b, k)
}

func WriteKLFile(fileObj KLFileType) error {
	if err := confighandler.WriteConfig(GetKLConfigPath(), fileObj, 0644); err != nil {
		return err
	}
	return nil
}

const (
	defaultKLFile = "kl.yml"
)

func GetKLConfigPath() string {
	klfilepath := os.Getenv("KLCONFIG_PATH")
	if klfilepath != "" {
		return klfilepath
	}
	return defaultKLFile
}

func GetKlFile(filePath string) (*KLFileType, error) {
	if filePath == "" {
		s := GetKLConfigPath()
		filePath = s
	}

	klfile, err := confighandler.ReadConfig[KLFileType](filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrorKlFileNotExists
		}
		return nil, err
	}

	return klfile, nil
}
