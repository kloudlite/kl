package lib

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

func cacheFile(dir string) string {
	return filepath.Join(dir, "cache.json")
}

func lookupKLFile(dir string) (string, error) {
	files := []string{"kl.yml", "kl.yaml"}
	for _, name := range files {
		fi, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			continue
			// return nil, err
		}
		if fi.IsDir() {
			continue
			// return nil, fmt.Errorf("config file: %s is a directory, must be a file", fi.Name())
		}
		return filepath.Join(dir, name), nil
	}

	return lookupKLFile(filepath.Dir(dir))
}

func FindKLFile() (*KLConfig, error) {
	file, ok := os.LookupEnv("KLCONFIG_PATH")
	if !ok {
		wd, _ := os.Getwd()
		var err error
		file, err = lookupKLFile(wd)
		if err != nil {
			return nil, err
		}
	}

	slog.Debug("got", "file", file)

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	var klConfig KLConfig

	if err := yaml.NewDecoder(f).Decode(&klConfig); err != nil {
		return nil, err
	}

	klConfig.ConfigFile = file

	klConfig.packagesMap = make(map[string]int)
	for i, item := range klConfig.Packages {
		sp := strings.Split(item, "|")
		if len(sp) != 2 {
			continue
		}
		name := sp[0]
		s := strings.SplitN(name, "@", 2)
		klConfig.packagesMap[s[0]] = i
	}

	klConfig.librariesMap = make(map[string]int)
	for i, item := range klConfig.Libraries {
		sp := strings.Split(item, "|")
		if len(sp) != 2 {
			continue
		}
		name := sp[0]
		s := strings.SplitN(name, "@", 2)
		klConfig.librariesMap[s[0]] = i
	}

	return &klConfig, nil
}

func CreateKLCacheDir(klfile string) (cacheDir string, err error) {
	cacheDir = filepath.Join(filepath.Dir(klfile), ".kl")

	if err := os.Mkdir(cacheDir, 0o700); err != nil {
		if !os.IsExist(err) {
			panic("failed to create .kl directory")
		}
	}

	return cacheDir, nil
}

func pullVariablesFromCache(cacheFile string) (*ParsedKLConfig, error) {
	b, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var p ParsedKLConfig
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func PullVariablesFromAPI(klc *KLConfig, cacheDir string) (*ParsedKLConfig, error) {
	// TODO: add this cachedir to `.gitignore`
	envVars, mounts, err := ParseEnvVarsAndMounts(klc)
	if err != nil {
		panic(err)
	}

	hash := GenHash(GenHashArgs{
		EnvVars:   envVars,
		Mounts:    mounts,
		Packages:  []string{},
		Libraries: []string{},
	})

	pkl := &ParsedKLConfig{
		ConfigFile: klc.ConfigFile,
		CacheFile:  cacheFile(cacheDir),
		CacheDir:   cacheDir,

		EnvVars: envVars,
		Mounts:  mounts,
		Hash:    hash,
	}

	b, err := json.Marshal(pkl)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(cacheFile(cacheDir), b, 0o500); err != nil {
		return nil, err
	}

	return pkl, nil
}

func PullVariables(klc *KLConfig, cacheDir string) (*ParsedKLConfig, error) {
	cf := cacheFile(cacheDir)
	_, err := os.Stat(cf)
	if err != nil {
		if os.IsNotExist(err) {
			return PullVariablesFromAPI(klc, cacheDir)
		}
	}
	return pullVariablesFromCache(cf)
}
