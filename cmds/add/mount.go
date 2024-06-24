package add

import (
	"errors"
	"fmt"
	"os"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/fzf"
	"github.com/kloudlite/kl2/pkg/ui/text"
	"github.com/kloudlite/kl2/server"
	"github.com/kloudlite/kl2/types"
	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/kloudlite/kl2/utils/envhash"
	"github.com/kloudlite/kl2/utils/klfile"
	"github.com/spf13/cobra"
)

var mountCommand = &cobra.Command{
	Use:   "config-mount [path]",
	Short: "add file mount to your kl-config file by selection from the all the [ config | secret ] available in current environemnt",
	Long: `
	This command will help you to add mounts to your kl-config file.
	You can add a config or secret to your kl-config file by providing the path of the config/secret you want to add.
	`,
	Example: `
  kl add config-mount [path] --config=<config_name>	# add mount from config.
  kl add config-mount [path] --secret=<secret_name>	# add secret from secret.
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := configMount(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func configMount(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || args[0] == "" {
		return errors.New("please specify the path of the config you want to add, example: kl add config-mount /tmp/sample")
	}

	path := args[0]
	c := cmd.Flag("config").Value.String()
	s := cmd.Flag("secret").Value.String()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	env, err := server.EnvAtPath(cwd)
	if err != nil {
		return err
	}

	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return err
	}

	var cors types.CSType = ""

	if c != "" || s != "" {
		if c != "" {
			cors = types.ConfigType
		} else {
			cors = types.SecretType
		}
	} else {
		csName := []types.CSType{types.ConfigType, types.SecretType}
		corsValue, err := fzf.FindOne(
			csName,
			func(item types.CSType) string {
				return string(item)
			},
			fzf.WithPrompt("Mount from Config/Secret >"),
		)
		if err != nil {
			return err
		}

		cors = types.CSType(*corsValue)
	}

	items := make([]server.ConfigORSecret, 0)
	if cors == types.ConfigType {
		configs, e := server.ListConfigs([]fn.Option{
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)

		if e != nil {
			return e
		}

		for _, c := range configs {
			items = append(items, server.ConfigORSecret{
				Entries: c.Data,
				Name:    c.Metadata.Name,
			})
		}

	} else {
		secrets, e := server.ListSecrets([]fn.Option{
			fn.MakeOption("envName", env.Name),
			fn.MakeOption("accountName", klFile.AccountName),
		}...)

		if e != nil {
			return e
		}

		for _, c := range secrets {
			items = append(items, server.ConfigORSecret{
				Entries: c.StringData,
				Name:    c.Metadata.Name,
			})
		}
	}

	if len(items) == 0 {
		return fmt.Errorf("no %ss created yet on server ", cors)
	}

	selectedItem := server.ConfigORSecret{}

	if c != "" || s != "" {
		csId := func() string {
			if c != "" {
				return c
			}
			return s
		}()

		for _, co := range items {
			if co.Name == csId {
				selectedItem = co
				break
			}
		}

		return fmt.Errorf("provided %s name not found", cors)
	} else {
		selectedItemVal, err := fzf.FindOne(
			items,
			func(item server.ConfigORSecret) string {
				return item.Name
			},
			fzf.WithPrompt(fmt.Sprintf("Select %s >", cors)),
		)

		if err != nil {
			fn.PrintError(err)
		}

		selectedItem = *selectedItemVal
	}

	matchedIndex := -1
	for i, fe := range klFile.Mounts {
		if fe.Path == path {
			matchedIndex = i
		}
	}

	key, err := fzf.FindOne(func() []string {
		res := make([]string, 0)
		for k := range selectedItem.Entries {
			res = append(res, k)
		}
		return res
	}(), func(item string) string {
		return item
	}, fzf.WithPrompt("Select Config/Secret >"))

	if err != nil {
		return err
	}

	fe := klFile.Mounts.GetMounts()

	if matchedIndex == -1 {
		fe = append(fe, types.FileEntry{
			Type: cors,
			Path: path,
			Name: selectedItem.Name,
			Key:  *key,
		})
	} else {
		fe[matchedIndex] = types.FileEntry{
			Type: cors,
			Path: path,
			Name: selectedItem.Name,
			Key:  *key,
		}
	}

	klFile.Mounts.AddMounts(fe)
	if err := klfile.WriteKLFile(*klFile); err != nil {
		return err
	}

	fn.Log("added mount to your kl-file")
	if err = envhash.SyncBoxHash(env.Name); err != nil {
		return err
	}

	if !(os.Getenv("IN_DEV_BOX") == "true") {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		_, err = devbox.ContainerAtPath(cwd)
		if err != nil && err.Error() == devbox.NO_RUNNING_CONTAINERS {
			return nil
		} else if err != nil {
			return err
		}
		fn.Printf(text.Yellow("environments may have been updated. to reflect the changes, do you want to restart the container? [Y/n] "))
		if fn.Confirm("Y", "Y") {
			err = devbox.Stop(cwd)
			if err != nil {
				fn.PrintError(err)
				return err
			}
			err = devbox.Start(cwd)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
