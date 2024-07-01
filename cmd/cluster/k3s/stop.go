package k3s

// StopCluster implements K3SClient.
func (k *K3sClientImpl) StopCluster(name string) error {
	return k.stop(name)
}
