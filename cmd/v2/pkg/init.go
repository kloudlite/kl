package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kloudlite/kl/cmd/v2/internal/lib"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "v2:pkg",
	Short: "",
}

func init() {
	Command.AddCommand(
		&cobra.Command{
			Use: "add [name]",
			// Args: func(cmd *cobra.Command, args []string) error {
			// 	if len(args) != 0 && len(args) != 2 {
			// 		return fmt.Errorf("must be either no arguments or 2 arguments")
			// 	}
			// 	return nil
			// },
			Run: func(cmd *cobra.Command, args []string) {
				k, _, err := lib.PreCommand()
				if err != nil {
					panic(err)
				}

				_, err = exec.LookPath("nix")
				if err != nil {
					panic(err)
				}

				packages := make([]string, 0, len(args))
				for i := range args {
					if !strings.HasPrefix(args[i], "nixpkgs/") {
						packages = append(packages, "nixpkgs/"+args[i])
					}
					packages = append(packages, args[i])
				}

				c := exec.Command("sh", "-c", fmt.Sprintf("nix shell %s --command echo downloaded", strings.Join(packages, " ")))

				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				c.Stdin = os.Stdin
				if err := c.Run(); err != nil {
					panic(err)
				}

				k.AddPackage(packages...)
			},
		},
		&cobra.Command{
			Use:     "remove [name]",
			Aliases: []string{"rm"},
			Run: func(cmd *cobra.Command, args []string) {
				k, _, err := lib.PreCommand()
				if err != nil {
					panic(err)
				}

				if err := k.RemovePackage(args...); err != nil {
					panic(err)
				}
			},
		},
		&cobra.Command{
			Use:     "list",
			Aliases: []string{"ls"},
			Run: func(cmd *cobra.Command, args []string) {
				k, _, err := lib.PreCommand()
				if err != nil {
					panic(err)
				}

				rows := make([]table.Row, 0, len(k.Packages))
				for i := range k.Packages {
					pp, err := lib.ParsePackage(k.Packages[i])
					if err != nil {
						panic(err)
					}
					rows = append(rows, table.Row{pp.Name})
				}

				if len(rows) == 0 {
					fmt.Printf("No Packages Installed")
					return
				}

				fmt.Print(table.Table(&table.Row{table.HeaderText("package")}, rows))
			},
		},
		&cobra.Command{
			Use: "search",
			Run: func(cmd *cobra.Command, args []string) {
			},
		},
	)
}
