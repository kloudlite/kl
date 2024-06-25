package box

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "box",
	Short: "start, stop, reload, ssh and get running box info",
}

func init() {
	Cmd.AddCommand(psCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(execCmd)
	Cmd.AddCommand(sshCmd)
	Cmd.AddCommand(infoCmd)
	Cmd.AddCommand(restartCmd)
}
