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
package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// CreateNetworkIfNotPresent creates a new docker network
// @return: network name, exists, error
func (d Docker) CreateNetworkIfNotPresent(name string) (string, bool, error) {

	// (0) create new docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Errorln("Failed to create docker client")
		return "", false, err
	}

	// (1) configure list filters
	args := GetDefaultObjectLabelsFilter("")
	args.Del("label", fmt.Sprintf("k3d.cluster=%s", "")) // We don't filter for cluster name here
	args.Add("name", name)

	// (2) get filtered list of networks
	networkList, err := docker.NetworkList(ctx, types.NetworkListOptions{
		Filters: args,
	})
	if err != nil {
		log.Errorln("Failed to list docker networks")
		return "", false, err
	}

	// (2.1) If possible, return an existing network
	if len(networkList) > 1 {
		log.Warnf("Found %d networks instead of only one. Choosing the first one: '%s'.", len(networkList), networkList[0].ID)
	}

	if len(networkList) > 0 {
		log.Infof("Network with name '%s' already exists with ID '%s'", name, networkList[0].ID)
		return networkList[0].ID, true, nil
	}

	// (3) Create a new network
	network, err := docker.NetworkCreate(ctx, name, types.NetworkCreate{
		Labels: k3d.DefaultObjectLabels,
	})
	if err != nil {
		log.Errorln("Failed to create network")
		return "", false, err
	}

	log.Infof("Created network '%s'", name)
	return network.ID, false, nil
}

// DeleteNetwork deletes a network
func (d Docker) DeleteNetwork(ID string) error {
	// (0) create new docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Errorln("Failed to create docker client")
		return err
	}

	// (3) delete network
	return docker.NetworkRemove(ctx, ID)
}
