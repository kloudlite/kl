package auth

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/fileclient"
	"os"
	"strings"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/apiclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to kloudlite",
	Run: func(_ *cobra.Command, _ []string) {
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
		loginId, err := apic.CreateRemoteLogin()
		if err != nil {
			fn.PrintError(err)
			return
		}

		link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

		fn.Log(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
		fmt.Println(text.Colored(text.Blue(link), 21))
		fn.Log("\n")

		go func() {
			fn.Log("press enter to open link in browser")
			reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				fn.PrintError(err)
				return
			}
			if strings.Contains(reader, "\n") {
				err := fn.OpenUrl(link)
				if err != nil {
					fn.PrintError(err)
					return
				}
			} else {
				fn.Log("Invalid input\n")
			}
		}()

		if err = apic.Login(loginId); err != nil {
			fn.PrintError(err)
			return
		}

		if err = createClusters(apic, fc); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("successfully logged in\n")
	},
}

func createClusters(apic apiclient.ApiClient, fc fileclient.FileClient) error {
	account, err := apic.ListAccounts()
	if err != nil {
		return fn.NewE(err)
	}
	if len(account) == 0 {
		return nil
	}
	for _, a := range account {
		clusterConfig, err := apic.GetClusterConfig(a.Metadata.Name)
		if err != nil {
			return fn.NewE(err)
		}
		savedClusterConfig, err := fc.GetClusterConfig(a.Metadata.Name)
		if err != nil {
			return fn.NewE(err)
		}
		if savedClusterConfig.ClusterToken != clusterConfig.ClusterToken {
			if err = fc.SetClusterConfig(a.Metadata.Name, clusterConfig); err != nil {
				return fn.NewE(err)
			}
		}

		if err := createCluster(a.Metadata.Name, clusterConfig); err != nil {
			return fn.NewE(err)
		}
	}
	return nil
}

func createCluster(accountName string, clusterConfig *fileclient.AccountClusterConfig) error {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return fn.NewE(err)
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", boxpkg.CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-account", accountName)),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}

	//script, err := boxpkg.GenerateConnectionScript(clusterConfig)
	//if err != nil {
	//	return fn.Error("failed to generate connection script")
	//}
	//execConfig := container.ExecOptions{
	//	Cmd: []string{"sh", "-c", script},
	//}

	if existingContainers != nil && len(existingContainers) > 0 {
		if existingContainers[0].State == "running" {
			if err := cli.ContainerStop(context.Background(), existingContainers[0].ID, container.StopOptions{}); err != nil {
				return fn.Error("failed to stop container")
			}
		}
		//_, err = cli.ContainerExecCreate(context.Background(), existingContainers[0].ID, execConfig)
		//if err != nil {
		//	return fn.Error("failed to create exec")
		//}
	} else {
		_, err = cli.ContainerCreate(context.Background(), &container.Config{
			Labels: map[string]string{
				boxpkg.CONT_MARK_KEY: "true",
				"kl-k3s":             "true",
				"kl-account":         accountName,
			},
			Image: constants.GetK3SImageName(),
			Cmd:   []string{"server"},
		}, &container.HostConfig{
			Privileged:  true,
			NetworkMode: "host",
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
			Binds: []string{
				"kl-k3s-cache:/var/lib/rancher/k3s",
			},
		}, &network.NetworkingConfig{}, nil, "")
		//if err != nil {
		//	return fn.Error("failed to create container")
		//}
		//_, err = cli.ContainerExecCreate(context.Background(), createdContainer.ID, execConfig)
		//if err != nil {
		//	return fn.Error("failed to create exec")
		//}
	}

	return nil
}
