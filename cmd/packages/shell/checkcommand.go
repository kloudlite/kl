package shell

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:    "checkchanges",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		denv, _ := clients.File.DirEnv()

		if denv == "" {
			denv = "no env"
		}

		ck, err := getCache()
		if err != nil || ck.Hash != os.Getenv("KL_HASH") {
			if err == fileclient.KLYamlNotFound {
				fmt.Printf("(%s - %s)", denv, text.Yellow("outside"))
				return
			}
			fmt.Printf("(%s - %s)", denv, text.Yellow("reload needed"))
			return
		}

		depth := 0
		ldepth, ok := os.LookupEnv("KL_DEPTH")
		if ok {
			i, err := strconv.Atoi(ldepth)
			if err == nil {
				depth = i
			}
		}
		if depth > 1 {
			fmt.Printf("%s%s%s", text.Blue(fmt.Sprintf("(%s", denv)), text.Yellow(strings.Repeat(">", depth-1)), text.Blue(")"))
			return
		}

		fmt.Printf(text.Blue(fmt.Sprintf("(%s)", denv)))
		return
	},
}
