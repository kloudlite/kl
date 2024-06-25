package add

import (
	"fmt"
	"os"
	"strings"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/envhash"
	"github.com/kloudlite/kl/utils/envvars"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

var secCmd = &cobra.Command{
	Use:   "secret [name]",
	Short: "add secret references to your kl-config",
	Long:  `This command will add secret entry references from current environement to your kl-config file.`,
	Example: `
  kl add secret 		# add secret and entry by selecting from list (default)
  kl add secret [name] 	# add entry by providing secret name
	`,
	Run: func(_ *cobra.Command, args []string) {
		err := selectAndAddSecret(args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectAndAddSecret(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return functions.Error(err)
	}

	env, err := server.EnvAtPath(cwd)
	if err != nil {
		return functions.Error(err)
	}

	name := ""
	if len(args) >= 1 {
		name = args[0]
	}

	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return functions.Error(err, "please run 'kl init' if you are not initialized the file already")
	}

	secrets, err := server.ListSecrets([]fn.Option{
		fn.MakeOption("envName", env.Name),
		fn.MakeOption("accountName", klFile.AccountName),
	}...)
	if err != nil {
		return functions.Error(err)
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets created yet on server")
	}

	selectedSecretGroup := server.Secret{}
	if name != "" {
		for _, c := range secrets {
			if c.Metadata.Name == name {
				selectedSecretGroup = c
				break
			}
		}
		return functions.Error(err, "can't find secrets with provided name")

	} else {
		selectedGroup, err := fzf.FindOne(
			secrets,
			func(item server.Secret) string {
				return item.Metadata.Name
			},
			fzf.WithPrompt("Select Secret Group >"),
		)
		if err != nil {
			return functions.Error(err)
		}

		selectedSecretGroup = *selectedGroup
	}

	if len(selectedSecretGroup.StringData) == 0 {
		return fmt.Errorf("no secrets added yet to %s secret", selectedSecretGroup.Metadata.Name)
	}

	type KV struct {
		Key   string
		Value string
	}

	selectedSecretKey := &KV{}
	m := ""

	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return functions.Error(err, "map must be in format of secret_key=your_var_key")
		}

		for k, v := range selectedSecretGroup.StringData {
			if k == kk[0] {
				selectedSecretKey = &KV{
					Key:   k,
					Value: v,
				}
				break
			}
		}

		return functions.Error(err, "secret_key not found in selected secret")

	} else {
		selectedSecretKey, err = fzf.FindOne(
			func() []KV {
				var kvs []KV

				for k, v := range selectedSecretGroup.StringData {
					kvs = append(kvs, KV{
						Key:   k,
						Value: v,
					})
				}

				return kvs
			}(),
			func(val KV) string {
				return val.Key
			},
			fzf.WithPrompt(fmt.Sprintf("Select Key of %s >", selectedSecretGroup.Metadata.Name)),
		)
		if err != nil {
			return functions.Error(err)
		}
	}

	currSecs := klFile.EnvVars.GetSecrets()

	matchedGroupIndex := -1
	for i, rt := range currSecs {
		if rt.Name == selectedSecretGroup.Metadata.Name {
			matchedGroupIndex = i
			break
		}
	}

	if matchedGroupIndex != -1 {
		matchedKeyIndex := -1

		for i, ret := range currSecs[matchedGroupIndex].Env {
			if ret.RefKey == selectedSecretKey.Key {
				matchedKeyIndex = i
				break
			}
		}

		if matchedKeyIndex == -1 {
			currSecs[matchedGroupIndex].Env = append(currSecs[matchedGroupIndex].Env, envvars.ResEnvType{
				Key: renameKey(func() string {
					if m != "" {
						kk := strings.Split(m, "=")
						return kk[1]
					}
					return selectedSecretKey.Key
				}()),
				RefKey: selectedSecretKey.Key,
			})
		}
	} else {
		currSecs = append(currSecs, envvars.ResType{
			Name: selectedSecretGroup.Metadata.Name,
			Env: []envvars.ResEnvType{
				{
					Key: renameKey(func() string {
						if m != "" {
							kk := strings.Split(m, "=")
							return kk[1]
						}
						return selectedSecretKey.Key
					}()),
					RefKey: selectedSecretKey.Key,
				},
			},
		})

	}

	klFile.EnvVars.AddResTypes(currSecs, envvars.Res_secret)
	err = klfile.WriteKLFile(*klFile)
	if err != nil {
		return functions.Error(err)
	}

	fn.Log(fmt.Sprintf("added secret %s/%s to your kl-file\n", selectedSecretGroup.Metadata.Name, selectedSecretKey.Key))

	if err := envhash.SyncBoxHash(env.Name, cwd, klFile); err != nil {
		return functions.Error(err)
	}

	if !(os.Getenv("IN_DEV_BOX") == "true") {
		cwd, err := os.Getwd()
		if err != nil {
			return functions.Error(err)
		}
		_, err = devbox.ContainerAtPath(cwd)
		if err != nil && err.Error() == devbox.NO_RUNNING_CONTAINERS {
			return nil
		} else if err != nil {
			return functions.Error(err)
		}
		fn.Printf(text.Yellow("environments may have been updated. to reflect the changes, do you want to restart the container? [Y/n] "))
		if fn.Confirm("Y", "Y") {
			err = devbox.Stop(cwd)
			if err != nil {
				return functions.Error(err)
			}
			err = devbox.Start(cwd, klFile)
			if err != nil {
				return functions.Error(err)
			}
		}
	}

	return nil
}
