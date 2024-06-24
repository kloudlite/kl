package expose

import (
	"os"
	"slices"
	"strconv"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/klfile"
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
		if err := exposePorts(args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func exposePorts(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return functions.Error(err)
	}
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		return functions.Error(err, "please run 'kl init' if you are not initialized the file already")
	}

	if len(args) == 0 {
		return functions.Error(err, "no ports provided. please provide ports using "+text.Yellow("kl expose port 8080 3000"))
	}

	for _, arg := range args {
		port, error := strconv.Atoi(arg)
		if error != nil {
			return functions.Error(err, "port should be an integer")
		}
		if !slices.Contains(klFile.Ports, port) {
			klFile.Ports = append(klFile.Ports, port)
		}
	}

	if err := klfile.WriteKLFile(*klFile); err != nil {
		return functions.Error(err)
	}
	devbox.Start(cwd)
	return nil
}
