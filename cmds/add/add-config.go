package add

import (
	"errors"
	"fmt"
	"os"
	"strings"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/fzf"
	"github.com/kloudlite/kl2/server"
	"github.com/kloudlite/kl2/utils"
	"github.com/kloudlite/kl2/utils/envhash"
	"github.com/kloudlite/kl2/utils/envvars"
	"github.com/kloudlite/kl2/utils/klfile"
	"github.com/spf13/cobra"
)

var confCmd = &cobra.Command{
	Use:   "config [name]",
	Short: "add config references to your kl-config",
	Long: `
This command will add config entry references from current environment to your kl-config file.
	`,
	Example: `
  kl add config 		# add config and entry by selecting from list
  kl add config [name] 		# add all entries of config by providing config name
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := selectAndAddConfig(args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func selectAndAddConfig(args []string) error {

	name := ""
	if len(args) >= 1 {
		name = args[0]
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env, err := utils.EnvAtPath(cwd)
	if err != nil {
		return err
	}

	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return err
	}

	configs, err := server.ListConfigs([]fn.Option{
		fn.MakeOption("envName", string(env)),
		fn.MakeOption("accountName", klFile.AccountName),
	}...)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return errors.New("no configs created yet on server")
	}

	selectedConfigGroup := server.Config{}

	if name != "" {
		for _, c := range configs {
			if c.Metadata.Name == name {
				selectedConfigGroup = c
				break
			}
		}
		return errors.New("can't find configs with provided name")
	} else {

		selectedGroup, e := fzf.FindOne(
			configs,
			func(item server.Config) string { return item.Metadata.Name },
			fzf.WithPrompt("Select Config Group >"),
		)
		if e != nil {
			return e
		}

		selectedConfigGroup = *selectedGroup
	}

	if len(selectedConfigGroup.Data) == 0 {
		return fmt.Errorf("no configs added yet to %s config", selectedConfigGroup.Metadata.Name)
	}

	type KV struct {
		Key   string
		Value string
	}

	selectedConfigKey := &KV{}

	m := ""
	if m != "" {
		kk := strings.Split(m, "=")
		if len(kk) != 2 {
			return errors.New("map must be in format of config_key=your_var_key")
		}

		for k, v := range selectedConfigGroup.Data {
			if k == kk[0] {
				selectedConfigKey = &KV{
					Key:   k,
					Value: v,
				}
				break
			}
		}

		return errors.New("config_key not found in selected config")

	} else {
		selectedConfigKey, err = fzf.FindOne(
			func() []KV {
				var kvs []KV

				for k, v := range selectedConfigGroup.Data {
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
			fzf.WithPrompt(fmt.Sprintf("Select Key of %s >", selectedConfigGroup.Metadata.Name)),
		)
		if err != nil {
			return err
		}
	}

	matchedGroupIndex := -1
	for i, rt := range klFile.EnvVars.GetConfigs() {
		if rt.Name == selectedConfigGroup.Metadata.Name {
			matchedGroupIndex = i
			break
		}
	}

	currConfigs := klFile.EnvVars.GetConfigs()

	if matchedGroupIndex != -1 {
		matchedKeyIndex := -1

		for i, ret := range currConfigs[matchedGroupIndex].Env {
			if ret.RefKey == selectedConfigKey.Key {
				matchedKeyIndex = i
				break
			}
		}
		if matchedKeyIndex == -1 {
			currConfigs[matchedGroupIndex].Env = append(currConfigs[matchedGroupIndex].Env, envvars.ResEnvType{
				Key: renameKey(func() string {
					if m != "" {
						kk := strings.Split(m, "=")
						return kk[1]
					}
					return selectedConfigKey.Key
				}()),
				RefKey: selectedConfigKey.Key,
			})
		}
	} else {
		currConfigs = append(currConfigs, envvars.ResType{
			Name: selectedConfigGroup.Metadata.Name,
			Env: []envvars.ResEnvType{
				{
					Key: renameKey(func() string {
						if m != "" {
							kk := strings.Split(m, "=")
							return kk[1]
						}
						return selectedConfigKey.Key
					}()),
					RefKey: selectedConfigKey.Key,
				},
			},
		})
	}

	klFile.EnvVars.AddResTypes(currConfigs, envvars.Res_config)

	err = klfile.WriteKLFile(*klFile)
	if err != nil {
		return err
	}

	fn.Log(fmt.Sprintf("added config %s/%s to your kl-file\n", selectedConfigGroup.Metadata.Name, selectedConfigKey.Key))
	env, err = utils.EnvAtPath(cwd)
	if err != nil {
		return err
	}

	if err := envhash.SyncBoxHash(string(env)); err != nil {
		return err
	}

	return nil
}
