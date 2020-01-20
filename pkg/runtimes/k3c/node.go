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

package k3c

import (
	"context"
	"io"

	"github.com/rancher/k3c/pkg/client"
	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// CreateNode creates a new k3d node
func (d K3c) CreateNode(node *k3d.Node) error {
	log.Debugln("k3c.CreateNode...")

	ctx := context.Background()

	k3cclient, err := client.New(ctx, "")

	log.Printf("%+v", k3cclient)

	return nil
}

// DeleteNode deletes an existing k3d node
func (d K3c) DeleteNode(node *k3d.Node) error {
	log.Debugln("k3c.DeleteNode...")
	
	return nil
}

// StartNode starts an existing node
func (d K3c) StartNode(node *k3d.Node) error {
	return nil // TODO: fill
}

// StopNode stops an existing node
func (d K3c) StopNode(node *k3d.Node) error {
	return nil // TODO: fill
}

func (d K3c) GetNodesByLabel(labels map[string]string) ([]*k3d.Node, error) {
	return nil, nil
}

// GetNode tries to get a node container by its name
func (d K3c) GetNode(node *k3d.Node) (*k3d.Node, error) {
	return nil, nil
}

// GetNodeLogs returns the logs from a given node
func (d K3c) GetNodeLogs(node *k3d.Node) (io.ReadCloser, error) {
	return nil, nil
}

// ExecInNode execs a command inside a node
func (d K3c) ExecInNode(node *k3d.Node, cmd []string) error {
	return nil
}
