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
	"io"
	"io/ioutil"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// CreateNode creates a new container
func (d Docker) CreateNode(node *k3d.Node) error {
	log.Debugln("docker.CreateNode...")

	// translate node spec to docker container specs
	dockerNode, err := TranslateNodeToContainer(node)
	if err != nil {
		log.Errorln("Failed to translate k3d node specification to docker container specifications")
		return err
	}

	// create node
	if err := createContainer(dockerNode, node.Name); err != nil {
		log.Errorln("Failed to create k3d node")
		return err
	}

	return nil
}

// DeleteNode deletes a node
func (d Docker) DeleteNode(nodeSpec *k3d.Node) error {
	log.Debugln("docker.DeleteNode...")
	return removeContainer(nodeSpec.Name)
}

// GetNodesByLabel returns a list of existing nodes
func (d Docker) GetNodesByLabel(labels map[string]string) ([]*k3d.Node, error) {

	// (0) get containers
	containers, err := getContainersByLabel(labels)
	if err != nil {
		return nil, err
	}

	// (1) convert them to node structs
	nodes := []*k3d.Node{}
	for _, container := range containers {
		node, err := TranslateContainerToNode(&container)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil

}

// StartNode starts an existing node
func (d Docker) StartNode(node *k3d.Node) error {
	// (0) create docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("Failed to create docker client. %+v", err)
	}

	// get container which represents the node
	nodeContainer, err := getNodeContainer(node)
	if err != nil {
		log.Errorf("Failed to get container for node '%s'", node.Name)
		return err
	}

	// check if the container is actually managed by
	if v, ok := nodeContainer.Labels["app"]; !ok || v != "k3d" {
		return fmt.Errorf("Failed to determine if container '%s' is managed by k3d (needs label 'app=k3d')", nodeContainer.ID)
	}

	// actually start the container
	if err := docker.ContainerStart(ctx, nodeContainer.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}

// StopNode stops an existing node
func (d Docker) StopNode(node *k3d.Node) error {
	// (0) create docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("Failed to create docker client. %+v", err)
	}

	// get container which represents the node
	nodeContainer, err := getNodeContainer(node)
	if err != nil {
		log.Errorf("Failed to get container for node '%s'", node.Name)
		return err
	}

	// check if the container is actually managed by
	if v, ok := nodeContainer.Labels["app"]; !ok || v != "k3d" {
		return fmt.Errorf("Failed to determine if container '%s' is managed by k3d (needs label 'app=k3d')", nodeContainer.ID)
	}

	// actually stop the container
	if err := docker.ContainerStop(ctx, nodeContainer.ID, nil); err != nil {
		return err
	}

	return nil
}

func getContainersByLabel(labels map[string]string) ([]types.Container, error) {
	// (0) create docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("Failed to create docker client. %+v", err)
	}

	// (1) list containers which have the default k3d labels attached
	filters := filters.NewArgs()
	for k, v := range k3d.DefaultObjectLabels {
		filters.Add("label", fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range labels {
		filters.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	containers, err := docker.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters,
		All:     true,
	})
	if err != nil {
		log.Errorln("Failed to list containers")
		return nil, err
	}
	return containers, nil
}

// GetNode tries to get a node container by its name
func (d Docker) GetNode(node *k3d.Node) (*k3d.Node, error) {
	container, err := getNodeContainer(node)
	if err != nil {
		log.Errorf("Failed to get container for node '%s'", node.Name)
		return nil, err
	}

	node, err = TranslateContainerToNode(container)
	if err != nil {
		log.Errorf("Failed to translate container for node '%s' to node object", node.Name)
		return nil, err
	}

	return node, nil

}

// GetNodeLogs returns the logs from a given node
func (d Docker) GetNodeLogs(node *k3d.Node) (io.ReadCloser, error) {
	// get the container for the given node
	container, err := getNodeContainer(node)
	if err != nil {
		return nil, err
	}

	// create docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Errorln("Failed to create docker client")
		return nil, err
	}

	logreader, err := docker.ContainerLogs(ctx, container.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Errorf("Failed to get logs from node '%s' (container '%s')", node.Name, container.ID)
		return nil, err
	}

	return logreader, nil
}

// ExecInNode execs a command inside a node
func (d Docker) ExecInNode(node *k3d.Node, cmd []string) error {

	log.Debugf("Exec cmds '%+v' in node '%s'", cmd, node.Name)

	// get the container for the given node
	container, err := getNodeContainer(node)
	if err != nil {
		return err
	}

	// create docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Errorln("Failed to create docker client")
		return err
	}

	// exec
	exec, err := docker.ContainerExecCreate(ctx, container.ID, types.ExecConfig{
		Privileged:   true,
		Tty:          true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          cmd,
	})
	if err != nil {
		log.Errorf("Failed to create exec config for node '%s'", node.Name)
		return err
	}

	execConnection, err := docker.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		log.Errorf("Failed to connect to exec process in node '%s'", node.Name)
		return err
	}
	defer execConnection.Close()

	if err := docker.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{Tty: true}); err != nil {
		log.Errorf("Failed to start exec process in node '%s'", node.Name)
		return err
	}

	for {
		// get info about exec process inside container
		execInfo, err := docker.ContainerExecInspect(ctx, exec.ID)
		if err != nil {
			log.Errorln("Failed to inspect exec process in node '%s'", node.Name)
			return err
		}

		// if still running, continue loop
		if execInfo.Running {
			log.Debugf("Exec process '%+v' still running in node '%s'.. sleeping for 1 second...", cmd, node.Name)
			time.Sleep(1 * time.Second)
			continue
		}

		// check exitcode
		if execInfo.ExitCode == 0 { // success
			log.Debugf("Exec process in node '%s' exited with '0'", node.Name)
			break
		}

		if execInfo.ExitCode != 0 { // failed

			logs, err := ioutil.ReadAll(execConnection.Reader)
			if err != nil {
				log.Errorln("Failed to get logs from node '%s'", node.Name)
				return err
			}

			return fmt.Errorf("Logs from failed access process:\n%s", string(logs))
		}

	}

	return nil
}
