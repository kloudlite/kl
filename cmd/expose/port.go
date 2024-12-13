package expose

import (
	"slices"
	"strconv"

	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var portsCmd = &cobra.Command{
	Use:   "port",
	Short: "expose ports",
	Long: `
This command will add ports to your kl-config file.
`,
	Example: ` 
  kl expose ports 8080 3000
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := exposePorts(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func exposePorts(cmd *cobra.Command, args []string) error {
	fc := clients.File
	if err != nil {
		return functions.NewE(err)
	}

	klFile, err := fc.GetKlFile()
	if err != nil {
		return functions.NewE(err)
	}

	if len(args) == 0 {
		return functions.Errorf("no ports provided. please provide ports using %s", text.Yellow("kl expose port 8080 3000"))
	}

	for _, arg := range args {
		port, err := strconv.Atoi(arg)
		if err != nil {
			return functions.NewE(err, "port should be an integer")
		}
		if !slices.Contains(klFile.Ports, port) {
			klFile.Ports = append(klFile.Ports, port)
		}
	}

	if err := klFile.Save(); err != nil {
		return functions.NewE(err)
	}

	return nil
}
