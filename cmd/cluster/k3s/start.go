package k3s

// StartCluster implements K3SClient.
func (k *K3sClientImpl) StartCluster(name string) error {
	fCluster, err := k.fc.GetCluster(name)
	if err != nil {
		return err
	}
	return k.start(fCluster.AccountName, name)
}
