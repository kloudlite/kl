package vpn

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/input"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new infra vpn device",
	Long: `This command let you create new infra vpn device.
Example:
  # create new infra vpn device
  kl infra vpn new

	# creating new infra vpn device with name
	kl infra vpn  --name <device_name>
	`,
	Run: func(cmd *cobra.Command, _ []string) {
		err := newInfraDevice(cmd)
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func newInfraDevice(cmd *cobra.Command) error {

	deviceName := ""
	if cmd.Flags().Changed("device") {
		deviceName, _ = cmd.Flags().GetString("device")
	}
	if deviceName == "" {
		var err error
		deviceName, err = input.Prompt(input.Options{
			Placeholder: "Enter device name",
			CharLimit:   25,
			Password:    false,
		})
		if err != nil {
			return err
		}
	}
	if deviceName == "" {
		return errors.New("device name is required")
	}
	suggestedNames, err := server.GetInfraDeviceName(deviceName)
	if err != nil {
		return err
	}
	selectedDeviceName := ""
	if suggestedNames.Result == true {
		selectedDeviceName = deviceName
	} else {
		selectedDeviceName, err = server.SelectInfraDeviceName(suggestedNames.SuggestedNames)
		if err != nil {
			return err
		}
	}
	device, err := server.CreateInfraDevice(selectedDeviceName, deviceName)
	if err != nil {
		return err
	}
	infraContext, err := client.GetActiveInfraContext()
	if err != nil {
		return err
	}
	infraContext.DeviceName = device.Metadata.Name
	err = client.WriteInfraContextFile(*infraContext)
	if err != nil {
		return err
	}
	fn.Log(fmt.Sprintf("device %s has been created\n", device.Metadata.Name))
	return nil
}

func init() {
	newCmd.Flags().StringP("device", "d", "", "device name")
}
