package auth

import (
	"bufio"
	"fmt"
	"github.com/kloudlite/kl/domain/fileclient"
	"os"
	"strings"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Run: func(_ *cobra.Command, _ []string) {
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		loginId, err := apic.CreateRemoteLogin()
		if err != nil {
			fn.PrintError(err)
			return
		}

		link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

		fn.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fmt.Println(text.Colored(text.Blue(link), 21))
		fn.Log("\n")

		go func() {
			fn.Log("press enter to open link in browser")
			reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				fn.PrintError(err)
				return
			}
			if strings.Contains(reader, "\n") {
				err := fn.OpenUrl(link)
				if err != nil {
					fn.PrintError(err)
					return
				}
			} else {
				fn.Log("Invalid input\n")
			}
		}()

		if err = apic.Login(loginId); err != nil {
			fn.PrintError(err)
			return
		}

		if err = createClustersAccounts(apic, fc); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("successfully logged in\n")
	},
}

func createClustersAccounts(apic apiclient.ApiClient, fc fileclient.FileClient) error {
	account, err := apic.ListAccounts()
	if err != nil {
		return fn.NewE(err)
	}
	if len(account) == 0 {
		return nil
	}
	for _, a := range account {
		// TODO(nxtCoder36): remove below condition, was only for testing
		if a.Metadata.Name != "development-team" {
			continue
		}
		clusterConfig, err := apic.GetClusterConfig(a.Metadata.Name)
		if err != nil {
			return fn.NewE(err)
		}
		savedClusterConfig, err := fc.GetClusterConfig(a.Metadata.Name)

		if err != nil {
			return fn.NewE(err)
		}
		if savedClusterConfig.ClusterToken != clusterConfig.ClusterToken {
			if err = fc.SetClusterConfig(a.Metadata.Name, clusterConfig); err != nil {
				return fn.NewE(err)
			}
		}
	}
	return nil
}
