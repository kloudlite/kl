package server

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Account struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
}

func ListAccounts() ([]Account, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listAccounts", map[string]any{}, &cookie)
	if err != nil {
		return nil, err
	}

	type AccList []Account
	if fromResp, err := GetFromResp[AccList](respData); err != nil {
		return nil, err
	} else {
		return *fromResp, nil
	}
}

func SelectAccount(accountName string) (*Account, error) {
	persistSelectedAcc := func(accName string) error {
		err := client.SelectAccount(accName)
		if err != nil {
			return err
		}
		return nil
	}

	accounts, err := ListAccounts()
	if err != nil {
		return nil, err
	}

	if accountName != "" {
		for _, a := range accounts {
			if a.Metadata.Name == accountName {
				if err := persistSelectedAcc(a.Metadata.Name); err != nil {
					return nil, err
				}
				return &a, nil
			}
		}
		return nil, errors.New("you don't have access to this account")
	}

	account, err := fzf.FindOne(
		accounts,
		func(account Account) string {
			return account.DisplayName
		},
		fzf.WithPrompt("Select Account > "),
	)

	if err != nil {
		return nil, err
	}
	if err := persistSelectedAcc(account.Metadata.Name); err != nil {
		return nil, err
	}

	return account, nil
}

func EnsureAccount(options ...fn.Option) (string, error) {
	accountName := fn.GetOption(options, "accountName")
	if accountName != "" {
		return accountName, nil
	}

	var s string
	s, err := client.CurrentAccountName()
	if err != nil {
		return "", err
	}
	if s == "" {
		return "", errors.New(fmt.Sprintf("no account selected, please select an account using '%s use account'", flags.CliName))
	}

	return s, nil
}
