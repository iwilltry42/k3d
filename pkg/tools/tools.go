/*
Copyright © 2020 The k3d Author(s)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package tools

import (
	"fmt"
	"sync"
	"time"

	k3dc "github.com/rancher/k3d/pkg/cluster"
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// LoadImagesIntoCluster starts up a k3d tools container for the selected cluster and uses it to export
// images from the runtime to import them into the nodes of the selected cluster
func LoadImagesIntoCluster(runtime runtimes.Runtime, images []string, cluster *k3d.Cluster, keepTarball bool) error {
	cluster, err := k3dc.GetCluster(cluster, runtime)
	if err != nil {
		log.Errorf("Failed to find the specified cluster")
		return err
	}

	if cluster.Network.Name == "" {
		return fmt.Errorf("Failed to get network for cluster '%s'", cluster.Name)
	}

	if _, ok := cluster.Nodes[0].Labels["k3d.cluster.volumes.imagevolume"]; !ok { // TODO: add failover solution
		return fmt.Errorf("Failed to find image volume for cluster '%s'", cluster.Name)
	}
	imageVolume := cluster.Nodes[0].Labels["k3d.cluster.volumes.imagevolume"]

	// create tools node to export images
	log.Infoln("Starting k3d-tools node...")
	toolsNode, err := startToolsNode( // TODO: re-use existing container
		runtime,
		cluster,
		cluster.Network.Name,
		[]string{
			fmt.Sprintf("%s:%s", imageVolume, k3d.DefaultImageVolumeMountPath),
			fmt.Sprintf("%s:%s", runtime.GetRuntimePath(), runtime.GetRuntimePath()),
		})
	if err != nil {
		log.Errorf("Failed to start tools container for cluster '%s'", cluster.Name)
	}

	// save image to tarfile in shared volume
	log.Infoln("Saving images...")
	tarName := fmt.Sprintf("%s/k3d-%s-images-%s.tar", k3d.DefaultImageVolumeMountPath, cluster.Name, time.Now().Format("20060102150405")) // FIXME: change
	if err := runtime.ExecInNode(toolsNode, append([]string{"./k3d-tools", "save-image", "-d", tarName}, images...)); err != nil {
		log.Errorf("Failed to save images in tools container for cluster '%s'", cluster.Name)
		return err
	}

	// import image in each node
	log.Infoln("Importing images into nodes...")
	var importWaitgroup sync.WaitGroup
	for _, node := range cluster.Nodes {
		importWaitgroup.Add(1)
		go func(node *k3d.Node, wg *sync.WaitGroup) {
			log.Infof("Importing images into node '%s'...", node.Name)
			if err := runtime.ExecInNode(node, []string{"ctr", "image", "import", tarName}); err != nil {
				log.Errorf("Failed to import images in node '%s'", node.Name)
				log.Errorln(err)
			}
			wg.Done()
		}(node, &importWaitgroup)
	}
	importWaitgroup.Wait()

	// remove tarball
	if !keepTarball {
		log.Infoln("Removing the tarball...")
		if err := runtime.ExecInNode(cluster.Nodes[0], []string{"rm", "-f", tarName}); err != nil { // TODO: do this in tools node (requires rm)
			log.Errorf("Failed to delete tarball '%s'", tarName)
			log.Errorln(err)
		}
	}

	// delete tools container
	log.Infoln("Removing k3d-tools node...")
	if err := runtime.DeleteNode(toolsNode); err != nil {
		log.Errorln("Failed to delete tools node '%s': Try to delete it manually", toolsNode.Name)
	}

	log.Infoln("...Done")

	return nil

}

// startToolsNode will start a new k3d tools container and connect it to the network of the chosen cluster
func startToolsNode(runtime runtimes.Runtime, cluster *k3d.Cluster, network string, volumes []string) (*k3d.Node, error) {
	node := &k3d.Node{
		Name:    fmt.Sprintf("%s-%s-tools", k3d.DefaultObjectNamePrefix, cluster.Name),
		Image:   k3d.DefaultToolsContainerImage,
		Role:    k3d.NoRole,
		Volumes: volumes,
		Network: network,
		Cmd:     []string{},
		Args:    []string{"noop"},
	}
	if err := runtime.CreateNode(node); err != nil {
		log.Errorf("Failed to create tools container for cluster '%s'", cluster.Name)
		return node, err
	}

	return node, nil
}
