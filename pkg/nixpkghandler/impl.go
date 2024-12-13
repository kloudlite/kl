package nixpkghandler

import (
	"fmt"
	"strings"

	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

func (p *pkgHandler) Search(query string) (*SearchResults, error) {
	return search(p.cmd.Context(), query)
}

func (p *pkgHandler) AddLibrary(name, pkghash string) error {
	if err := p.syncPackage(pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addLibToLock(name, pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addLibToKlFile(name); err != nil {
		return fn.NewE(err)
	}

	if err := p.SyncLockfile(); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (p *pkgHandler) RemoveLibrary(name string) error {
	if err := p.rmLibFromKlFile(name); err != nil {
		return fn.NewE(err)
	}

	return p.SyncLockfile()
}

func (p *pkgHandler) AddPackage(name, pkghash string) error {
	if err := p.syncPackage(pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addPackageToLock(name, pkghash); err != nil {
		return fn.NewE(err)
	}

	if err := p.addPackageKlFile(name); err != nil {
		return fn.NewE(err)
	}

	if err := p.SyncLockfile(); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (p *pkgHandler) RemovePackage(name string) error {
	if err := p.rmPackageFromKlFile(name); err != nil {
		return fn.NewE(err)
	}

	if err := p.SyncLockfile(); err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (p *pkgHandler) Find(pname string) (string, string, error) {
	var name string
	var ver string

	if !strings.Contains(pname, "@") {
		sr, err := p.Search(pname)
		if err != nil {
			return "", "", fn.NewE(err)
		}

		pkg, err := fzf.FindOne(sr.Packages, func(item Package) string {
			return item.Name
		}, fzf.WithPrompt("select a package"))

		if err != nil {
			return "", "", fn.NewE(err)
		}

		version, err := fzf.FindOne(pkg.Versions, func(item PackageVersion) string {
			return fmt.Sprintf("%s %s", item.Version, item.Summary)
		}, fzf.WithPrompt("select a version"))

		if err != nil {
			return "", "", fn.NewE(err)
		}
		name = version.Name
		ver = version.Version
	} else {
		splits := strings.Split(name, "@")

		if strings.TrimSpace(splits[0]) == "" || strings.TrimSpace(splits[1]) == "" {
			return "", "", fn.Errorf("package %s is invalid", name)
		}
		name = splits[0]
		ver = splits[1]
	}

	return p.resolve(fmt.Sprintf("%s@%s", name, ver))
}

func (p *pkgHandler) SyncLockfile() error {
	type System struct {
		AttrPaths []string `json:"attr_paths"`
	}

	type Res struct {
		CommitHash string            `json:"commit_hash"`
		Version    string            `json:"version"`
		Systems    map[string]System `json:"systems"`
	}

	kf, err := p.fc.GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	newLock := fileclient.HashData{}
	lf, err := p.fc.GetLockfile()
	if err != nil {
		return fn.NewE(err)
	}

	for _, v := range kf.Packages {
		if hash, ok := lf.Packages[v]; ok {
			newLock[v] = hash
			continue
		}

		_, pkgHash, err := p.resolve(v)
		if err != nil {
			return fn.NewE(err)
		}

		newLock[v] = pkgHash
	}
	lf.Packages = newLock

	newLock = fileclient.HashData{}
	for _, v := range kf.Libraries {
		if hash, ok := lf.Libraries[v]; ok {
			newLock[v] = hash
			continue
		}

		_, pkgHash, err := p.resolve(v)
		if err != nil {
			return fn.NewE(err)
		}

		newLock[v] = pkgHash
	}
	lf.Libraries = newLock

	return lf.Save()
}
