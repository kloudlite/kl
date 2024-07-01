package k3s

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// ListenClusters implements K3SClient.
func (k *K3sClientImpl) ListClusters() ([]Cluster, error) {

	allContainers, err := k.dClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			dockerLabelFilter("kloudlite.io/cluster", "true"),
		),
	})
	if err != nil {
		return nil, err
	}

	clusters := make([]Cluster, 0)

	fclusters, err := k.fc.Clusters()

	for _, cluster := range fclusters {
		status := "stopped"
		for _, container := range allContainers {
			if container.Labels["kloudlite.io/cluster/name"] == cluster.Name && container.Labels["kloudlite.io/account/name"] == cluster.AccountName {
				status = "running"
				break
			}
		}
		clusters = append(clusters, Cluster{
			Name:        cluster.Name,
			AccountName: cluster.AccountName,
			Status:      status,
		})
	}

	return clusters, err
}
