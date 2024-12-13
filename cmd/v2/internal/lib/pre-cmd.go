package lib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func PreCommand() (*KLConfig, *ParsedKLConfig, error) {
	klc, err := FindKLFile()
	if err != nil {
		return nil, nil, err
	}

	cacheDir, err := CreateKLCacheDir(klc.ConfigFile)
	if err != nil {
		return nil, nil, err
	}

	pk, err := PullVariables(klc, cacheDir)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range pk.Mounts {
		rk := filepath.Join(cacheDir, "mounts", k)
		_, err := os.Stat(rk)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				panic(err)
			}

			if err := os.MkdirAll(filepath.Dir(rk), 0o760); err != nil {
				if !os.IsExist(err) {
					panic(fmt.Errorf("failed to create mount dir (%s)", filepath.Dir(rk)))
				}
			}

			if err := os.WriteFile(rk, v, 0o444); err != nil {
				panic(err)
			}
			continue
		}
		// panic(fmt.Errorf("filepath: %s, already exists", fi.Name()))
	}

	return klc, pk, nil
}
