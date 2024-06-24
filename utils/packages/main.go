package packages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/kloudlite/kl2/utils/klfile"
)

func SyncLockfileWithNewConfig(config klfile.KLFileType) (map[string]string, error) {
	_, err := os.Stat("kl.lock")
	packages := map[string]string{}
	if err == nil {
		file, err := os.ReadFile("kl.lock")
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(file, &packages); err != nil {
			return nil, err
		}
	}

	for p := range config.Packages {
		splits := strings.Split(config.Packages[p], "@")
		if len(splits) == 1 {
			splits = append(splits, "latest")
		}

		if _, ok := packages[splits[0]+"@"+splits[1]]; ok {
			continue
		}

		resp, err := http.Get(fmt.Sprintf("https://search.devbox.sh/v1/resolve?name=%s&version=%s", splits[0], splits[1]))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch package %s", config.Packages[p])
		}

		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		type Res struct {
			CommitHash string `json:"commit_hash"`
			Version    string `json:"version"`
		}

		var res Res
		err = json.Unmarshal(all, &res)
		if err != nil {
			return nil, err
		}

		packages[splits[0]+"@"+res.Version] = fmt.Sprintf("nixpkgs/%s#%s", res.CommitHash, splits[0])
	}

	for k := range packages {
		splits := strings.Split(k, "@")
		if !slices.Contains(config.Packages, splits[0]) && !slices.Contains(config.Packages, k) && !slices.Contains(config.Packages, splits[0]+"@latest") {
			delete(packages, k)
		}
	}

	marshal, err := json.Marshal(packages)
	if err != nil {
		return nil, err
	}

	bjson := new(bytes.Buffer)
	if err = json.Indent(bjson, marshal, "", "  "); err != nil {
		return nil, err
	}

	if err = os.WriteFile("kl.lock", bjson.Bytes(), 0644); err != nil {
		return nil, err
	}

	return packages, nil
}
