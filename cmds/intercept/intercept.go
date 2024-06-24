package intercept

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/server"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept app to tunnel trafic to your device",
	Long:  `use this command to intercept an app to tunnel trafic to your device`,
}

func inti() {
	Cmd.Aliases = append(Cmd.Aliases, "inc")

	startCmd.Aliases = append(startCmd.Aliases, "add", "begin", "connect")
	Cmd.AddCommand(startCmd)

	stopCmd.Aliases = append(startCmd.Aliases, "close", "end", "leave", "quit", "terminate", "exit", "remove", "disconnect")
	Cmd.AddCommand(stopCmd)

}

func EnsuseApp(apps []server.App) (*server.App, error) {
	if len(apps) == 0 {
		return nil, errors.New("no apps found")
	}

	app, err := fzf.FindOne(apps, func(item server.App) string {
		return fmt.Sprintf("%s (%s)%s", item.DisplayName, item.Metadata.Name, func() string {
			if item.IsMainApp {
				return ""
			}

			return " [external]"
		}())
	}, fzf.WithPrompt("Select App>"))
	if err != nil {
		return nil, err
	}

	return app, nil
}
