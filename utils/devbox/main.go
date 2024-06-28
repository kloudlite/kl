package devbox

import (
	"bufio"
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
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/server"
	"github.com/kloudlite/kl/utils/envhash"
	"github.com/kloudlite/kl/utils/klfile"
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
		return fn.Error(err, "failed to create docker client")
	}
	defer cli.Close()

	if imageExists, err := imageExists(cli, i); err == nil && imageExists {
		return nil
	}

	out, err := cli.ImagePull(context.Background(), i, image.PullOptions{})
	if err != nil {
		return fn.Error(err, "failed to pull image")
	}
	defer out.Close()

	jsonmessage.DisplayJSONMessagesStream(out, os.Stdout, os.Stdout.Fd(), true, nil)
	return nil
}

func getImageName() string {
	return fmt.Sprintf("ghcr.io/kloudlite/kl/box:%s", flags.Version)
}

func getFreePort(path string) (int, error) {
	localEnv, err := server.EnvAtPath(path)
	if err != nil {
		return 0, err
	}

	if localEnv.SShPort != 0 {
		return localEnv.SShPort, nil
	}

	defer server.SetEnvAtPath(path, localEnv)

	var resp int
	for {
		port := rand.Intn(65535-1024) + 1025
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			resp = port
			localEnv.SShPort = resp
			break
		}
	}

	return resp, nil
}

func ContainerInfo(fpath string) error {
	cli, err := dockerClient()
	if err != nil {
		return fn.Error(err, "failed to create docker client")
	}

	existingContainers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
			filters.Arg("label", fmt.Sprintf("working_dir=%s", fpath)),
		),
		All: true,
	})
	if len(existingContainers) == 0 {
		return fn.NewError("no container running in current directory")
	}

	cr := existingContainers[0]

	sshPort := cr.Labels["ssh_port"]
	fn.Println()

	table.KVOutput("User:", "kl", true)

	table.KVOutput("Name:", strings.Join(cr.Names, ", "), true)
	table.KVOutput("State:", cr.State, true)
	table.KVOutput("Path:", fpath, true)
	table.KVOutput("SSH Port:", sshPort, true)

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", GetSSHDomainFromPath(fpath)), "-p", fmt.Sprint(sshPort), "-oStrictHostKeyChecking=no"}, " ")))

	return nil
}

func stopContainer(path string) error {
	cli, err := dockerClient()
	if err != nil {
		return fn.Error(err, "failed to create docker client")
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
		return fn.Error(err, "failed to list containers")
	}

	timeOut := 0
	if err := cli.ContainerStop(context.Background(), existingContainers[0].ID, container.StopOptions{
		Timeout: &timeOut,
	}); err != nil {
		return fn.Error(err)
	}

	// if err := cli.ContainerRemove(context.Background(), existingContainers[0].ID, container.RemoveOptions{
	// 	Force: true,
	// }); err != nil {
	// 	return fn.Error(err)
	// }

	return nil
}

func ensureCacheExist() error {
	cli, err := dockerClient()
	if err != nil {
		return fn.Error(err, "failed to create docker client")
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
			return fn.Error(err)
		}

		if len(vlist.Volumes) == 0 {
			if _, err := cli.VolumeCreate(context.Background(), volume.CreateOptions{
				Labels: map[string]string{
					cache: "true",
				},
				Name: cache,
			}); err != nil {
				return fn.Error(err)
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
		return fn.Error(err, "failed to create docker client")
	}

	networks, err := cli.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
		),
	})
	if err != nil {
		return fn.Error(err)
	}

	if len(networks) == 0 {
		_, err := cli.NetworkCreate(context.Background(), "kloudlite", network.CreateOptions{
			Driver: "bridge",
			Labels: map[string]string{
				"kloudlite": "true",
			},
		})
		if err != nil {
			return fn.Error(err)
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
			return fn.Error(err)
		}
	}

	return nil
}

func setup() error {
	if err := ensurePublicKey(); err != nil {
		return fn.Error(err)
	}

	if err := ensureCacheExist(); err != nil {
		return fn.Error(err)
	}
	return nil
}

func generateMounts() ([]mount.Mount, error) {
	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return nil, err
	}

	if err := userOwn(td); err != nil {
		return nil, err
	}

	homeDir, err := GetUserHomeDir()
	if err != nil {
		return nil, err
	}

	sshPath := path.Join(homeDir, ".ssh", "id_rsa.pub")

	akByte, err := os.ReadFile(sshPath)
	if err != nil {
		return nil, err
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
			return fn.Error(err)
		}

		for _, de2 := range de {
			pth := path.Join(usersPath, de2.Name(), ".ssh", "id_rsa.pub")
			if _, err := os.Stat(pth); err != nil {
				continue
			}

			b, err := os.ReadFile(pth)
			if err != nil {
				return fn.Error(err)
			}

			ak += fmt.Sprint("\n", string(b))
		}

		return nil
	}(); err != nil {
		return nil, err
	}

	if err := writeOnUserScope(akTmpPath, []byte(ak)); err != nil {
		return nil, err
	}

	configFolder, err := getConfigFolder()
	if err != nil {
		return nil, err
	}

	volumes := []mount.Mount{
		{Type: mount.TypeBind, Source: akTmpPath, Target: "/tmp/ssh2/authorized_keys", ReadOnly: true},
		{Type: mount.TypeVolume, Source: "kl-home-cache", Target: "/home"},
		{Type: mount.TypeVolume, Source: "kl-nix-store", Target: "/nix"},
		{Type: mount.TypeBind, Source: configFolder, Target: "/.cache/kl"},
	}

	dockerSock := "/var/run/docker.sock"
	// if runtime.GOOS == constants.RuntimeWindows {
	// 	dockerSock = "\\\\.\\pipe\\docker_engine"
	// }

	volumes = append(volumes,
		mount.Mount{Type: mount.TypeVolume, Source: dockerSock, Target: "/var/run/docker.sock"},
	)

	return volumes, nil
}

func EnsureContainerRunning(containerId string) error {
	cli, err := dockerClient()
	if err != nil {
		return fn.Error(err, "failed to create docker client")
	}

	c, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return fn.Error(err, "failed to inspect container")
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
		return fn.Error(err, "failed to create docker client")
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
						return fn.Error(err)
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
			if err := cli.ContainerStart(context.Background(), existingContainers[0].ID, container.StartOptions{}); err != nil {
				return "", err
			}

			// if err := waitForContainerReady(existingContainers[0].ID); err != nil {
			// 	return "", err
			// }
		}

		return existingContainers[0].ID, nil
	}

	sshPort, err := getFreePort(path)
	if err != nil {
		return "", errors.New("failed to get free port")
	}

	vmounts, err := generateMounts()
	if err != nil {
		return "", err
	}

	boxhashFileName, err := envhash.BoxHashFileName(path)
	if err != nil {
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
			"KLCONFIG_PATH=/workspace/kl.yml",
			"KL_DNS=100.64.0.1",
			fmt.Sprintf("KL_BASE_URL=%s", constants.BaseURL),
		},
		Hostname:     "box",
		ExposedPorts: nat.PortSet{nat.Port(fmt.Sprintf("%d/tcp", sshPort)): {}},
	}, &container.HostConfig{
		Privileged:  true,
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
	}, nil, fmt.Sprintf("kl-%s", boxhashFileName[len(boxhashFileName)-8:]))
	if err != nil {
		return "", functions.Error(err, "failed to create container")
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return "", functions.Error(err, "failed to start container")
	}

	// if err := waitForContainerReady(resp.ID); err != nil {
	// 	return "", err
	// }

	return resp.ID, nil
}

func GetContainerLogs(ctx context.Context, containerId string) (io.ReadCloser, error) {
	cli, err := dockerClient()
	if err != nil {
		return nil, errors.New("failed to create docker client")
	}
	return cli.ContainerLogs(ctx, containerId, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Since:      time.Now().Format(time.RFC3339),
	})
}

func waitForContainerReady(containerId string) error {
	timeoutCtx, cf := context.WithTimeout(context.TODO(), 1*time.Minute)

	cancelFn := func() {
		defer cf()
	}

	defer cancelFn()

	status := make(chan int, 1)
	go func() {
		ok, err := readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:CRASHED", "stderr", true, false)
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
		ok, err := readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:SETUP_COMPLETE", "stdout", true, false)
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
				readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:SETUP_COMPLETE", "stdout", false, true)
				readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:CRASHED", "stderr", false, true)
				return fn.NewError("failed to start container")
			}

			// functions.Log(text.Blue("container started successfully"))
		}
	}

	return nil
}

func readTillLine(ctx context.Context, containerId string, desiredLine, stream string, follow bool, verbose bool) (bool, error) {

	cli, err := dockerClient()
	if err != nil {
		return false, err
	}

	cout, err := cli.ContainerLogs(ctx, containerId, container.LogsOptions{
		ShowStdout: func() bool {
			return stream == "stdout"
		}(),
		ShowStderr: func() bool {
			return stream == "stderr"
		}(),
		Follow: true,
		Since:  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(cout)

	for scanner.Scan() {
		txt := scanner.Text()

		if len(txt) > 8 {
			txt = txt[8:]
		}

		if txt == desiredLine {
			return true, nil
		}
		if txt == "kloudlite-entrypoint:INSTALLING_PACKAGES" {
			spinner.Client.UpdateMessage("installing nix packages")
			continue
		}

		if txt == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
			spinner.Client.UpdateMessage("loading please wait")
			continue
		}

		if verbose {
			switch stream {
			case "stderr":
				functions.Logf("%s: %s", text.Yellow("[stderr]"), txt)
			default:
				functions.Logf("%s: %s", text.Blue("[stdout]"), txt)
			}
		}

	}

	return false, nil
}

func Stop(path string) error {
	return stopContainer(path)
}

func Restart(fpath string, klConfig *klfile.KLFileType) error {
	if err := Stop(fpath); err != nil {
		return err
	}
	return Start(fpath, klConfig)
}

func Start(fpath string, klConfig *klfile.KLFileType) error {
	env, err := server.EnvAtPath(fpath)
	if err != nil {
		return functions.Error(err)
	}

	err = ensureKloudliteNetwork()
	if err != nil {
		return functions.Error(err)
	}

	if err := ensureImage(getImageName()); err != nil {
		return fn.Error(err)
	}

	boxHash, err := envhash.BoxHashFile(fpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = envhash.SyncBoxHash(env.Name, fpath, klConfig)
			if err != nil {
				return functions.Error(err)
			}
		}
		return functions.Error(err)
	} else {
		klconfHash, err := envhash.GenerateKLConfigHash(klConfig)
		if err != nil {
			return functions.Error(err)
		}
		if klconfHash != boxHash.KLConfHash {
			err = envhash.SyncBoxHash(env.Name, fpath, klConfig)
			if err != nil {
				return functions.Error(err)
			}
		}
	}

	_, err = startContainer(fpath)
	if err != nil {
		return fn.Error(err)
	}

	if err = SyncProxy(ProxyConfig{
		ExposedPorts:        klConfig.Ports,
		TargetContainerPath: fpath,
	}); err != nil {
		return fn.Error(err)
	}

	vpnCfg, err := GetAccVPNConfig(klConfig.AccountName)
	if err != nil {
		return functions.Error(err)
	}

	err = SyncVpn(vpnCfg.WGconf)
	if err != nil {
		return functions.Error(err)
	}

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", GetSSHDomainFromPath(fpath)), "-p", fmt.Sprint(env.SShPort), "-oStrictHostKeyChecking=no"}, " ")))

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
		return 0, fn.Error(err, "failed to create exec")
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
