package packagectrl

import (
	"encoding/json"
	"fmt"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/fjson"
	fn "github.com/kloudlite/kl/pkg/functions"
)

//type Packages map[string]string

type Pkgs struct {
	Packages  map[string]string `json:"packages" yaml:"packages"`
	Libraries map[string]string `json:"libraries" yaml:"libraries"`
}

func (p *Pkgs) Marshal() ([]byte, error) {
	return fjson.Marshal(p)
}

func (p *Pkgs) Unmarshal(b []byte) error {
	return fjson.Unmarshal(b, p)
}

func SyncLockfileWithNewConfig(config fileclient.KLFileType) (*Pkgs, error) {
	defer spinner.Client.UpdateMessage("installing nix packages")()
	_, err := os.Stat("kl.lock")
	pkgs := Pkgs{
		Packages:  make(map[string]string),
		Libraries: make(map[string]string),
	}
	if err == nil {
		file, err := os.ReadFile("kl.lock")
		if err != nil {
			return nil, fn.NewE(err)
		}

		if err := pkgs.Unmarshal(file); err != nil {
			return nil, fn.NewE(err)
		}
	}

	packagesMap := make(map[string]string)
	for k := range pkgs.Packages {
		splits := strings.Split(k, "@")
		if len(splits) != 2 {
			continue
		}
		packagesMap[splits[0]] = splits[1]
	}

	for p := range config.Packages {
		splits := strings.Split(config.Packages[p], "@")
		if len(splits) == 1 {
			if _, ok := packagesMap[splits[0]]; ok {
				continue
			}
			splits = append(splits, "latest")
		}

		if _, ok := pkgs.Packages[splits[0]+"@"+splits[1]]; ok {
			continue
		}

		platform := os.Getenv("PLATFORM_ARCH") + "-linux"
		if platform == "-linux" {
			platform = "x86_64-linux"
		}

		resp, err := http.Get(fmt.Sprintf("https://search.devbox.sh/v1/resolve?name=%s&version=%s&platform=%s", splits[0], splits[1], platform))
		if err != nil {
			return nil, fn.NewE(err)
		}

		if resp.StatusCode != 200 {
			return nil, fn.Errorf("failed to fetch package %s", config.Packages[p])
		}

		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fn.NewE(err)
		}

		type System struct {
			AttrPaths []string `json:"attr_paths"`
		}

		type Res struct {
			CommitHash string            `json:"commit_hash"`
			Version    string            `json:"version"`
			Systems    map[string]System `json:"systems"`
		}

		var res Res
		err = json.Unmarshal(all, &res)
		if err != nil {
			return nil, fn.NewE(err)
		}

		pkgs.Packages[splits[0]+"@"+res.Version] = fmt.Sprintf("nixpkgs/%s#%s", res.CommitHash, res.Systems[platform].AttrPaths[0])
	}

	for k := range pkgs.Packages {
		splits := strings.Split(k, "@")
		if (!slices.Contains(config.Packages, splits[0]) && !slices.Contains(config.Packages, k) && !slices.Contains(config.Packages, splits[0]+"@latest")) && (!slices.Contains(config.Libraries, splits[0]) && !slices.Contains(config.Libraries, k) && !slices.Contains(config.Libraries, splits[0]+"@latest")) {
			delete(pkgs.Packages, k)
		}
	}

	marshal, err := pkgs.Marshal()
	if err != nil {
		return nil, fn.NewE(err)
	}

	if err = os.WriteFile("kl.lock", marshal, 0644); err != nil {
		return nil, fn.NewE(err)
	}

	return &pkgs, nil
}

func SyncLockfileWithNewConfigLibs(config fileclient.KLFileType) (*Pkgs, error) {
	defer spinner.Client.UpdateMessage("installing nix libraries")()
	_, err := os.Stat("kl.lock")
	pkgs := Pkgs{
		Packages:  make(map[string]string),
		Libraries: make(map[string]string),
	}
	if err == nil {
		file, err := os.ReadFile("kl.lock")
		if err != nil {
			return nil, fn.NewE(err)
		}

		if err := pkgs.Unmarshal(file); err != nil {
			return nil, fn.NewE(err)
		}
	}

	librariesMap := make(map[string]string)
	for k := range pkgs.Libraries {
		splits := strings.Split(k, "@")
		if len(splits) != 2 {
			continue
		}
		librariesMap[splits[0]] = splits[1]
	}

	for p := range config.Libraries {
		splits := strings.Split(config.Libraries[p], "@")
		if len(splits) == 1 {
			if _, ok := librariesMap[splits[0]]; ok {
				continue
			}
			splits = append(splits, "latest")
		}

		if _, ok := pkgs.Libraries[splits[0]+"@"+splits[1]]; ok {
			continue
		}

		platform := os.Getenv("PLATFORM_ARCH") + "-linux"
		if platform == "-linux" {
			platform = "x86_64-linux"
		}

		resp, err := http.Get(fmt.Sprintf("https://search.devbox.sh/v1/resolve?name=%s&version=%s&platform=%s", splits[0], splits[1], platform))
		if err != nil {
			return nil, fn.NewE(err)
		}

		if resp.StatusCode != 200 {
			return nil, fn.Errorf("failed to fetch library %s", config.Libraries[p])
		}

		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fn.NewE(err)
		}

		type System struct {
			AttrPaths []string `json:"attr_paths"`
		}

		type Res struct {
			CommitHash string            `json:"commit_hash"`
			Version    string            `json:"version"`
			Systems    map[string]System `json:"systems"`
		}

		var res Res
		err = json.Unmarshal(all, &res)
		if err != nil {
			return nil, fn.NewE(err)
		}

		pkgs.Libraries[splits[0]+"@"+res.Version] = fmt.Sprintf("nixpkgs/%s#%s", res.CommitHash, res.Systems[platform].AttrPaths[0])
	}

	for k := range pkgs.Libraries {
		splits := strings.Split(k, "@")
		if (!slices.Contains(config.Libraries, splits[0]) && !slices.Contains(config.Libraries, k) && !slices.Contains(config.Libraries, splits[0]+"@latest")) && (!slices.Contains(config.Packages, splits[0]) && !slices.Contains(config.Packages, k) && !slices.Contains(config.Packages, splits[0]+"@latest")) {
			delete(pkgs.Libraries, k)
		}
	}

	marshal, err := pkgs.Marshal()
	if err != nil {
		return nil, fn.NewE(err)
	}

	if err = os.WriteFile("kl.lock", marshal, 0644); err != nil {
		return nil, fn.NewE(err)
	}

	return &pkgs, nil
}
