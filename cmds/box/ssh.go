package box

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/sshclient"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/utils/devbox"
	"github.com/kloudlite/kl/utils/klfile"
	"github.com/spf13/cobra"
)

func isPortOpen(port int) bool {
	ln, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func waitForPort(port int, checkInterval time.Duration) {
	for {
		if isPortOpen(port) {
			return
		}
		<-time.After(checkInterval)
	}
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh into devbox",
	Run: func(*cobra.Command, []string) {
		_, err := klfile.GetKlFile("")
		if err != nil {

			if errors.Is(err, klfile.ErrorKlFileNotExists) {
				containers, err := devbox.AllWorkspaceContainers()
				if err != nil {
					fn.PrintError(err)
					return
				}
				selectedContainer, err := fzf.FindOne(containers, func(c types.Container) string {
					if c.State == "running" {
						return c.Labels["working_dir"] + "\t Active"
					}
					return c.Labels["working_dir"]
				}, fzf.WithPrompt("Select a workspace to ssh into: "))
				if err != nil {
					fn.PrintError(err)
					return
				}
				err = devbox.EnsureContainerRunning(selectedContainer.ID)
				if err != nil {
					fn.PrintError(err)
					return
				}
				port, err := strconv.Atoi(selectedContainer.Labels["ssh_port"])
				if err != nil {
					fn.PrintError(err)
					return
				}
				waitForPort(port, 100*time.Millisecond)
				connectSSH(devbox.GetSSHDomainFromPath(selectedContainer.Labels["working_dir"]), port)
			} else {
				fn.PrintError(err)
				return
			}
		} else {
			dir, _ := os.Getwd()
			if os.Getenv("IN_DEV_BOX") == "true" {
				dir = os.Getenv("KL_WORKSPACE")
			}
			err = devbox.Start(dir)
			if err != nil {
				fn.PrintError(err)
				if err := devbox.Stop(dir); err != nil {
					fn.PrintError(err)
				}
				return
			}
			c, err := devbox.ContainerAtPath(dir)
			if err != nil {
				fn.PrintError(err)
				return
			}
			port, err := strconv.Atoi(c.Labels["ssh_port"])
			if err != nil {
				fn.PrintError(err)
				return
			}
			waitForPort(port, 100*time.Millisecond)
			connectSSH(devbox.GetSSHDomainFromPath(c.Labels["working_dir"]), port)
		}
	},
}

func connectSSH(host string, port int) {

	if !isPortOpen(port) {
		fn.PrintError(fmt.Errorf("port %d is not open", port))
		return
	}

	err := sshclient.DoSSH(sshclient.SSHConfig{
		User:    "kl",
		Host:    host,
		SSHPort: port,
		KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
	})
	if err != nil {
		fn.PrintError(err)
		return
	}
}
