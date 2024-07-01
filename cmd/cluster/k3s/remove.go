package k3s

// RemoveCluster implements K3SClient.
func (k *K3sClientImpl) RemoveCluster(name string) error {
	err := k.stop(name)
	if err != nil {
		return err
	}
	err = k.deleteVolume(name + "_varlib")
	if err != nil {
		return err
	}
	err = k.deleteVolume(name + "_varlog")
	if err != nil {
		return err
	}

	err = k.fc.DeleteCluster(name)
	if err != nil {
		return err
	}
	return nil
}
