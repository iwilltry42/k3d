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
package delete

import (
	"github.com/rancher/k3d/pkg/cluster"
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCmdDeleteNode returns a new cobra command
func NewCmdDeleteNode() *cobra.Command {

	// create new cobra command
	cmd := &cobra.Command{
		Use:   "node (NAME | --all)",
		Short: "Delete a node.",
		Long:  `Delete a node.`,
		Args:  cobra.MinimumNArgs(1), // at least one node has to be specified
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugln("delete node called")

			runtime, nodes := parseDeleteNodeCmd(cmd, args)

			if len(nodes) == 0 {
				log.Infoln("No nodes found")
			} else {
				for _, node := range nodes {
					if err := cluster.DeleteNode(runtime, node); err != nil {
						log.Fatalln(err)
					}
				}
			}

			log.Debugln("...Finished")
		},
	}

	// add subcommands

	// add flags
	cmd.Flags().BoolP("all", "a", false, "Delete all existing clusters")

	// done
	return cmd
}

// parseDeleteNodeCmd parses the command input into variables required to delete nodes
func parseDeleteNodeCmd(cmd *cobra.Command, args []string) (runtimes.Runtime, []*k3d.Node) {
	// --runtime
	rt, err := cmd.Flags().GetString("runtime")
	if err != nil {
		log.Fatalln("No runtime specified")
	}
	runtime, err := runtimes.GetRuntime(rt)
	if err != nil {
		log.Fatalln(err)
	}

	// --all
	var nodes []*k3d.Node

	if all, err := cmd.Flags().GetBool("all"); err != nil {
		log.Fatalln(err)
	} else if all {
		nodes, err = cluster.GetNodes(runtime)
		if err != nil {
			log.Fatalln(err)
		}
		return runtime, nodes
	}

	if len(args) < 1 {
		log.Fatalln("Expecting at least one node name if `--all` is not set")
	}

	for _, name := range args {
		node, err := cluster.GetNode(&k3d.Node{Name: name}, runtime)
		if err != nil {
			log.Fatalln(err)
		}
		nodes = append(nodes, node)
	}

	return runtime, nodes
}
