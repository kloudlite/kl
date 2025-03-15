package envclient

import (
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
)

var (
	ErrNotFound = fn.Errorf("not found")
)

func IsBoxMode() bool {
	if mode := os.Getenv("KL_BOX_MODE"); mode == "true" {
		return true
	}

	return false
}

func GetDeviceNameFromEnv() (string, error) {
	if devName, b := os.LookupEnv("KL_DEVICE_NAME"); b && devName != "" {
		return devName, nil
	}

	return "", ErrNotFound
}
