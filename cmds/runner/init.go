package runner

import (
	"errors"
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/klfile"

	"github.com/spf13/cobra"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "initialize a kl-config file",
	Long:  `use this command to initialize a kl-config file`,
	Run: func(_ *cobra.Command, _ []string) {
		if os.Getenv("IN_DEV_BOX") == "true" {
			fn.PrintError(errors.New("cannot re-initialize workspace in dev box"))
			return
		}
		_, err := klfile.GetKlFile("")
		if err == nil {
			fn.Printf(text.Yellow("Workspace is already initilized. Do you want to override? (y/N): "))
			if !fn.Confirm("Y", "N") {
				return
			}
		} else if !errors.Is(err, klfile.ErrorKlFileNotExists) {
			fn.PrintError(err)
			return
		}

		if selectedAccount, err := selectAccount(); err != nil {
			fn.PrintError(err)
			return
		} else {
			if selectedEnv, err := selectEnv(*selectedAccount); err != nil {
				fn.PrintError(err)
			} else {
				newKlFile := klfile.KLFileType{
					AccountName: *selectedAccount,
					DefaultEnv:  *selectedEnv,
					Version:     "v1",
					Packages: []string{"neovim","git"},
				}
				if err := klfile.WriteKLFile(newKlFile); err != nil {
					fn.PrintError(err)
				} else {
					fn.Printf(text.Green("Workspace initialized successfully.\n"))
				}
			}
		}

	},
}

func selectAccount() (*string, error) {
	if accounts, err := server.ListAccounts(); err == nil {
		if selectedAccount, err := fzf.FindOne(
			accounts,
			func(account server.Account) string {
				return account.Metadata.Name + " #" + account.Metadata.Name
			},
			fzf.WithPrompt("select kloudlite team > "),
		); err != nil {
			return nil, err
		} else {
			return &selectedAccount.Metadata.Name, nil
		}
	} else {
		return nil, err
	}
}

func selectEnv(accountName string) (*string, error) {
	if accounts, err := server.ListEnvs(
		fn.Option{
			Key:   "accountName",
			Value: accountName,
		},
	); err == nil {
		if selectedEnv, err := fzf.FindOne(
			accounts,
			func(env server.Env) string {
				return env.Metadata.Name + " #" + env.Metadata.Name
			},
			fzf.WithPrompt("select environment > "),
		); err != nil {
			return nil, err
		} else {
			return &selectedEnv.Metadata.Name, nil
		}
	} else {
		return nil, err
	}
}

func init() {
	InitCommand.Flags().StringP("account", "a", "", "account name")
	InitCommand.Flags().StringP("file", "f", "", "file name")
}
