package shell

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:    "checkchanges",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		ck, err := getCache()
		if err != nil || ck.Hash != os.Getenv("KL_HASH") {
			fmt.Printf("(kl - %s)", text.Yellow("reload needed"))
			return
		}

		depth := 0
		s, ok := os.LookupEnv("KL_DEPTH")
		if ok {
			i, err := strconv.Atoi(s)
			if err == nil {
				depth = i
			}
		}
		if depth > 1 {
			fmt.Printf("%s%s%s", text.Blue("(kl"), text.Yellow(strings.Repeat(">", depth-1)), text.Blue(")"))
			return
		}

		fmt.Printf(text.Blue("(kl)"))
		return
	},
}
