package k3s

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/errdefs"
	"github.com/kloudlite/kl/constants"
)

func dockerLabelFilter(key, value string) filters.KeyValuePair {
	return filters.Arg("label", fmt.Sprintf("%s=%s", key, value))
}

func (k *K3sClientImpl) CreateCluster(accName, name string) error {
	err := k.fc.AddCluster(accName, name)
	if err != nil {
		return err
	}
	err = k.ensureImage(constants.K3SImage)
	if err != nil {
		return err
	}
	err = k.ensureVolume(name + "_varlib")
	if err != nil {
		return err
	}
	err = k.ensureVolume(name + "_varlog")
	if err != nil {
		return err
	}
	return k.start(accName, name)
}

func (k *K3sClientImpl) stop(name string) error {
	data, err := k.dClient.ContainerInspect(context.Background(), "kl-cluster-"+name)
	if err != nil {
		if !errdefs.IsInvalidParameter(err) {
			return nil
		}
		return err
	}
	if !data.State.Running {
		return nil
	}
	timeout := 0
	err = k.dClient.ContainerStop(context.Background(), "kl-cluster-"+name, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return err
	}
	err = k.dClient.ContainerRemove(context.Background(), "kl-cluster-"+name, container.RemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (k *K3sClientImpl) start(accName, name string) error {
	data, err := k.dClient.ContainerInspect(context.Background(), "kl-cluster-"+name)
	if err == nil && data.State.Running {
		return nil
	}
	resp, err := k.dClient.ContainerCreate(context.Background(), &container.Config{
		Image: constants.K3SImage,
		Labels: map[string]string{
			"kloudlite.io/cluster":      "true",
			"kloudlite.io/cluster/name": name,
			"kloudlite.io/account/name": accName,
		},
		Cmd: []string{"server", "--tls-san=0.0.0.0"},
	}, &container.HostConfig{
		Privileged: true,
		Binds: []string{
			fmt.Sprintf("%s:/var/lib", name+"_varlib"),
			fmt.Sprintf("%s:/var/log", name+"_varlog"),
		},
	}, nil, nil, "kl-cluster-"+name)
	if err != nil {
		return err
	}
	err = k.dClient.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}
