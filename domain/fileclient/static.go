package fileclient

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func IsOnlyPkgMode() bool {
	fc, err := New()
	if err != nil {
		return true
	}

	if _, err = fc.GetDataContext().GetSession(); err != nil {
		return true
	}

	fn.Warn(text.Yellow("session not found, but you can use as pure package manager"))
	return false
}
