package hash

import (
	"fmt"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var Cmd = &cobra.Command{
	Use:   "hash",
	Short: "hash commands",
	Long:  "generate hash for the workspace",
	Run: func(_ *cobra.Command, args []string) {
		err := generateHashContent(args)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func generateHashContent(args []string) error {
	workspace := "/workspace"
	if len(args) > 0 {
		workspace = args[0]
	}
	fc, err := fileclient.New()
	if err != nil {
		return err
	}

	apic, err := apiclient.New()
	if err != nil {
		return err
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(workspace, homePath) && workspace != "/workspace" {
		workspace = fmt.Sprintf("%s/%s", homePath, workspace)
	}

	if err := hashctrl.SyncBoxHash(apic, fc, workspace); err != nil {
		return err
	}
	return nil
}
