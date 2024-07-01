package cluster

import (
	"github.com/kloudlite/kl/cmd/cluster/k3s"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fn.PrintError(fn.Error("cluster name is required"))
			cmd.Help()
			return
		}
		clusterName := args[0]
		accountName, err := selectAccount()
		if err != nil {
			fn.PrintError(err)
			return
		}
		clusterClient, err := k3s.NewK3sClient()
		if err != nil {
			fn.PrintError(err)
			return
		}
		err = clusterClient.CreateCluster(*accountName, clusterName)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectAccount() (*string, error) {
	if accounts, err := apiclient.ListAccounts(); err == nil {
		if selectedAccount, err := fzf.FindOne(
			accounts,
			func(account apiclient.Account) string {
				return account.Metadata.Name + " #" + account.Metadata.Name
			},
			fzf.WithPrompt("select kloudlite team > "),
		); err != nil {
			return nil, functions.NewE(err)
		} else {
			return &selectedAccount.Metadata.Name, nil
		}
	} else {
		return nil, functions.NewE(err)
	}
}

func init() {
	createCmd.Aliases = append(createCmd.Aliases, "add")
	createCmd.Flags().StringP("accountname", "a", "", "account name")
}
