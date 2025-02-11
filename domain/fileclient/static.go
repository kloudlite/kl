package fileclient

import "os"

func IsBoxMode() bool {
	if mode := os.Getenv("KL_BOX_MODE"); mode == "true" {
		return true
	}

	return false
}
