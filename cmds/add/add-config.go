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
	Run: func(_ *cobra.Command, args []string) {
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
		return functions.Error(err)
	}

	env, err := server.EnvAtPath(cwd)
	if err != nil {
		return functions.Error(err)
	}

	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return functions.Error(err)
	}

	configs, err := server.ListConfigs([]fn.Option{
		fn.MakeOption("envName", env.Name),
		fn.MakeOption("accountName", klFile.AccountName),
	}...)
	if err != nil {
		return functions.Error(err)
	}

	if len(configs) == 0 {
		return functions.Error(err, "no configs created yet on server")
	}

	selectedConfigGroup := server.Config{}

	if name != "" {
		for _, c := range configs {
			if c.Metadata.Name == name {
				selectedConfigGroup = c
				break
			}
		}
		return functions.Error(err, "can't find configs with provided name")
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
			return functions.Error(err, "map must be in format of config_key=your_var_key")
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

		return functions.Error(err, "config_key not found in selected config")
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
			return functions.Error(err)
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
		return functions.Error(err)
	}

	fn.Log(fmt.Sprintf("added config %s/%s to your kl-file\n", selectedConfigGroup.Metadata.Name, selectedConfigKey.Key))
	env, err = server.EnvAtPath(cwd)
	if err != nil {
		return functions.Error(err)
	}

	if err := envhash.SyncBoxHash(env.Name, cwd); err != nil {
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
			err = devbox.Start(cwd)
			if err != nil {
				return functions.Error(err)
			}
		}
	}

	return nil
}
