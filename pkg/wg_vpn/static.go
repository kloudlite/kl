package wg_vpn

import (
	"encoding/csv"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func IsSystemdReslov() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	if err := ExecCmd("systemctl status systemd-resolved", false); err != nil {
		return false
	}

	return true
}

func ExecCmd(cmdString string, verbose bool) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if verbose {
		fn.Log("[#] " + strings.Join(cmdArr, " "))
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func GenerateWgKeys() ([]byte, []byte, error) {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	return []byte(key.PublicKey().String()), []byte(key.String()), nil
}

func GeneratePublicKey(privateKey string) ([]byte, error) {
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return nil, err
	}

	return []byte(key.PublicKey().String()), nil
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
