package auth

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/clients"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Run: func(cmd *cobra.Command, _ []string) {

		apic := clients.Api

		loginId, err := apic.CreateRemoteLogin()
		if err != nil {
			fn.PrintError(err)
			return
		}

		link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

		fn.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fn.Println(text.Colored(text.Blue(link), 21))
		fn.Log("\n")

		if err = apic.Login(loginId); err != nil {
			fn.PrintError(err)
			return
		}

		extraData, err := apic.GetFClient().GetExtraData()
		if err != nil {
			fn.PrintError(err)
			return
		}

		HostDNSSuffix, err := apic.GetHostDNSSuffix()
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := extraData.SetDnsHostSuffix(HostDNSSuffix); err != nil {
			fn.PrintError(err)
		}

		fn.Log("successfully logged in\n")
	},
}
