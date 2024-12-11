package wg_vpn

import (
	"fmt"
	"io"
	"net"
	"os"
	"runtime"

	"github.com/Fa1k3n/resolvconf"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func ResetDnsServers(devName string, verbose bool) error {
	return nil
	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	ed, err := fc.GetExtraData()
	if err != nil {
		return err
	}

	bkDns := ed.GetBackupDns()

	if len(bkDns) == 0 {
		return nil
	}

	ips := make([]net.IP, 0)
	for _, v := range bkDns {
		ips = append(ips, net.ParseIP(v))
	}

	if err := setDnsServers(ips, devName, verbose); err != nil {
		return err
	}

	if err := ed.SetBackupDns([]string{}); err != nil {
		fn.PrintError(err)
	}

	return nil
}

func SetDnsServers(dnsServers []net.IP, devName string, verbose bool) error {
	return nil
	warn := func(str ...interface{}) {
		if verbose {
			fn.Warn(str)
		}
	}

	log := func(str ...interface{}) {
		if verbose {
			fn.Warn(str)
		}
	}

	if len(dnsServers) == 0 {
		warn("# dns server is not configured")
		return nil
	}

	// backup ip
	if err := func() error {
		currDns, _ := getCurrentDns(verbose)
		if len(currDns) == 0 {
			warn("# no dns server is configured to backup")
			return nil
		}

		fc, err := fileclient.New()
		if err != nil {
			return err
		}

		ed, err := fc.GetExtraData()
		if err != nil {
			return err
		}

		bkDns := ed.GetBackupDns()

		if len(bkDns) != 0 {
			return nil
		}

		for _, i := range currDns {
			found := false
			for _, j := range dnsServers {
				if j.To4().String() == i {
					found = true
					break
				}
			}
			if !found {
				dnsServers = append(dnsServers, net.ParseIP(i))
			}
		}

		return ed.SetBackupDns(currDns)
	}(); err != nil {
		return err
	}

	if verbose {
		log("# updating dns server")
	}

	return setDnsServers(dnsServers, devName, verbose)
}

func ResetSearchDomain() error {
	if runtime.GOOS == constants.RuntimeDarwin {
		_UnsetDnsSearch()
	}

	bkpPath := "/etc/resolv.conf.kl-bkp"
	resPath := "/etc/resolv.conf"

	if _, err := os.Stat(bkpPath); err == nil {
		if err := copyFile(bkpPath, resPath); err != nil {
			return err
		}
		return os.Remove(bkpPath)
	}
	return nil
}

func SetSearchDomain(domain string) error {
	if runtime.GOOS == constants.RuntimeDarwin {
		return _SetDnsSearch(domain)
	}
	// make reader for this file
	bkpPath := "/etc/resolv.conf.kl-bkp"
	resPath := "/etc/resolv.conf"

	if _, err := os.Stat(bkpPath); os.IsNotExist(err) {
		if err := copyFile(resPath, bkpPath); err != nil {
			return err
		}
	}

	reader, err := os.Open(resPath)
	if err != nil {
		return err
	}

	conf, err := resolvconf.ReadConf(reader)
	if err != nil {
		return fn.NewE(err)
	}

	f := conf.GetSearchDomains()

	fmt.Println(f)

	// Add a nameservers
	conf.Add(resolvconf.NewSearchDomain("sample.svc.cluster.local"))

	writer, err := os.Create("/etc/resolv.conf")
	if err != nil {
		return err
	}

	// Dump to stdout
	if err := conf.Write(writer); err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Flush the contents to disk to ensure they are written
	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
