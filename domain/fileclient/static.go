package fileclient

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func IsOnlyPkgMode() bool {
	warnMessage := text.Yellow("session not found, but you can use as pure package manager")

	fc, err := New()
	if err != nil {
		fn.Warn(warnMessage)
		return true

	}

	if _, err = fc.GetDataContext().GetSession(); err != nil {
		fn.Warn(warnMessage)
		return true
	}

	return false
}
