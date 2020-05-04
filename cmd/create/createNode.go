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
package create

import (
	"fmt"

	"github.com/spf13/cobra"

	k3dc "github.com/rancher/k3d/pkg/cluster"
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	"github.com/rancher/k3d/version"
	log "github.com/sirupsen/logrus"
)

// NewCmdCreateNode returns a new cobra command
func NewCmdCreateNode() *cobra.Command {

	// create new command
	cmd := &cobra.Command{
		Use:   "node NAME",
		Short: "Create a new k3s node in docker",
		Long:  `Create a new containerized k3s node (k3s in docker).`,
		Args:  cobra.ExactArgs(1), // exactly one name accepted // TODO: if not specified, inherit from cluster that the node shall belong to, if that is specified
		Run: func(cmd *cobra.Command, args []string) {
			nodes, cluster := parseCreateNodeCmd(cmd, args)
			for _, node := range nodes {
				if err := k3dc.AddNodeToCluster(runtimes.SelectedRuntime, node, cluster); err != nil {
					log.Errorf("Failed to add node '%s' to cluster '%s'", node.Name, cluster.Name)
					log.Errorln(err)
				}
			}
		},
	}

	// add flags
	cmd.Flags().Int("replicas", 1, "Number of replicas of this node specification.")
	cmd.Flags().String("role", string(k3d.WorkerRole), "Specify node role [master, worker]")
	cmd.Flags().StringP("cluster", "c", k3d.DefaultClusterName, "Select the cluster that the node shall connect to.")
	if err := cmd.MarkFlagRequired("cluster"); err != nil {
		log.Fatalln("Failed to mark required flag '--cluster'")
	}

	cmd.Flags().StringP("image", "i", fmt.Sprintf("%s:%s", k3d.DefaultK3sImageRepo, version.GetK3sVersion(false)), "Specify k3s image used for the node(s)")

	// done
	return cmd
}

// parseCreateNodeCmd parses the command input into variables required to create a cluster
func parseCreateNodeCmd(cmd *cobra.Command, args []string) ([]*k3d.Node, *k3d.Cluster) {

	// --replicas
	replicas, err := cmd.Flags().GetInt("replicas")
	if err != nil {
		log.Errorln("No replica count specified")
		log.Fatalln(err)
	}

	// --role
	// TODO: createNode: for --role=master, update the nginx config and add TLS-SAN and server connection, etc.
	roleStr, err := cmd.Flags().GetString("role")
	if err != nil {
		log.Errorln("No node role specified")
		log.Fatalln(err)
	}
	if _, ok := k3d.NodeRoles[roleStr]; !ok {
		log.Fatalf("Unknown node role '%s'\n", roleStr)
	}
	role := k3d.NodeRoles[roleStr]

	// --image
	image, err := cmd.Flags().GetString("image")
	if err != nil {
		log.Errorln("No image specified")
		log.Fatalln(err)
	}

	// --cluster
	clusterName, err := cmd.Flags().GetString("cluster")
	if err != nil {
		log.Fatalln(err)
	}
	cluster := &k3d.Cluster{
		Name: clusterName,
	}

	// generate list of nodes
	nodes := []*k3d.Node{}
	for i := 0; i < replicas; i++ {
		node := &k3d.Node{
			Name:  fmt.Sprintf("%s-%s-%d", k3d.DefaultObjectNamePrefix, args[0], i),
			Role:  role,
			Image: image,
		}
		nodes = append(nodes, node)
	}

	return nodes, cluster
}
