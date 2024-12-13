package add

import (
	"fmt"
	"strings"

	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

//kl add envvar key=value

var envvarCommand = &cobra.Command{
	Use:   "envvar",
	Short: "add environment to your kl-config file",
	Long:  `add environment to your kl-config file`,
	Run: func(cmd *cobra.Command, args []string) {
		err := addEnvvar(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func addEnvvar(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fn.Errorf("wrong envvar format use key=value")
	}
	kv := strings.Split(args[0], "=")
	if len(kv) != 2 {
		return fn.Errorf("wrong envvar format use key=value")
	}

	filePath := fn.ParseKlFile(cmd)
	if filePath == "" {
		filePath = "/home/kl/workspace/kl.yml"
	}

	fc := clients.File

	kt, err := fc.GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	key := kv[0]
	value := kv[1]

	newEnv := fileclient.EnvType{
		Key:   key,
		Value: &value,
	}

	var found bool
	for i, env := range kt.EnvVars {
		if env.Key == key {
			kt.EnvVars[i].Value = &value
			found = true
			break
		}
	}

	if !found {
		kt.EnvVars = append(kt.EnvVars, newEnv)
	}

	if err = kt.Save(); err != nil {
		return fn.NewE(err)
	}

	fn.Log(text.Green(fmt.Sprintf("added envvar %s=%s to your kl-file", key, value)))

	return nil
}

func init() {
	envvarCommand.Aliases = append(envvarCommand.Aliases, "envvars", "envar")
}
