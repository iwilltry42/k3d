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

package cluster

import (
	"fmt"
	"strings"

	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// AddNodeToCluster adds a node to an existing cluster
func AddNodeToCluster(runtime runtimes.Runtime, node *k3d.Node, cluster *k3d.Cluster) error {
	clusterName := cluster.Name
	cluster, err := GetCluster(cluster, runtime)
	if err != nil {
		log.Errorf("Failed to find specified cluster '%s'", clusterName)
		return err
	}

	log.Debugf("Adding node to cluster %+v", cluster)

	// network
	node.Network = cluster.Network.Name

	// skeleton
	node.Labels = map[string]string{}
	node.Env = []string{}

	// copy labels and env vars from a similar node in the selected cluster
	for _, existingNode := range cluster.Nodes {
		if existingNode.Role == node.Role {

			log.Debugf("Copying configuration from existing node %+v", existingNode)

			for k, v := range existingNode.Labels {
				if strings.HasPrefix(k, "k3d") {
					node.Labels[k] = v
				}
				if k == "k3d.cluster.url" {
					node.Env = append(node.Env, fmt.Sprintf("K3S_URL=%s", v))
				}
				if k == "k3d.cluster.secret" {
					node.Env = append(node.Env, fmt.Sprintf("K3S_TOKEN=%s", v))
				}
			}

			for _, env := range existingNode.Env {
				if strings.HasPrefix(env, "K3S_") {
					node.Env = append(node.Env, env)
				}
			}
			break
		}
	}

	log.Debugf("Resulting node %+v", node)

	return CreateNode(node, runtime)
}

// CreateNodes creates a list of nodes
func CreateNodes(nodes []*k3d.Node, runtime runtimes.Runtime) { // TODO: pass `--atomic` flag, so we stop and return an error if any node creation fails?
	for _, node := range nodes {
		if err := CreateNode(node, runtime); err != nil {
			log.Error(err)
		}
	}
}

// CreateNode creates a new containerized k3s node
func CreateNode(node *k3d.Node, runtime runtimes.Runtime) error {
	log.Debugf("Creating node from spec\n%+v", node)

	/*
	 * CONFIGURATION
	 */

	/* global node configuration (applies for any node role) */

	// ### Labels ###
	labels := make(map[string]string)
	for k, v := range k3d.DefaultObjectLabels {
		labels[k] = v
	}
	for k, v := range node.Labels {
		labels[k] = v
	}
	node.Labels = labels

	// ### Environment ###
	node.Env = append(node.Env, k3d.DefaultNodeEnv...) // append default node env vars

	// specify options depending on node role
	if node.Role == k3d.WorkerRole { // TODO: check here AND in CLI or only here?
		if err := patchWorkerSpec(node); err != nil {
			return err
		}
	} else if node.Role == k3d.MasterRole {
		if err := patchMasterSpec(node); err != nil {
			return err
		}
		log.Debugf("spec = %+v\n", node)
	} else {
		return fmt.Errorf("Unknown node role '%s'", node.Role)
	}

	/*
	 * CREATION
	 */
	if err := runtime.CreateNode(node); err != nil {
		return err
	}

	return nil
}

// DeleteNode deletes an existing node
func DeleteNode(runtime runtimes.Runtime, node *k3d.Node) error {

	if err := runtime.DeleteNode(node); err != nil {
		log.Error(err)
	}
	return nil
}

// patchWorkerSpec adds worker node specific settings to a node
func patchWorkerSpec(node *k3d.Node) error {
	node.Args = append([]string{"agent"}, node.Args...)
	node.Labels["k3d.role"] = string(k3d.WorkerRole) // TODO: maybe put those in a global var DefaultWorkerNodeSpec?
	return nil
}

// patchMasterSpec adds worker node specific settings to a node
func patchMasterSpec(node *k3d.Node) error {

	// command / arguments
	node.Args = append([]string{"server"}, node.Args...)

	// role label
	node.Labels["k3d.role"] = string(k3d.MasterRole) // TODO: maybe put those in a global var DefaultMasterNodeSpec?

	// extra settings to expose the API port (if wanted)
	if node.MasterOpts.ExposeAPI.Port != "" {
		if node.MasterOpts.ExposeAPI.Host == "" {
			node.MasterOpts.ExposeAPI.Host = "0.0.0.0"
		}
		node.Labels["k3d.master.api.hostIP"] = node.MasterOpts.ExposeAPI.HostIP // TODO: maybe get docker machine IP here

		node.Labels["k3d.master.api.host"] = node.MasterOpts.ExposeAPI.Host

		node.Args = append(node.Args, "--tls-san", node.MasterOpts.ExposeAPI.Host) // add TLS SAN for non default host name
		node.Labels["k3d.master.api.port"] = node.MasterOpts.ExposeAPI.Port
		node.Ports = append(node.Ports, fmt.Sprintf("%s:%s:6443/tcp", node.MasterOpts.ExposeAPI.Host, node.MasterOpts.ExposeAPI.Port)) // TODO: get '6443' from defaultport variable
	}
	return nil
}

// GetNodes returns a list of all existing clusters
func GetNodes(runtime runtimes.Runtime) ([]*k3d.Node, error) {
	nodes, err := runtime.GetNodesByLabel(k3d.DefaultObjectLabels)
	if err != nil {
		log.Errorln("Failed to get nodes")
		return nil, err
	}

	return nodes, nil
}

// GetNode returns an existing cluster
func GetNode(node *k3d.Node, runtime runtimes.Runtime) (*k3d.Node, error) {
	// get node
	node, err := runtime.GetNode(node)
	if err != nil {
		log.Errorf("Failed to get node '%s'", node.Name)
	}

	return node, nil
}
