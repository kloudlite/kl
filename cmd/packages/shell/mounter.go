package shell

import (
	"os"
	"path"
	"strings"

	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func mount(mounts map[string]string, mountpath string) error {
	defer spinner.Client.UpdateMessage("mounting...")()

	for k, v := range mounts {
		key := strings.TrimPrefix(k, "$kl_mounts")

		prefixedPath := path.Join(mountpath, key)

		// fn.Logf(text.Yellow("mounting %s at %s\n"), key, prefixedPath)
		os.MkdirAll(path.Dir(prefixedPath), 0o700)

		if err := os.WriteFile(prefixedPath, []byte(v), 0o700); err != nil {
			return err
		}
	}

	return nil
}
