package shell

import (
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

var ErrHashMismatch = fn.Errorf("hash mismatch")

func getCache() (*fileclient.CacheKLConfig, error) {
	fc := clients.File

	wc, err := fc.GetWsContext()
	if err != nil {
		return nil, err
	}

	ck := wc.GetCache()
	if ck == nil {
		return nil, fn.Errorf("cache is nil")
	}

	b, err := fc.GetKlFileHash()
	if err != nil {
		return nil, err
	}

	if ck.Hash != string(b) {
		return nil, ErrHashMismatch
	}

	return ck, nil
}

func setCache(ck *fileclient.CacheKLConfig) error {
	fc := clients.File

	wc, err := fc.GetWsContext()
	if err != nil {
		return err
	}

	b, err := fc.GetKlFileHash()
	if err != nil {
		return err
	}

	ck.Hash = string(b)
	wc.SetCache(ck)
	return nil
}
