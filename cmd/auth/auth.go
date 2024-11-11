package auth

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "login, logout, and get login status",
	Long:  "use this command to login, logout, and get login status",
}

func init() {
	fileclient.OnlyOutsideBox(loginCmd)
	fileclient.OnlyOutsideBox(logoutCmd)
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(authStatusCmd)
}
