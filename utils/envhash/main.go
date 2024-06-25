package envhash

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/types"
	"github.com/kloudlite/kl/utils/envvars"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/kloudlite/kl/utils/packages"
)

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func generateBoxHashContent(envName string) ([]byte, error) {
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return nil, functions.Error(err)
	}

	persistedConfig, err := generatePersistedEnv(klFile, envName)
	if err != nil {
		return nil, functions.Error(err)
	}

	hash := md5.New()
	mountKeys := keys(persistedConfig.Mounts)
	slices.Sort(mountKeys)
	for _, v := range mountKeys {
		hash.Write([]byte(v))
		hash.Write([]byte(persistedConfig.Mounts[v]))
	}

	packages := keys(persistedConfig.PackageHashes)
	slices.Sort(packages)
	for _, v := range packages {
		hash.Write([]byte(v))
		hash.Write([]byte(persistedConfig.PackageHashes[v]))
	}

	envKeys := keys(persistedConfig.Env)
	slices.Sort(envKeys)
	for _, v := range envKeys {
		hash.Write([]byte(v))
		hash.Write([]byte(persistedConfig.Env[v]))
	}

	marshal, err := json.Marshal(map[string]any{
		"config": persistedConfig,
		"hash":   hash.Sum(nil),
	})
	if err != nil {
		return nil, err
	}

	return marshal, nil
}

func BoxHashFile(workspacePath string) (*types.PersistedEnv, error) {
	fileName, err := BoxHashFileName(workspacePath)
	if err != nil {
		return nil, err
	}
	configFolder, err := server.GetConfigFolder()
	if err != nil {
		return nil, err
	}
	filePath := path.Join(configFolder, "box-hash", fileName)
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if os.IsNotExist(err) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, functions.Error(err)
		}

		env, err := server.EnvAtPath(cwd)
		if err != nil {
			return nil, functions.Error(err)
		}
		if err = SyncBoxHash(env.Name); err != nil {
			return nil, functions.Error(err)
		}
		return BoxHashFile(cwd)
	}
	var r struct {
		Config types.PersistedEnv `json:"config"`
		Hash   string             `json:"hash"`
	}

	if err = json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r.Config, nil
}

func BoxHashFileName(path string) (string, error) {
	if os.Getenv("IN_DEV_BOX") == "true" {
		path = os.Getenv("KL_WORKSPACE")
	}

	hash := md5.New()
	if _, err := hash.Write([]byte(path)); err != nil {
		return "", nil
	}

	return fmt.Sprintf("hash-%x", hash.Sum(nil)), nil
}

func SyncBoxHash(envName string) error {

	if envName == "" {
		return functions.NewError("envName is required")
	}

	configFolder, err := server.GetConfigFolder()
	if err != nil {
		return functions.Error(err)
	}

	cwd, _ := os.Getwd()
	if os.Getenv("IN_DEV_BOX") == "true" {
		cwd = os.Getenv("KL_WORKSPACE")
	}

	boxHashFilePath, err := BoxHashFileName(cwd)
	if err != nil {
		return functions.Error(err)
	}

	content, err := generateBoxHashContent(envName)
	if err != nil {
		return functions.Error(err)
	}

	if err = os.MkdirAll(path.Join(configFolder, "box-hash"), 0755); err != nil {
		return functions.Error(err)
	}

	if err = os.WriteFile(path.Join(configFolder, "box-hash", boxHashFilePath), content, 0644); err != nil {
		return functions.Error(err)
	}

	return nil
}

func GenerateKLConfigHash(kf *klfile.KLFileType) (string, error) {
	klConfhash := md5.New()
	slices.SortFunc(kf.EnvVars, func(a, b envvars.EnvType) int {
		return strings.Compare(a.Key, b.Key)
	})
	for _, v := range kf.EnvVars {
		klConfhash.Write([]byte(v.Key))
		klConfhash.Write([]byte(func() string {
			if v.Value != nil {
				return *v.Value
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.ConfigRef != nil {
				return *v.ConfigRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.SecretRef != nil {
				return *v.SecretRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.MresRef != nil {
				return *v.MresRef
			}
			return ""
		}()))
	}
	slices.Sort(kf.Packages)
	for _, v := range kf.Packages {
		klConfhash.Write([]byte(v))
	}
	for _, v := range kf.Mounts {
		klConfhash.Write([]byte(v.Path))
		klConfhash.Write([]byte(func() string {
			if v.ConfigRef != nil {
				return *v.ConfigRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(func() string {
			if v.SecretRef != nil {
				return *v.SecretRef
			}
			return ""
		}()))
		klConfhash.Write([]byte(v.Path))
	}
	return fmt.Sprintf("%x", klConfhash.Sum(nil)), nil
}

func generatePersistedEnv(kf *klfile.KLFileType, envName string) (*types.PersistedEnv, error) {
	envs, mm, err := server.GetLoadMaps(envName)
	if err != nil {
		return nil, err
	}

	realPkgs, err := packages.SyncLockfileWithNewConfig(*kf)
	if err != nil {
		return nil, err
	}

	hashConfig := &types.PersistedEnv{
		Packages:      kf.Packages,
		PackageHashes: realPkgs,
	}

	fm := map[string]string{}
	for _, fe := range kf.Mounts.GetMounts() {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		fm[pth] = mm[pth]
	}

	ev := map[string]string{}
	for k, v := range envs {
		ev[k] = v
	}

	for _, ne := range kf.EnvVars.GetEnvs() {
		ev[ne.Key] = ne.Value
	}

	if err == nil {
		ev["PURE_PROMPT_SYMBOL"] = fmt.Sprintf("(%s) %s", envName, "❯")
	}
	klConfhash, err := GenerateKLConfigHash(kf)
	if err != nil {
		return nil, err
	}

	hashConfig.Env = ev
	hashConfig.Mounts = fm
	hashConfig.KLConfHash = klConfhash
	return hashConfig, nil
}
