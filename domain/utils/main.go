package utils

import (
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func IsOnlyPkgMode() bool {
	warnMessage := text.Yellow("session not found, but you can use as pure package manager")

	fc := clients.File
	if _, err := fc.GetDataContext().GetSession(); err != nil {
		functions.Warn(warnMessage)
		return true
	}

	return false
}
