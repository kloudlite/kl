package fileclient

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

var (
	ErrNotLoggedIn    = fn.Errorf("not logged in")
	ErrNotFound       = fn.Errorf("not found")
	ErrDeviceNotFound = fn.Errorf("device not found")
	ErrTeamNotFound   = fn.Errorf("team not found")
	ErrK3sPortNotFound = fn.Errorf("k3s port not found")
	ErrTeamMismatch   = fn.Errorf("selected team is not same as current working directory, please change selected team using %s", text.Blue("kl use team"))

	ErrSessionNotFound = fn.Errorf("session not found")

	ErrEnvNotFound = fn.Errorf("env not found, please run `kl use env` to select an environment")

	ErrEnvNotSelected = fn.Errorf("no environment selected")

	ErrEnvMismatch = fn.Errorf("selected env is not same as current working directory env, please change selected env using %s", text.Blue("kl use env"))
)
