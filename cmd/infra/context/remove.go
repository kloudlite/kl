package context

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove existing infra context from infra contexts list",
	Long: `This command let you remove existing infra context from infra contexts list.
Example:
  # remove infra context
  kl infra context [remove|rm] <infra_context_name>

	# interactive delete infra context
  kl infra context [remove|rm]
	`,
	Run: func(cmd *cobra.Command, _ []string) {

		err := removeInfraContext(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func removeInfraContext(cmd *cobra.Command) error {

	name := ""

	if cmd.Flags().Changed("name") {
		name, _ = cmd.Flags().GetString("name")
	}

	c, err := client.GetInfraContexts()
	if err != nil {
		return err
	}

	if len(c.InfraContexts) == 0 {
		return errors.New("no infra context found")
	}

	if name == "" {
		n, err := fzf.FindOne(func() []string {
			var infraContexts []string
			for _, ctx := range c.InfraContexts {
				infraContexts = append(infraContexts, ctx.Name)
			}
			return infraContexts
		}(), func(item string) string {
			return item
		})
		if err != nil {
			return err
		}

		name = *n
	}

	if err := client.DeleteInfraContext(name); err != nil {
		return err
	}

	fn.Log(fmt.Sprintf("infra context %s removed", name))
	return nil
}

func init() {
	removeCmd.Aliases = append(removeCmd.Aliases, "rm")
	removeCmd.Flags().StringP("name", "n", "", "infra context name")
}
