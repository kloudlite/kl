package fileclient

import (
	"bytes"
	"os"
	"path"

	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type ShellData struct {
	Envs map[string]string
}

type CacheKLConfig struct {
	Hash      string
	EnvVars   map[string]string
	Mounts    map[string]string
	ShellData *ShellData
}

type wsContextData struct {
	EnvName string `json:"envName"`

	Cache *CacheKLConfig `json:"cache"`
}

func (w wsContext) GetShellData() (*ShellData, error) {
	if w.Cache == nil {
		return nil, fn.Errorf("cache is nil")
	}

	if w.Cache.ShellData == nil {
		return nil, fn.Errorf("shell data is nil")
	}

	return w.Cache.ShellData, nil
}

func (w wsContext) SetShellData(shellData *ShellData) error {
	w.Cache.ShellData = shellData
	return w.handler.Write()
}

func (w wsContext) GetCache() *CacheKLConfig {
	return w.Cache
}

func (w wsContext) SetCache(cache *CacheKLConfig) error {
	w.Cache = cache
	return w.handler.Write()
}

type WsContext interface {
	SetEnv(string) error
	GetEnv() (string, error)
	GetCache() *CacheKLConfig
	SetCache(cache *CacheKLConfig) error
	GetShellData() (*ShellData, error)
	SetShellData(shellData *ShellData) error
}

func (w wsContext) GetEnv() (string, error) {
	if w.EnvName == "" {
		return "", ErrEnvNotSelected
	}

	// s, err := getCtxData()
	// if err != nil {
	// 	return "", err
	// }

	// menv, err := s.GetEnv()
	// if err != nil {
	// 	return "", err
	// }

	// if menv != w.EnvName {
	// 	return "", ErrEnvMismatch
	// }

	return w.EnvName, nil
}

func (w wsContext) SetEnv(env string) error {
	// s, err := getCtxData()
	// if err != nil {
	// 	return err
	// }

	// if err := s.SetEnv(env); err != nil {
	// 	return err
	// }
	//
	w.EnvName = env
	return w.handler.Write()
}

type wsContext struct {
	*wsContextData
	handler confighandler.Config[wsContextData]
}

func getNewWsContext() (WsContext, error) {
	cpath, err := assertConfigPath()
	if err != nil {
		return nil, err
	}

	cdir := path.Dir(cpath)

	cachePath := path.Join(cdir, ".kl")
	if err := os.MkdirAll(cachePath, os.ModePerm); err != nil {
		return nil, err
	}

	// ensure .kl in .gitignore
	b, err := os.ReadFile(path.Join(cdir, ".gitignore"))
	if err != nil {
		b = []byte{}
	}

	if !bytes.Contains(b, []byte(".kl")) {
		b = append(b, []byte("\n.kl\n")...)
		if err := os.WriteFile(path.Join(cdir, ".gitignore"), b, 0o644); err != nil {
			return nil, err
		}
	}

	chandler := confighandler.GetHandler[wsContextData](path.Join(cachePath, "config.yaml"))

	cdata, _ := chandler.Read()

	ch := &wsContext{
		handler:       chandler,
		wsContextData: cdata,
	}

	return ch, nil
}
