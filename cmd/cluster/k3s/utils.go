package k3s

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func (k *K3sClientImpl) deleteVolume(i string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("deleting volume %s", i))()
	err := k.dClient.VolumeRemove(context.Background(), i, true)
	if err != nil {
		return functions.Error("could not remove volume")
	}
	return nil
}

func (c *K3sClientImpl) ensureVolume(i string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("checking volume %s", i))()
	volumeList, err := c.dClient.VolumeList(context.Background(), volume.ListOptions{})
	if err != nil {
		return err
	}
	for _, v := range volumeList.Volumes {
		if v.Name == i {
			return nil
		}
	}
	_, err = c.dClient.VolumeCreate(context.Background(), volume.CreateOptions{
		Name: i,
	})
	if err != nil {
		return functions.Error("could not create volume")
	}
	return nil

}

func (c *K3sClientImpl) imageExists(imageName string) (bool, error) {
	images, err := c.dClient.ImageList(context.Background(), image.ListOptions{
		All: true,
	})
	if err != nil {
		return false, err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			i := strings.Replace(imageName, "docker.io/", "", -1)
			if tag == i {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *K3sClientImpl) ensureImage(i string) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("checking image %s", i))()

	if imageExists, err := c.imageExists(i); err == nil && imageExists {
		return nil
	}

	out, err := c.dClient.ImagePull(context.Background(), i, image.PullOptions{})
	if err != nil {
		return functions.NewE(err, "failed to pull image")
	}
	defer out.Close()

	jsonmessage.DisplayJSONMessagesStream(out, os.Stdout, os.Stdout.Fd(), true, nil)
	return nil
}
