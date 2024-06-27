package boxpkg

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/server"
	"io"
	"os"

	dockerclient "github.com/docker/docker/client"
	cl "github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

type client struct {
	cli        *dockerclient.Client
	cmd        *cobra.Command
	args       []string
	foreground bool
	verbose    bool
	cwd        string

	containerName string

	env *cl.Env

	configFolder string
	userHomeDir  string
}

type BoxClient interface {
	SyncProxy(config ProxyConfig) error
	StopAll() error
	Stop() error
	Start(*cl.KLFileType) error
	Ssh() error
	Reload() error
	PrintBoxes([]Cntr) error
	ListAllBoxes() ([]Cntr, error)
	Info() error
	Exec([]string, io.Writer) error
}

func (c *client) Context() context.Context {
	return c.cmd.Context()
}

func NewClient(cmd *cobra.Command, args []string) (BoxClient, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())

	if err != nil {
		return nil, functions.NewE(err)
	}

	foreground := fn.ParseBoolFlag(cmd, "foreground")
	cwd, _ := os.Getwd()

	hash := md5.New()
	hash.Write([]byte(cwd))
	contName := fmt.Sprintf("klbox-%s", fmt.Sprintf("%x", hash.Sum(nil))[:8])
	klFile, err := cl.GetKlFile("")
	if err != nil {
		return nil, functions.NewE(err)
	}
	env, err := cl.EnvOfPath(cwd)
	fmt.Println("here3", errors.Is(err, cl.NoEnvSelected))
	if err != nil && errors.Is(err, cl.NoEnvSelected) {
		environment, err := server.GetEnvironment(klFile.AccountName, klFile.DefaultEnv)
		if err != nil {
			return nil, functions.NewE(err)
		}
		env = &cl.Env{
			Name:        environment.DisplayName,
			TargetNs:    environment.Metadata.Namespace,
			SSHPort:     0,
			ClusterName: environment.ClusterName,
		}
		data, err := cl.GetExtraData()
		if err != nil {
			return nil, functions.NewE(err)
		}
		if data.SelectedEnvs == nil {
			data.SelectedEnvs = map[string]*cl.Env{
				cwd: env,
			}
		} else {
			data.SelectedEnvs[cwd] = env
		}
		if err := cl.SaveExtraData(data); err != nil {
			return nil, functions.NewE(err)
		}
	} else if err != nil {
		return nil, functions.NewE(err)
	}

	configFolder, err := cl.GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}

	userHomeDir, err := cl.GetUserHomeDir()
	if err != nil {
		return nil, functions.NewE(err)
	}

	return &client{
		cli:           cli,
		cmd:           cmd,
		args:          args,
		foreground:    foreground,
		verbose:       flags.IsVerbose,
		cwd:           cwd,
		containerName: contName,
		env:           env,
		configFolder:  configFolder,
		userHomeDir:   userHomeDir,
	}, nil
}
