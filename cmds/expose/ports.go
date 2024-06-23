package expose

import (
	"errors"
	"slices"
	"strconv"

	fn "github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/text"
	"github.com/kloudlite/kl2/utils/klfile"
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
	klFile, err := klfile.GetKlFile("")
	if err != nil {
		fn.PrintError(err)
		return errors.New("please run 'kl init' if you are not initialized the file already")
	}

	if len(args) == 0 {
		return errors.New("no ports provided. please provide ports using " + text.Yellow("kl expose port 8080 3000"))
	}

	for _, arg := range args {
		port, error := strconv.Atoi(arg)
		if error != nil {
			return errors.New("port should be an integer")
		}
		if !slices.Contains(klFile.Ports, port) {
			klFile.Ports = append(klFile.Ports, port)
		}
	}

	if err := klfile.WriteKLFile(*klFile); err != nil {
		return err
	}

	return nil
}
