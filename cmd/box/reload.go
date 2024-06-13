package box

import (
	"crypto/md5"
	"fmt"
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/types"
	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "reload the box according to the current kl.yml configuration",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := boxpkg.NewClient(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := c.Reload(); err != nil {
			fn.PrintError(err)
			return
		}

		if err := func() error {
			if !client.InsideBox() {
				return nil
			}

			if fn.ParseBoolFlag(cmd, "skip-restart") {
				return nil
			}

			fn.Warnf("this will close the current box and restart it, are you sure?[Y/n]")
			if fn.Confirm("y", "y") {
				p, err := proxy.NewProxy(true)
				if err != nil {
					return err
				}

				dir := os.Getenv("KL_WORKSPACE")

				hash := md5.New()
				hash.Write([]byte(dir))
				contName := fmt.Sprintf("klbox-%s", fmt.Sprintf("%x", hash.Sum(nil))[:8])

				if b, err := p.RestartContainer(types.RestartBody{Name: contName}); err != nil {
					return err
				} else {
					fmt.Println(b)
				}
			}

			return nil
		}(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func init() {
	setBoxCommonFlags(reloadCmd)
	reloadCmd.Flags().BoolP("skip-restart", "s", false, "skip restarting the box")
}
