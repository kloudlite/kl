package use

import (
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/k3s"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "use account",
	Run: func(_ *cobra.Command, _ []string) {
		if err := useAccount(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func useAccount() error {
	apic, err := apiclient.New()
	if err != nil {
		return fn.NewE(err)
	}
	accounts, err := apic.ListAccounts()
	if err != nil {
		return fn.NewE(err)
	}
	selectedAccount, err := fzf.FindOne(accounts, func(item apiclient.Account) string {
		return item.Metadata.Name
	}, fzf.WithPrompt("Select account to use >"))
	if err != nil {
		return err
	}
	k, err := k3s.NewClient()
	if err != nil {
		return err
	}
	if err = k.CreateClustersAccounts(selectedAccount.Metadata.Name); err != nil {
		return fn.NewE(err)
	}
	return nil
}
