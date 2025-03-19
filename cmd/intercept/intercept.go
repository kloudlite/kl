package intercept

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/clients"
	"github.com/kloudlite/kl/domain/envclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept service to tunnel trafic to your device",
	Long:  `use this command to intercept an service to tunnel trafic to your device`,
	Run: func(cmd *cobra.Command, args []string) {
		apic := clients.Api
		if err := startIntercept(apic); err != nil {
			fn.PrintError(err)
		}
	},
}

func startIntercept(apic apiclient.ApiClient) error {

	team, err := apic.GetFClient().GetDataContext().GetTeam()
	if err != nil {
		return err
	}

	currentEnv, err := apic.GetFClient().DirEnv()
	if err != nil {
		return err
	}

	servicesList, err := apic.ListServices(team, currentEnv)
	if err != nil {
		return err
	}

	type service struct {
		Ip       string             `json:"name"`
		Port     int                `json:"port"`
		Hostname string             `json:"displayName"`
		Service  *apiclient.Service `json:"service"`
	}

	var services []service

	for i := range servicesList {
		a := servicesList[i]
		for j := range a.Spec.Ports {
			services = append(services, service{
				Ip:       a.Metadata.Name,
				Hostname: a.Spec.Hostname,
				Port:     a.Spec.Ports[j].Port,
				Service:  &a,
			})
		}
	}

	if len(services) == 0 {
		return fn.Errorf("no services found")
	}

	selectedService, err := fzf.FindOne(services, func(item service) string {
		return fmt.Sprintf("%s:%d", item.Service.Spec.ServiceRef.Name, item.Port)
	}, fzf.WithPrompt("Select service to intercept "))
	if err != nil {
		return err
	}

	spinner.Client.Pause()
	fn.Printf("local port to forward %s: %d -> localhost: ", selectedService.Service.Spec.ServiceRef.Name, selectedService.Port)
	devicePortInput, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fn.PrintError(err)
	}
	devicePortInput = strings.TrimSpace(devicePortInput)
	defer spinner.Client.Resume()

	if devicePortInput == "" {
		devicePortInput = strconv.Itoa(selectedService.Port)
	}

	devicePort, err := strconv.Atoi(devicePortInput)
	if err != nil {
		fn.PrintError(err)
	}

	var ports []apiclient.ServicePort
	ports = append(ports, apiclient.ServicePort{
		ServicePort: selectedService.Port,
		DevicePort:  devicePort,
	})

	if err = apic.InterceptService(selectedService.Service, true, ports, currentEnv, []fn.Option{
		fn.MakeOption("serviceName", selectedService.Hostname),
	}...); err != nil {
		return err
	}

	fn.Log(text.Green(fmt.Sprintf("intercept service port forwarded to localhost:%v", devicePort)))

	if !envclient.IsBoxMode() {
		fn.Log("Please check if vpn is connected to your device, if not please connect it using sudo kl vpn start. Ignore this message if already connected.")
	}

	return nil
}

func init() {
	Cmd.AddCommand(stopCmd)
}
