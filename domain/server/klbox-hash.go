package server

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"os"
	"path"
)

func generateBoxHashContent() ([]byte, error) {
	config, err := LoadDevboxConfig()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = gob.NewEncoder(&b).Encode(config)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"hash":   fmt.Sprintf("%x", md5.Sum(b.Bytes())),
		"config": config,
	})
}

func SyncBoxHash() error {
	configFolder, err := client.GetConfigFolder()
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	hash := md5.New()
	hash.Write([]byte(cwd))
	boxHashFilePath := fmt.Sprintf("hash-%x", hash.Sum(nil))
	if err != nil {
		return err
	}
	content, err := generateBoxHashContent()
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(configFolder, "box-hash", boxHashFilePath), content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func EnsureBoxHash() {
	if err := ensureBoxHashFolder(); err != nil {
		fn.PrintError(err)
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		fn.PrintError(err)
		return
	}
	// check if kl.yml exists in cwd
	klFile := path.Join(cwd, "kl.yml")
	if _, err := os.Stat(klFile); err != nil {
		return
	}
	SyncBoxHash()
}

func ensureBoxHashFolder() error {
	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path.Join(configFolder, "box-hash"), 0755); err != nil {
		return err
	}
	return nil
}
