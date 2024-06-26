package box

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/sshclient"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
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

func waitForPort(cxt context.Context, port int, checkInterval time.Duration) {
	for {
		if cxt.Err() != nil {
			return
		}
		if isPortOpen(port) {
			return
		}
		<-time.After(checkInterval)
	}
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh into devbox",
	Run: func(cmd *cobra.Command, _ []string) {
		if klFile, err := klfile.GetKlFile(""); err != nil {
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
				ctx, cf := context.WithCancel(cmd.Context())
				defer cf()
				go func() {
					rc, err := devbox.GetContainerLogs(ctx, selectedContainer.ID)
					if err != nil && !errors.Is(err, context.Canceled) {
						fn.PrintError(err)
						cf()
						return
					}
					defer rc.Close()
					scanner := bufio.NewScanner(rc)
					for scanner.Scan() {
						if ctx.Err() != nil {
							break
						}
						txt := scanner.Text()
						if len(txt) > 8 {
							txt = txt[8:]
						}
						if txt == "kloudlite-entrypoint:INSTALLING_PACKAGES" {
							spinner.Client.UpdateMessage("installing nix packages")
							continue
						}

						if txt == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
							spinner.Client.UpdateMessage("loading please wait")
							continue
						}
					}
					cf()
				}()
				go func() {
					connectSSH(ctx, devbox.GetSSHDomainFromPath(selectedContainer.Labels["working_dir"]), port)
					cf()
				}()
				<-ctx.Done()
			} else {
				fn.PrintError(err)
				return
			}
		} else {
			dir, _ := os.Getwd()
			if os.Getenv("IN_DEV_BOX") == "true" {
				functions.PrintError(fmt.Errorf("you are already in a devbox"))
				return
			}
			err = devbox.Start(dir, klFile)
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

			ctx, cf := context.WithCancel(cmd.Context())
			defer cf()

			go func() {
				rc, err := devbox.GetContainerLogs(ctx, c.ID)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						fmt.Printf("stream error %v", err)
						fn.PrintError(err)
					}
					cf()
					return
				}
				defer rc.Close()
				scanner := bufio.NewScanner(rc)
				for scanner.Scan() {
					if ctx.Err() != nil {
						break
					}
					txt := scanner.Text()
					if len(txt) > 8 {
						txt = txt[8:]
					}
					fmt.Println(txt)
					// if txt == "kloudlite-entrypoint:INSTALLING_PACKAGES" {
					// 	spinner.Client.UpdateMessage("installing nix packages")
					// 	continue
					// }

					// if txt == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
					// 	spinner.Client.UpdateMessage("loading please wait")
					// 	continue
					// }
					// if verbose {
					functions.Logf("%s: %s", text.Yellow("[stderr]"), txt)
					// switch stream {
					// case "stderr":
					// 	functions.Logf("%s: %s", text.Yellow("[stderr]"), txt)
					// default:
					// 	functions.Logf("%s: %s", text.Blue("[stdout]"), txt)
					// }
					// }
				}
				cf()
			}()
			go func() {
				for {
					if ctx.Err() != nil {
						return
					}
					err := sshclient.CheckSSHConnection(sshConf(devbox.GetSSHDomainFromPath(c.Labels["working_dir"]), port))
					if err != nil {
						<-time.After(1 * time.Second)
						continue
					} else {
						cf()
						return
					}
				}
			}()
			<-ctx.Done()
			connectSSH(cmd.Context(), devbox.GetSSHDomainFromPath(c.Labels["working_dir"]), port)
		}
	},
}

func sshConf(host string, port int) sshclient.SSHConfig {
	return sshclient.SSHConfig{
		User:    "kl",
		Host:    host,
		SSHPort: port,
		KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
	}
}

func connectSSH(ctx context.Context, host string, port int) {
	err := sshclient.DoSSH(sshclient.SSHConfig{
		User:    "kl",
		Host:    host,
		SSHPort: port,
		KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
	})
	if err != nil {
		fn.PrintError(err)
	}
}
