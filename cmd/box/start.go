package box

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

type EnvironmentVariable struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

type KLConfigType struct {
	Packages []string              `yaml:"packages" json:"packages"`
	EnvVars  []EnvironmentVariable `yaml:"envVars" json:"envVars"`
	Mounts   map[string]string     `yaml:"mounts" json:"mounts"`
}

type VolMount struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type FileMountType struct {
	FileMount client.MountType `yaml:"filemount" json:"filemount"`
}

var imageName string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := startBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startBox(cmd *cobra.Command, args []string) error {
	foreground := fn.ParseBoolFlag(cmd, "foreground")
	debug := fn.ParseBoolFlag(cmd, "debug")

	cont, err := getRunningContainer()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if cwd != cont.Path && cont.Path != "" {
		fn.Log(fmt.Sprintf("container is already running in workspace %s, \nDo you want to stop it and start a new container with current workspace? [y/N] ", cont.Path))
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" {
			return errors.New("could not start container in current workspace")
		}
		if err := stopBox(cmd, args); err != nil {
			return err
		}
	}

	s := spinner.NewSpinner("starting container please wait")

	imageName = constants.BoxDockerImage
	if len(args) > 0 {
		imageName = args[0]
	}

	if debug {
		fn.Logf("starting container in: %s", text.Blue(cwd))
	}

	{
		// Global setup
		ensurePublicKey()
		ensureCacheExist()
	}

	{

		envs, mmap, err := server.GetLoadMaps()
		if err != nil {
			return err
		}

		// local setup
		kConf, err := loadConfig()
		if err != nil {
			return err
		}

		fMounts, err := loadFileMount(mmap)
		if err != nil {
			return err
		}

		var ev []EnvironmentVariable
		for k, v := range envs {
			ev = append(ev, EnvironmentVariable{k, v})
		}

		kConf.EnvVars = ev
		if kConf.EnvVars == nil {
			kConf.EnvVars = []EnvironmentVariable{}
		}
		kConf.Mounts = fMounts

		if err := ensureBoxExist(*kConf, foreground, debug); err != nil {
			return err
		}

		s.Start()
		ensureBoxRunning()
		defer s.Stop()

	}

	if debug {
		fn.Logf("Started container of: %s", text.Blue(cwd))
	}

	return nil
}

func loadFileMount(mm server.MountMap) (map[string]string, error) {
	kt, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}

	fm := map[string]string{}

	for _, fe := range kt.FileMount.Mounts {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		fm[pth] = mm[pth]
	}

	return fm, nil
}

func loadConfig() (*KLConfigType, error) {
	kf, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}

	// read kl.yml into struct
	klConfig := &KLConfigType{
		Packages: kf.Packages,
	}
	return klConfig, nil
}

func getCwdHash() string {
	cwd, _ := os.Getwd()
	hash := md5.New()
	hash.Write([]byte(cwd))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func ensurePublicKey() {
	sshPath := path.Join(xdg.Home, ".ssh")
	if _, err := os.Stat(path.Join(sshPath, "id_rsa")); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", path.Join(sshPath, "id_rsa"), "-N", "")
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}

func ensureCacheExist() {
	command := exec.Command("docker", "volume", "create", "nix-store")
	err := command.Run()
	if err != nil {
		fn.PrintError(errors.New("error creating nix-store cache volume"))
	}

	command = exec.Command("docker", "volume", "create", "kl-home-cache")
	err = command.Run()
	if err != nil {
		fn.PrintError(errors.New("error creating home cache volume"))
	}
}

func ensureBoxExist(klConfig KLConfigType, foreground, debug bool) error {
	containerName := "kl-box-" + getCwdHash()
	cwd, _ := os.Getwd()
	o, err := exec.Command("docker", "inspect", containerName).Output()
	startContainer := func() error {
		conf, err := json.Marshal(klConfig)
		if err != nil {
			return err
		}

		dockerArgs := []string{"run"}
		if !foreground {
			dockerArgs = append(dockerArgs, "-d")
		}

		sshPath := path.Join(xdg.Home, ".ssh", "id_rsa.pub")

		akByte, err := os.ReadFile(sshPath)
		if err != nil {
			return err
		}

		ak := string(akByte)

		td, err := os.MkdirTemp("", "kl-tmp")
		if err != nil {
			return err
		}
		akTmpPath := path.Join(td, "authorized_keys")

		akByte, err = os.ReadFile(path.Join(xdg.Home, ".ssh", "authorized_keys"))
		if err == nil {
			ak += fmt.Sprint("\n", string(akByte))
		}

		if err := os.WriteFile(akTmpPath, []byte(ak), fs.ModePerm); err != nil {
			return err
		}

		switch runtime.GOOS {
		case constants.RuntimeWindows:
			fn.Warn("docker support inside container not implemented yet")
		default:
			dockerArgs = append(dockerArgs, "-v", "/var/run/docker.sock:/var/run/docker.sock:ro")
		}

		if runtime.GOOS != constants.RuntimeWindows {
			dockerArgs = append(dockerArgs, "-v", "/var/run/docker.sock:/var/run/docker.sock:ro")
		}

		cwd, _ := os.Getwd()

		dockerArgs = append(dockerArgs,
			"--name", containerName,
			"--label", "kl-box=true",
			"--label", fmt.Sprintf("kl-box-cwd=%s", cwd),
			"-v", fmt.Sprintf("%s:/tmp/ssh2/authorized_keys:ro", akTmpPath),
			"-v", "kl-home-cache:/home:rw",
			"-v", "nix-store:/nix:rw",
			"--hostname", "box",
			"-v", fmt.Sprintf("%s:/home/kl/workspace:rw", cwd),
			"-p", "1729:22",
			imageName, "--", string(conf),
		)

		command := exec.Command("docker", dockerArgs...)

		// command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if debug {
			fn.Logf("docker container started with cmd: %s\n", command.String())
		}

		out, err := command.Output()
		if err != nil {
			return fmt.Errorf("error running kl-box container [%s]", err.Error())
		}

		if debug {
			fn.Logf("docker container started with output: %s\n", string(out))
		}

		return nil
	}

	if err != nil {
		return startContainer()
	} else {
		// Get all volume mounts
		type Container struct {
			Mounts []struct {
				Type        string `json:"Type"`
				Source      string `json:"Source"`
				Destination string `json:"Destination"`
			}
		}
		var containers []Container
		err := json.Unmarshal(o, &containers)
		if err != nil {
			return fmt.Errorf("error parsing docker inspect output [%s]", err.Error())
		}
		for _, container := range containers {
			for _, mount := range container.Mounts {
				if mount.Destination == "/home/kl/workspace" {
					if fmt.Sprintf("/host_mnt%s", cwd) != mount.Source {
						fn.Warn("kl-box is running with a different workspace.")
					} else {
						return nil
					}
				}
			}
		}

		fn.Log("Do you want to reload with current workspace? [y/N] ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" {
			return nil
		}
		fn.Log("Reloading kl-box container...")
		command := exec.Command(
			"docker",
			"stop", containerName)
		err = command.Run()
		if err != nil {
			fn.PrintError(errors.New("error stopping kl-box container"))
		}
		command = exec.Command(
			"docker",
			"rm", containerName)
		err = command.Run()
		if err != nil {
			fn.PrintError(errors.New("error removing kl-box container"))
		}
		return startContainer()
	}

}

func ensureBoxRunning() {
	containerName := "kl-box-" + getCwdHash()
	command := exec.Command("docker", "start", containerName)
	err := command.Run()
	if err != nil {
		fn.PrintError(errors.New("error starting kl-box container"))
	}
}

func init() {
	startCmd.Flags().BoolP("debug", "d", false, "run in debug mode")
	startCmd.Flags().BoolP("foreground", "f", false, "run in foreground mode")
}
