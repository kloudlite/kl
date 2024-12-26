package add

import (
	"fmt"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/spf13/cobra"
)

var mresCmd = &cobra.Command{
	Use:   "mres [name]",
	Short: "add managed resource references to your kl-config",
	Long: `
This command will add secret entry of managed resource references from current environement to your kl-config file.
`,
	Example: ` 
  kl add mres # add mres secret entry to your kl-config as env var
  kl add  mres [name] # add specific mres secret entry to your kl-config as env var by providing mres name
`,
	Run: func(cmd *cobra.Command, args []string) {
		apic := clients.Api

		if err := AddMres(apic, cmd, args); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func AddMres(apic apiclient.ApiClient, cmd *cobra.Command, args []string) error {

	filePath := fn.ParseKlFile(cmd)
	if filePath == "" {
		filePath = "/home/kl/workspace/kl.yml"
	}
	kt, err := apic.GetFClient().GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	//TODO: add changes to the klbox-hash file
	// mresName := fn.ParseStringFlag(cmd, "resource")

	mres, err := selectMres(apic, apic.GetFClient())

	if err != nil {
		return fn.NewE(err)
	}

	mresKey, err := selectMresKey(apic, apic.GetFClient(), mres.SecretRefName.Name)

	if err != nil {
		return fn.NewE(err)
	}

	currMreses := kt.EnvVars.GetMreses()

	if currMreses == nil {
		currMreses = []fileclient.ResType{
			{
				Name: mres.SecretRefName.Name,
				Env: []fileclient.ResEnvType{
					{
						Key:    RenameKey(fmt.Sprintf("%s_%s", mres.SecretRefName.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			},
		}
	}

	if currMreses != nil {
		matchedMres := false
		for i, rt := range currMreses {
			if rt.Name == mres.SecretRefName.Name {
				currMreses[i].Env = append(currMreses[i].Env, fileclient.ResEnvType{
					Key:    RenameKey(fmt.Sprintf("%s_%s", mres.SecretRefName.Name, *mresKey)),
					RefKey: *mresKey,
				})
				matchedMres = true
				break
			}
		}

		if !matchedMres {
			currMreses = append(currMreses, fileclient.ResType{
				Name: mres.SecretRefName.Name,
				Env: []fileclient.ResEnvType{
					{
						Key:    RenameKey(fmt.Sprintf("%s_%s", mres.SecretRefName.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			})
		}
	}

	kt.EnvVars.AddResTypes(currMreses, fileclient.Res_mres)
	if err := kt.Save(); err != nil {
		return fn.NewE(err)
	}

	fn.Log(fmt.Sprintf("added mres %s/%s to your kl-file", mres.SecretRefName.Name, *mresKey))

	fn.WarnReload()
	return nil
}

func selectMres(apic apiclient.ApiClient, fc fileclient.FileClient) (*apiclient.Mres, error) {
	currentEnv, err := apic.EnsureEnv()
	if err != nil {
		return nil, fn.NewE(err)
	}
	currentTeam, err := fc.GetDataContext().GetWsTeam()
	if err != nil {
		return nil, fn.NewE(err)
	}
	m, err := apic.ListMreses(currentTeam, currentEnv)
	if err != nil {
		return nil, fn.NewE(err)
	}
	if len(m) == 0 {
		return nil, fn.Errorf("no managed resources created yet on server")
	}

	mres, err := fzf.FindOne(m, func(item apiclient.Mres) string {
		return item.DisplayName
	}, fzf.WithPrompt("Select managed resource >"))

	return mres, err
}

func init() {
	mresCmd.Aliases = append(mresCmd.Aliases, "res", "managed-resources", "mreses")
	fn.WithKlFile(mresCmd)
}

func selectMresKey(apic apiclient.ApiClient, fc fileclient.FileClient, secretName string) (*string, error) {
	selectedTeam, err := fc.GetDataContext().GetWsTeam()
	if err != nil {
		return nil, fn.NewE(err)
	}
	secret, err := apic.GetSecret(selectedTeam, secretName)
	if err != nil {
		return nil, fn.NewE(err)
	}

	keys := []string{}
	for k := range secret.StringData {
		keys = append(keys, k)
	}

	key, err := fzf.FindOne(keys, func(item string) string {
		return item
	}, fzf.WithPrompt("Select key >"))

	return key, err
}
