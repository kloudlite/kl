package auth

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/server"
	"github.com/spf13/cobra"
	"os"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(*cobra.Command, []string) {
		configFolder, err := server.GetConfigFolder()
		if err != nil {
			fn.Log(err)
			return
		}

		if err := os.RemoveAll(configFolder); err != nil {
			fn.Log(err)
			return
		}
	},
}

//func logout(configPath string) error {
//	sessionFile, err := os.Stat(path.Join(configPath, server.SessionFileName))
//	if err != nil && os.IsNotExist(err) {
//		return errors.New("not logged in")
//	}
//	if err != nil {
//		return err
//	}
//	extraDataFile, _ := os.Stat(path.Join(configPath, server.ExtraDataFileName))
//	if extraDataFile != nil {
//		if err := os.Remove(path.Join(configPath, extraDataFile.Name())); err != nil {
//			return err
//		}
//	}
//	return os.Remove(path.Join(configPath, sessionFile.Name()))
//}
