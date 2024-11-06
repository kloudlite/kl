package env

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "env",
	Short: "suspend and resume environment",
}

func init() {
	Cmd.AddCommand(pauseCmd)
	Cmd.AddCommand(resumeCmd)
	Cmd.Aliases = append(Cmd.Aliases, "environment")
	pauseCmd.Aliases = append(pauseCmd.Aliases, "suspend")
}
