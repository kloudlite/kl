package auth

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Long: `This command let you login to the kloudlite.
Example:
  # Login to kloudlite
  kl login 

  when you execute the above command a link will be opened on your browser. 
  visit your browser and approve there to access your account using this cli.
	`,
	Run: func(_ *cobra.Command, _ []string) {
		err := login()
		if err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func login() error {

	loginId, err := server.CreateRemoteLogin()
	if err != nil {
		return err
	}

	link := text.Blue(fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId))

	functions.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
	functions.Println(text.Colored(link, 21))
	functions.Log("\n")

	if err = server.Login(loginId); err != nil {
		functions.PrintError(err)
		return err
	}

	functions.Log("successfully logged in\n")
	return nil

}
