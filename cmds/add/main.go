package add

import (
	"fmt"
	"regexp"
	"strings"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "add",
	Short: "add environment resources to your kl-config file",
	Long:  "This command will add the environment resources to your kl-config file",
}

func init() {
	Command.AddCommand(confCmd)
	Command.AddCommand(mresCmd)
	Command.AddCommand(secCmd)
	mountCommand.Flags().StringP("config", "c", "", "config name")
	mountCommand.Flags().StringP("secret", "s", "", "secret name")
	Command.AddCommand(mountCommand)
}

func renameKey(key string) string {
	regexPattern := `[^a-zA-Z0-9]`

	regexpCompiled, err := regexp.Compile(regexPattern)
	if err != nil {
		fn.Log(text.Yellow(fmt.Sprintf("[#] error compiling regex pattern: %s", regexPattern)))
		return key
	}

	resultString := regexpCompiled.ReplaceAllString(key, "_")

	return strings.ToUpper(resultString)
}
