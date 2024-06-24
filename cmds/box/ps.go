package box

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/table"
	"github.com/kloudlite/kl2/utils/devbox"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "list all running boxes",
	Run: func(*cobra.Command, []string) {
		cs, err := devbox.AllWorkspaceContainers()
		if err != nil {
			functions.PrintError(err)
			return
		}

		// Print CS
		if err := printConts(cs); err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func printConts(conts []types.Container) error {
	if len(conts) == 0 {
		return errors.New("no containers found")
	}

	header := table.Row{
		table.HeaderText("Name"),
		table.HeaderText("State"),
	}

	rows := make([]table.Row, 0)

	for _, c := range conts {
		rows = append(rows, table.Row{strings.Join(c.Names, ", "), c.State})
	}

	fmt.Println(table.Table(&header, rows))
	table.TotalResults(len(conts), true)
	return nil
}
