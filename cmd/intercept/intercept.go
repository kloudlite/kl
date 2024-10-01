package intercept

import (
	"bufio"
	"fmt"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/k3s"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"os"
	"slices"
	"strconv"
	"strings"
)

var Cmd = &cobra.Command{
	Use:   "intercept",
	Short: "intercept app to tunnel trafic to your device",
	Long:  `use this command to intercept an app to tunnel trafic to your device`,
	Run: func(cmd *cobra.Command, args []string) {
		apic, err := apiclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		if err := startIntercept(apic, fc, cmd, args); err != nil {
			fn.PrintError(err)
		}
	},
}

func startIntercept(apic apiclient.ApiClient, fc fileclient.FileClient, cmd *cobra.Command, args []string) error {
	accName, err := fc.CurrentAccountName()
	if err != nil {
		return err
	}
	currentEnv, err := fc.CurrentEnv()
	if err != nil {
		return err
	}

	appsList, err := apic.ListApps(accName, currentEnv.Name)
	if err != nil {
		return err
	}

	type app struct {
		Name        string         `json:"name"`
		Port        int            `json:"port"`
		DisplayName string         `json:"displayName"`
		App         *apiclient.App `json:"app"`
	}

	var apps []app

	for _, a := range appsList {
		for _, p := range a.Spec.Services {
			apps = append(apps, app{
				Name:        a.Metadata.Name,
				DisplayName: a.DisplayName,
				Port:        p.Port,
				App:         &a,
			})
		}
	}

	if len(apps) == 0 {
		return fmt.Errorf("no apps found")
	}

	selectedApp, err := fzf.FindOne[app](apps, func(item app) string {
		return fmt.Sprintf("%s - %s:%d", item.DisplayName, item.Name, item.Port)
	}, fzf.WithPrompt("Select app to intercept "))
	if err != nil {
		return err
	}

	spinner.Client.Pause()
	fn.Printf("local port to forward %s:%d -> localhost: ", selectedApp.Name, selectedApp.Port)
	devicePortInput, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fn.PrintError(err)
	}
	devicePortInput = strings.TrimSpace(devicePortInput)
	defer spinner.Client.Resume()

	if devicePortInput == "" {
		devicePortInput = strconv.Itoa(selectedApp.Port)
	}

	devicePort, err := strconv.Atoi(devicePortInput)
	if err != nil {
		fn.PrintError(err)
	}

	var ports []apiclient.AppPort
	ports = append(ports, apiclient.AppPort{
		AppPort:    selectedApp.Port,
		DevicePort: devicePort,
	})

	err = apic.InterceptApp(selectedApp.App, true, ports, currentEnv.Name, []fn.Option{
		fn.MakeOption("appName", selectedApp.Name),
	}...)

	k3sClient, err := k3s.NewClient()
	if err != nil {
		return err
	}
	if err = k3sClient.StartAppInterceptService(ports); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	bc, err := boxpkg.NewClient(cmd, args)
	if err != nil {
		return err
	}

	kt, err := fc.GetKlFile("")
	if err != nil {
		return err
	}

	s, err := envclient.GetWorkspacePath()
	if err != nil {
		return err
	}

	eports := []int{}
	eports = append(eports, kt.Ports...)

	for _, v := range ports {
		if !slices.Contains(eports, v.DevicePort) {
			eports = append(eports, v.DevicePort)
		}
	}

	if err := bc.SyncProxy(boxpkg.ProxyConfig{
		TargetContainerPath: s,
		ExposedPorts:        eports,
	}); err != nil {
		return err
	}

	p := kt.Ports
	for _, v := range ports {
		if !slices.Contains(p, v.DevicePort) {
			p = append(p, v.DevicePort)
		}
	}
	kt.Ports = p
	if err = fc.WriteKLFile(*kt); err != nil {
		return err
	}

	fn.Log(text.Green(fmt.Sprintf("intercept app port forwarded to localhost:%v", devicePort)))
	fn.Log("Please check if vpn is connected to your device, if not please connect it using sudo kl vpn start. Ignore this message if already connected.")

	return nil
}

func init() {

	fileclient.OnlyInsideBox(Cmd)

	fileclient.OnlyInsideBox(stopCmd)
	Cmd.AddCommand(stopCmd)

}