package vpn

import (
	"errors"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

var maps []string
var deleteFlag bool

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "expose port of selected device",
	Long: `expose port
Examples:
  # expose port of selected device
	kl vpn expose port -p <port>:<your_local_port>

  # delete exposed port of selected device
	kl vpn expose port -d -p <port>:<your_local_port> 
`,
	Run: func(_ *cobra.Command, _ []string) {
		err := exposeVPNPorts()
		if err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func exposeVPNPorts() error {

	if len(maps) == 0 {
		return errors.New("no port maps provided")
	}

	ports := make([]server.DevicePort, 0)

	for _, v := range maps {
		mp := strings.Split(v, ":")
		if len(mp) != 2 {
			return errors.New("wrong map format use <server_port>:<local_port> eg: 80:3000")
		}

		pp, err := strconv.ParseInt(mp[0], 10, 32)
		if err != nil {
			return err
		}

		tp, err := strconv.ParseInt(mp[1], 10, 32)
		if err != nil {
			return err
		}

		ports = append(ports, server.DevicePort{
			Port:       int(pp),
			TargetPort: int(tp),
		})
	}

	if !deleteFlag {
		if err := server.UpdateDevice(ports); err != nil {
			return err
		}

		fn.Log("ports exposed")
	} else {
		if err := server.DeleteDevicePort(ports); err != nil {
			return err
		}

		fn.Log("ports deleted")
	}

	return nil
}

func init() {
	exposeCmd.Flags().StringArrayVarP(
		&maps, "port", "p", []string{},
		"expose port <server_port>:<local_port>",
	)
	exposeCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "delete ports")
}
