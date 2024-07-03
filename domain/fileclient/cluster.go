package fileclient

import (
	"fmt"
	"io"
	"os"

	fn "github.com/kloudlite/kl/pkg/functions"
	"sigs.k8s.io/yaml"
)

type Cluster struct {
	ReferenceName string `json:"reference"`
	Name          string `json:"name"`
	AccountName   string `json:"accountName"`
}

type ClustersFile struct {
	Clusters []*Cluster `json:"clusters"`
}

func (f *fclient) ensureClusterFile() error {
	// Check if file exists
	if _, err := os.Stat(f.configPath + "/clusters.yml"); err != nil {
		if os.IsNotExist(err) {
			// Create the file
			file, err := os.Create(f.configPath + "/clusters.yml")
			if err != nil {
				return err
			}
			defer file.Close()
			file.WriteString("clusters:\n")
		}
	}
	return nil
}

func (f *fclient) getClusters() ([]*Cluster, error) {
	if err := f.ensureClusterFile(); err != nil {
		return nil, err
	}

	file, err := os.Open(f.configPath + "/clusters.yml")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	cFile := ClustersFile{}
	err = yaml.Unmarshal(data, &cFile)
	if err != nil {
		return nil, err
	}
	return cFile.Clusters, nil
}

func (f *fclient) writeClusters(clusters []*Cluster) error {
	if err := f.ensureClusterFile(); err != nil {
		return err
	}
	cFile := ClustersFile{}
	cFile.Clusters = clusters
	data, err := yaml.Marshal(cFile)
	if err != nil {
		return err
	}
	file, err := os.Create(f.configPath + "/clusters.yml")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// AddCluster implements FileClient.
func (f *fclient) AddCluster(accName, clusterName, referenceName string) error {
	clusters, err := f.getClusters()
	if err != nil {
		return err
	}
	// check if cluster already exists
	for _, c := range clusters {
		if c.Name == clusterName {
			return fn.Error(fmt.Sprintf("cluster with name %s already exists", clusterName))
		}
	}
	clusters = append(clusters, &Cluster{
		AccountName:   accName,
		Name:          clusterName,
		ReferenceName: referenceName,
	})
	if err := f.writeClusters(clusters); err != nil {
		return err
	}
	return nil
}

// Clusters implements FileClient.
func (f *fclient) Clusters() ([]*Cluster, error) {
	return f.getClusters()
}

// DeleteCluster implements FileClient.
func (f *fclient) DeleteCluster(clusterName string) error {
	clusters, err := f.getClusters()
	if err != nil {
		return err
	}
	for i, c := range clusters {
		if c.Name == clusterName {
			clusters = append(clusters[:i], clusters[i+1:]...)
			if err := f.writeClusters(clusters); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (fc *fclient) GetCluster(clusterName string) (*Cluster, error) {
	clusters, err := fc.Clusters()
	if err != nil {
		return nil, err
	}
	for _, c := range clusters {
		if c.Name == clusterName {
			return c, nil
		}
	}
	return nil, fn.NewE(fmt.Errorf("cluster %s not found", clusterName))
}
