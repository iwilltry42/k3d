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

package containerd

import (
	"context"
	"io"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/containers"
	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// CreateNode creates a new k3d node
func (d Containerd) CreateNode(node *k3d.Node) error {
	log.Debugln("containerd.CreateNode...")

	// create containerd client
	ctx := context.Background()
	clientOpts := []containerd.ClientOpt{
		containerd.WithDefaultNamespace("k3d"),
	}
	client, err := containerd.New("/run/containerd/containerd.sock", clientOpts...) // TODO: this is the default address on UNIX, different on Windows
	if err != nil {
		log.Errorln("Failed to create containerd client")
		return err
	}

	// create container
	newContainerOpts := []containerd.NewContainerOpts{
		func(ctx context.Context, _ *containerd.Client, c *containers.Container) error {
			c.Image = node.Image
			c.Labels = node.Labels
			return nil
		},
	}
	container, err := client.NewContainer(ctx, node.Name, newContainerOpts...)
	if err != nil {
		log.Errorln("Couldn't create container")
		return err
	}

	/*
		// start container
		task, err := container.NewTask(ctx, cio.NewCreator()) // TODO: how the hell does this work?
		if err != nil {
			log.Errorln("Failed to create task in container", container.ID)
			return err
		}

		task.Start(ctx)
	*/

	log.Infoln("Created container with ID", container.ID())
	return nil
}

// DeleteNode deletes an existing k3d node
func (d Containerd) DeleteNode(node *k3d.Node) error {
	log.Debugln("containerd.DeleteNode...")
	ctx := context.Background()
	clientOpts := []containerd.ClientOpt{
		containerd.WithDefaultNamespace("k3d"),
	}
	client, err := containerd.New("/run/containerd/containerd.sock", clientOpts...) // TODO: this is the default address on UNIX, different on Windows
	if err != nil {
		log.Errorln("Failed to create containerd client")
		return err
	}

	container, err := client.LoadContainer(ctx, node.Name)
	if err != nil {
		log.Errorln("Couldn't load container", node.Name)
		return err
	}
	if err = container.Delete(ctx, []containerd.DeleteOpts{}...); err != nil {
		log.Errorln("Failed to delete container", container.ID)
		return err
	}

	return nil
}

// StartNode starts an existing node
func (d Containerd) StartNode(node *k3d.Node) error {
	return nil // TODO: fill
}

// StopNode stops an existing node
func (d Containerd) StopNode(node *k3d.Node) error {
	return nil // TODO: fill
}

func (d Containerd) GetNodesByLabel(labels map[string]string) ([]*k3d.Node, error) {
	return nil, nil
}

// GetNode tries to get a node container by its name
func (d Containerd) GetNode(node *k3d.Node) (*k3d.Node, error) {
	return nil, nil
}

// GetNodeLogs returns the logs from a given node
func (d Containerd) GetNodeLogs(node *k3d.Node) (io.ReadCloser, error) {
	return nil, nil
}

// ExecInNode execs a command inside a node
func (d Containerd) ExecInNode(node *k3d.Node, cmd []string) error {
	return nil
}
