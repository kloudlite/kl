package add

import (
	"fmt"
	"os"

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

var mresCmd = &cobra.Command{
	Use:   "mres",
	Short: "add managed resource references to your kl-config",
	Long: `
This command will add secret entry of managed resource references from current environement to your kl-config file.
`,
	Example: ` 
  kl add mres # add mres secret entry to your kl-config as env var
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := AddMres(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func AddMres(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return functions.Error(err)
	}

	env, err := server.EnvAtPath(cwd)
	if err != nil {
		return functions.Error(err)
	}

	mresName := fn.ParseStringFlag(cmd, "resource")
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		fn.PrintError(err)
		return functions.Error(err, "please run 'kl init' if you are not initialized the file already")
	}

	mres, err := SelectMres([]fn.Option{
		fn.MakeOption("mresName", mresName),
		fn.MakeOption("envName", env.Name),
		fn.MakeOption("accountName", klFile.AccountName),
	}...)
	if err != nil {
		return functions.Error(err)
	}

	mresKey, err := SelectMresKey([]fn.Option{
		fn.MakeOption("mresName", mres.Metadata.Name),
		fn.MakeOption("envName", env.Name),
		fn.MakeOption("accountName", klFile.AccountName),
	}...)
	if err != nil {
		return functions.Error(err)
	}

	currMreses := klFile.EnvVars.GetMreses()
	if currMreses == nil {
		currMreses = []envvars.ResType{
			{
				Name: mres.Metadata.Name,
				Env: []envvars.ResEnvType{
					{
						Key:    renameKey(fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			},
		}
	}

	if currMreses != nil {
		matchedMres := false
		for i, rt := range currMreses {
			if rt.Name == mres.Metadata.Name {
				currMreses[i].Env = append(currMreses[i].Env, envvars.ResEnvType{
					Key:    renameKey(fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey)),
					RefKey: *mresKey,
				})
				matchedMres = true
				break
			}
		}

		if !matchedMres {
			currMreses = append(currMreses, envvars.ResType{
				Name: mres.Metadata.Name,
				Env: []envvars.ResEnvType{
					{
						Key:    renameKey(fmt.Sprintf("%s_%s", mres.Metadata.Name, *mresKey)),
						RefKey: *mresKey,
					},
				},
			})
		}
	}

	klFile.EnvVars.AddResTypes(currMreses, envvars.Res_mres)
	if err := klfile.WriteKLFile(*klFile); err != nil {
		return functions.Error(err)
	}

	fn.Log(fmt.Sprintf("added mres %s/%s to your kl-file", mres.Metadata.Name, *mresKey))

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

func SelectMresKey(options ...fn.Option) (*string, error) {
	mresName := fn.GetOption(options, "mresName")

	keys, err := server.ListMresKeys(options...)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys found in %s managed resource", mresName)
	}

	key, err := fzf.FindOne(keys, func(item string) string {
		return item
	}, fzf.WithPrompt("Select key >"))

	return key, err
}

func SelectMres(options ...fn.Option) (*server.Mres, error) {
	mresName := fn.GetOption(options, "mresName")

	m, err := server.ListMreses(options...)
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("no managed resources created yet on server")
	}

	if mresName != "" {
		for _, a := range m {
			if a.Metadata.Name == mresName {
				return &a, nil
			}
		}
		return nil, fmt.Errorf("you don't have access to this managed resource")
	}

	mres, err := fzf.FindOne(m, func(item server.Mres) string {
		return fmt.Sprintf("%s (%s)", item.DisplayName, item.Metadata.Name)
	}, fzf.WithPrompt("Select managed resource >"))

	return mres, err
}
