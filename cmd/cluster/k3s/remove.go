package k3s

import (
	"github.com/kloudlite/kl/domain/apiclient"
	"github.com/kloudlite/kl/pkg/functions"
)

// RemoveCluster implements K3SClient
func (k *K3sClientImpl) RemoveCluster(name string) error {
	c, err := k.fc.GetCluster(name)
	if err != nil {
		return err
	}
	err = apiclient.DeleteClusterReference(c.ReferenceName, functions.MakeOption("accountName", c.AccountName))
	if err != nil {
		return err
	}
	err = k.stop(name)
	if err != nil {
		return err
	}
	err = k.deleteVolume(name + "_var_lib_cni")
	if err != nil {
		return err
	}
	err = k.deleteVolume(name + "_var_lib_kubelet")
	if err != nil {
		return err
	}
	err = k.deleteVolume(name + "_var_lib_rancher_k3s")
	if err != nil {
		return err
	}
	err = k.deleteVolume(name + "_var_log")
	if err != nil {
		return err
	}
	err = k.fc.DeleteCluster(name)
	if err != nil {
		return err
	}
	return nil
}
