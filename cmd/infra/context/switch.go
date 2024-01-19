package context

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch infra context",
	Long: `This command let switch between infra contexts.
Example:
  # switch to existing infra context by selecting one from list 
  kl infra context switch

	# switch to infra context with infra context name
	kl infra context switch --name <infra context_name>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		err := switchInfraContext(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func switchInfraContext(cmd *cobra.Command) error {

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

	if name == "" {
		return errors.New("infra context name is required")
	}

	if _, ok := c.InfraContexts[name]; !ok {
		return errors.New(fmt.Sprintf("infra context %s not found", name))
	}

	if err := client.SetActiveInfraContext(name); err != nil {
		return err
	}
	return nil
}

func init() {
	switchCmd.Flags().StringP("name", "n", "", "infra context name")
	switchCmd.Aliases = append(switchCmd.Aliases, "sw")
}
