package set_base_url

import (
	"github.com/kloudlite/kl2/constants"
	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/text"
	"github.com/kloudlite/kl2/utils"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "setbaseurl",
	Short:   "set base url for the cli",
	Example: fn.Desc("{cmd} status"),
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {

		if b := fn.ParseBoolFlag(cmd, "reset"); b {
			if err := utils.SaveBaseURL(constants.DefaultBaseURL); err != nil {
				fn.PrintError(err)
			} else {
				fn.Log("Base url reset successfully")
			}

			return
		}

		b := fn.ParseBoolFlag(cmd, "check")
		if b {
			fn.Println(constants.BaseURL)
			return
		}

		if len(args) == 0 {
			fn.Log(text.Yellow("Please provide a base url"))
			return
		}

		if err := utils.SaveBaseURL(args[0]); err != nil {
			fn.PrintError(err)
		} else {
			fn.Log("Base url set successfully")
		}
	},
}

func init() {
	Cmd.Flags().BoolP("check", "c", false, "check the current base url")
	Cmd.Flags().BoolP("reset", "r", false, "reset the base url to default")
}
