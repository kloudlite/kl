package envhash

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/types"
	"github.com/kloudlite/kl/utils"
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
		return nil, err
	}

	persistedConfig, err := generatePersistedEnv(klFile, envName)
	if err != nil {
		return nil, err
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
	configFolder, err := utils.GetConfigFolder()
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

	hashConfig.Env = ev
	hashConfig.Mounts = fm
	return hashConfig, nil
}
