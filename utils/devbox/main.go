package devbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/kloudlite/kl2/constants"
	"github.com/kloudlite/kl2/pkg/functions"
	"github.com/kloudlite/kl2/pkg/ui/spinner"
	"github.com/kloudlite/kl2/pkg/ui/text"
	"github.com/kloudlite/kl2/server"

	"github.com/kloudlite/kl2/utils/envhash"
	"github.com/kloudlite/kl2/utils/klfile"
	"github.com/nxadm/tail"
)

const (
	NO_RUNNING_CONTAINERS = "no container running"
)

func imageExists(cli *client.Client, imageName string) (bool, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", imageName)
	images, err := cli.ImageList(context.Background(), image.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}
	return false, nil
}

func dockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func ensureImage(i string) error {
	cli, err := dockerClient()
	if err != nil {
		return errors.New("failed to create docker client")
	}
	defer cli.Close()

	if imageExists, err := imageExists(cli, i); err == nil && imageExists {
		return nil
	}

	out, err := cli.ImagePull(context.Background(), i, image.PullOptions{})
	if err != nil {
		return errors.New("failed to pull image")
	}
	defer out.Close()

	jsonmessage.DisplayJSONMessagesStream(out, os.Stdout, os.Stdout.Fd(), true, nil)
	return nil
}

func getImageName() string {
	return fmt.Sprintf("ghcr.io/kloudlite/kl/box:%s", "v1.0.0-nightly")
}

func getFreePort() (int, error) {
	for {
		port := rand.Intn(65535-1024) + 1025
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
}

func stopContainer(path string) error {
	cli, err := dockerClient()
	if err != nil {
		return errors.New("failed to create docker client")
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
			filters.Arg("label", fmt.Sprintf("working_dir=%s", path)),
		),
		All: true,
	})
	if len(existingContainers) == 0 {
		return nil
	}

	if err != nil {
		return errors.New("failed to list containers")
	}

	timeOut := 0
	if err := cli.ContainerStop(context.Background(), existingContainers[0].ID, container.StopOptions{
		Timeout: &timeOut,
	}); err != nil {
		return err
	}

	if err := cli.ContainerRemove(context.Background(), existingContainers[0].ID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return err
	}

	return nil
}

func ensureCacheExist() error {
	cli, err := dockerClient()
	if err != nil {
		return errors.New("failed to create docker client")
	}

	caches := []string{"kl-nix-store"}

	for _, cache := range caches {
		vlist, err := cli.VolumeList(context.Background(), volume.ListOptions{
			Filters: filters.NewArgs(filters.KeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("%s=true", cache),
			}),
		})
		if err != nil {
			return err
		}

		if len(vlist.Volumes) == 0 {
			if _, err := cli.VolumeCreate(context.Background(), volume.CreateOptions{
				Labels: map[string]string{
					cache: "true",
				},
				Name: cache,
			}); err != nil {
				return err
			}
		}

	}

	return nil
}

func GetSSHDomainFromPath(pth string) string {
	s := strings.ReplaceAll(pth, xdg.Home, "")
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ":\\", "/")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", ".")
	s = strings.ReplaceAll(s, "\\", ".")
	s = strings.Trim(s, ".")
	s = fmt.Sprintf("%s.local.khost.dev", s)
	return s
}

func ensureKloudliteNetwork() error {
	cli, err := dockerClient()
	if err != nil {
		return errors.New("failed to create docker client")
	}

	networks, err := cli.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
		),
	})
	if err != nil {
		return err
	}

	if len(networks) == 0 {
		_, err := cli.NetworkCreate(context.Background(), "kloudlite", network.CreateOptions{
			Driver: "bridge",
			Labels: map[string]string{
				"kloudlite": "true",
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func ensurePublicKey() error {
	sshPath := path.Join(xdg.Home, ".ssh")
	if _, err := os.Stat(path.Join(sshPath, "id_rsa")); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", path.Join(sshPath, "id_rsa"), "-N", "")
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func setup() error {
	if err := ensurePublicKey(); err != nil {
		return err
	}

	if err := ensureCacheExist(); err != nil {
		return err
	}
	return nil
}

func generateMounts() ([]mount.Mount, string, string, error) {
	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return nil, "", "", err
	}

	if err := userOwn(td); err != nil {
		return nil, "", "", err
	}

	homeDir, err := GetUserHomeDir()
	if err != nil {
		return nil, "", "", err
	}

	sshPath := path.Join(homeDir, ".ssh", "id_rsa.pub")

	akByte, err := os.ReadFile(sshPath)
	if err != nil {
		return nil, "", "", err
	}

	ak := string(akByte)

	akTmpPath := path.Join(td, "authorized_keys")

	akByte, err = os.ReadFile(path.Join(homeDir, ".ssh", "authorized_keys"))
	if err == nil {
		ak += fmt.Sprint("\n", string(akByte))
	}

	// for wsl
	if err := func() error {
		if runtime.GOOS != constants.RuntimeLinux {
			return nil
		}

		usersPath := "/mnt/c/Users"
		_, err := os.Stat(usersPath)
		if err != nil {
			return nil
		}

		de, err := os.ReadDir(usersPath)
		if err != nil {
			return err
		}

		for _, de2 := range de {
			pth := path.Join(usersPath, de2.Name(), ".ssh", "id_rsa.pub")
			if _, err := os.Stat(pth); err != nil {
				continue
			}

			b, err := os.ReadFile(pth)
			if err != nil {
				return err
			}

			ak += fmt.Sprint("\n", string(b))
		}

		return nil
	}(); err != nil {
		return nil, "", "", err
	}
	stdErrPath := path.Join(td, "stderr.log")
	stdOutPath := path.Join(td, "stdout.log")

	if err := writeOnUserScope(stdOutPath, []byte("")); err != nil {
		return nil, "", "", err
	}

	if err := writeOnUserScope(stdErrPath, []byte("")); err != nil {
		return nil, "", "", err
	}

	if err := writeOnUserScope(akTmpPath, []byte(ak)); err != nil {
		return nil, "", "", err
	}

	configFolder, err := getConfigFolder()
	if err != nil {
		return nil, "", "", err
	}

	volumes := []mount.Mount{
		{Type: mount.TypeBind, Source: akTmpPath, Target: "/tmp/ssh2/authorized_keys", ReadOnly: true},
		{Type: mount.TypeBind, Source: stdOutPath, Target: "/tmp/stdout.log"},
		{Type: mount.TypeBind, Source: stdErrPath, Target: "/tmp/stderr.log"},
		{Type: mount.TypeVolume, Source: "kl-home-cache", Target: "/home"},
		{Type: mount.TypeVolume, Source: "kl-nix-store", Target: "/nix"},
		{Type: mount.TypeBind, Source: configFolder, Target: "/.cache/kl"},
	}

	return volumes, stdOutPath, stdErrPath, nil
}

func EnsureContainerRunning(containerId string) error {
	cli, err := dockerClient()
	if err != nil {
		return errors.New("failed to create docker client")
	}

	c, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return errors.New("failed to inspect container")
	}

	if !c.State.Running {
		return cli.ContainerStart(context.Background(), containerId, container.StartOptions{})
	}
	return nil
}

func AllWorkspaceContainers() ([]types.Container, error) {
	cli, err := dockerClient()
	if err != nil {
		return nil, errors.New("failed to create docker client")
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
		),
	})
	if err != nil {
		return nil, functions.Error(err, "failed to list containers")
	}

	return existingContainers, nil
}

func ContainerAtPath(path string) (*types.Container, error) {
	cli, err := dockerClient()
	if err != nil {
		return nil, errors.New("failed to create docker client")
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
			filters.Arg("label", fmt.Sprintf("working_dir=%s", path)),
		),
	})
	if err != nil {
		return nil, errors.New("failed to list containers")
	}
	if len(existingContainers) == 0 {
		return nil, errors.New(NO_RUNNING_CONTAINERS)
	}
	return &existingContainers[0], nil
}

var ErrNoRunningEnvironment = errors.New("no running environment")

func stopOtherContainers(path string) error {
	cli, err := dockerClient()
	if err != nil {
		return errors.New("failed to create docker client")
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
		),
	})

	if err != nil {
		for _, c := range existingContainers {
			if c.State == "running" {
				if c.Labels["working_dir"] != path {
					if err := stopContainer(c.Labels["working_dir"]); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func startContainer(path string) (string, error) {
	if err := stopOtherContainers(path); err != nil {
		return "", err
	}

	if err := setup(); err != nil {
		return "", err
	}

	cli, err := dockerClient()
	if err != nil {
		return "", errors.New("failed to create docker client")
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
			filters.Arg("label", fmt.Sprintf("working_dir=%s", path)),
		),
	})

	if err != nil {
		return "", errors.New("failed to list containers")
	}

	if len(existingContainers) > 0 {
		if existingContainers[0].State != "running" {
			return "", cli.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{})
		}

		return existingContainers[0].ID, nil
	}

	sshPort, err := getFreePort()
	if err != nil {
		return "", errors.New("failed to get free port")
	}

	vmounts, stdOut, stdErr, err := generateMounts()
	if err != nil {
		return "", err
	}

	boxhashFileName, err := envhash.BoxHashFileName(path)
	if err != nil {
		return "", err
	}

	e, err := server.EnvAtPath(path)
	if err != nil {
		return "", err
	}

	if err := envhash.SyncBoxHash(e.Name); err != nil {
		return "", err
	}

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: getImageName(),
		Labels: map[string]string{
			"kloudlite":    "true",
			"workspacebox": "true",
			"working_dir":  path,
			"ssh_port":     fmt.Sprintf("%d", sshPort),
		},
		Env: []string{
			fmt.Sprintf("KL_HASH_FILE=/.cache/kl/box-hash/%s", boxhashFileName),
			fmt.Sprintf("SSH_PORT=%d", sshPort),
			fmt.Sprintf("KL_WORKSPACE=%s", path),
		},
		Hostname:     "box",
		ExposedPorts: nat.PortSet{nat.Port(fmt.Sprintf("%d/tcp", sshPort)): {}},
	}, &container.HostConfig{
		NetworkMode: "kloudlite",
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", sshPort)): []nat.PortBinding{
				{
					HostPort: fmt.Sprintf("%d", sshPort),
				},
			},
		},
		Binds: func() []string {
			binds := make([]string, 0, len(vmounts))
			for _, m := range vmounts {
				binds = append(binds, fmt.Sprintf("%s:%s:z", m.Source, m.Target))
			}
			binds = append(binds, fmt.Sprintf("%s:/workspace:z", path))
			return binds
		}(),
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"kloudlite": {},
		},
	}, nil, "")
	if err != nil {
		fmt.Println(err)
		return "", errors.New("failed to create container")
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return "", functions.Error(err, "failed to start container")
	}

	if err := waitForContainerReady(stdOut, stdErr); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func waitForContainerReady(stdOutPath string, stdErrPath string) error {
	timeoutCtx, cf := context.WithTimeout(context.TODO(), 1*time.Minute)

	cancelFn := func() {
		defer cf()
	}

	status := make(chan int, 1)
	go func() {
		ok, err := readTillLine(timeoutCtx, stdErrPath, "kloudlite-entrypoint:CRASHED", "stderr", true, false)
		if err != nil {
			functions.PrintError(err)
			status <- 2
			cf()
			return
		}
		if ok {
			status <- 1
		}
	}()

	go func() {
		ok, err := readTillLine(timeoutCtx, stdOutPath, "kloudlite-entrypoint: SETUP_COMPLETE", "stdout", true, false)
		if err != nil {
			functions.PrintError(err)
			status <- 2
			return
		}

		if ok {
			status <- 0
		}
	}()

	select {
	case exitCode := <-status:
		{
			spinner.Client.Stop()
			cancelFn()
			if exitCode != 0 {
				readTillLine(timeoutCtx, stdOutPath, "kloudlite-entrypoint: SETUP_COMPLETE", "stdout", false, true)
				readTillLine(timeoutCtx, stdErrPath, "kloudlite-entrypoint:CRASHED", "stderr", false, true)
				return errors.New("failed to start container")
			}

			functions.Log(text.Blue("container started successfully"))
		}
	}

	return nil
}

func readTillLine(_ context.Context, file string, desiredLine, stream string, follow bool, verbose bool) (bool, error) {

	t, err := tail.TailFile(file, tail.Config{Follow: follow, ReOpen: follow, Poll: runtime.GOOS == constants.RuntimeWindows})

	if err != nil {
		return false, err
	}

	for l := range t.Lines {

		if l.Text == desiredLine {
			return true, nil
		}

		if l.Text == "kloudlite-entrypoint:INSTALLING_PACKAGES" {
			spinner.Client.UpdateMessage("installing nix packages")
			continue
		}

		if l.Text == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
			spinner.Client.UpdateMessage("loading please wait")
			continue
		}

		if verbose {
			switch stream {
			case "stderr":
				functions.Logf("%s: %s", text.Yellow("[stderr]"), l.Text)
			default:
				functions.Logf("%s: %s", text.Blue("[stdout]"), l.Text)
			}
		}

	}

	return false, nil
}

func Stop(path string) error {
	return stopContainer(path)
}

func Start(path string) error {
	env, err := server.EnvAtPath(path)
	if err != nil {
		return err
	}
	err = envhash.SyncBoxHash(env.Name)
	if err != nil {
		return err
	}
	err = ensureKloudliteNetwork()
	if err != nil {
		return err
	}

	if err := ensureImage(getImageName()); err != nil {
		return err
	}

	klConfig, err := klfile.GetKlFile(path + "/kl.yml")
	if err != nil {
		return err
	}

	containerId, err := startContainer(path)
	if err != nil {
		return err
	}

	if err = SyncProxy(ProxyConfig{
		ExposedPorts:        klConfig.Ports,
		TargetContainerId:   containerId,
		TargetContainerPath: path,
	}); err != nil {
		return err
	}

	vpnCfg, err := vpnConfigForAccount(klConfig.AccountName)
	if err != nil {
		return err
	}
	err = SyncVpn(vpnCfg)
	if err != nil {
		return err
	}

	return nil
}

func Exec(path string, command []string, out io.Writer) (int, error) {

	if len(command) == 0 {
		return 0, errors.New("command not provided")
	}

	cli, err := dockerClient()
	if err != nil {
		return 0, errors.New("failed to create docker client")
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
			filters.Arg("label", fmt.Sprintf("working_dir=%s", path)),
		),
	})
	if err != nil {
		return 0, errors.New("failed to list containers")
	}
	if len(existingContainers) == 0 {
		return 0, errors.New("container not running")
	}

	execIDResp, err := cli.ContainerExecCreate(context.Background(), existingContainers[0].ID, container.ExecOptions{
		Cmd:          command,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	})
	if err != nil {
		return 0, functions.Error(err, "failed to create exec")
	}

	execID := execIDResp.ID
	if execID == "" {
		return 0, fmt.Errorf("exec ID empty")
	}

	resp, err := cli.ContainerExecAttach(context.Background(), execID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return 0, err
	}

	defer resp.Close()
	if out == nil {
		out = os.Stdout
	}

	_, err = io.Copy(out, resp.Reader)
	if err != nil && err != io.EOF {
		return 0, err
	}

	return getExecExitCode(context.Background(), cli, execID)
}

func getExecExitCode(ctx context.Context, cli *client.Client, execID string) (int, error) {
	for {
		inspectResp, err := cli.ContainerExecInspect(ctx, execID)
		if err != nil {
			return 0, err
		}

		if !inspectResp.Running {
			return inspectResp.ExitCode, nil
		}
	}
}
