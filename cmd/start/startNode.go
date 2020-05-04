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
package start

import (
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// NewCmdStartNode returns a new cobra command
func NewCmdStartNode() *cobra.Command {

	// create new command
	cmd := &cobra.Command{
		Use:   "node NAME", // TODO: startNode: allow one or more names or --all
		Short: "Start an existing k3d node",
		Long:  `Start an existing k3d node.`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugln("start node called")
			node := parseStartNodeCmd(cmd, args)
			if err := runtimes.SelectedRuntime.StartNode(node); err != nil {
				log.Fatalln(err)
			}
		},
	}

	// done
	return cmd
}

// parseStartNodeCmd parses the command input into variables required to start a node
func parseStartNodeCmd(cmd *cobra.Command, args []string) *k3d.Node {
	// node name // TODO: startNode: allow node filters, e.g. `k3d start nodes mycluster@worker` to start all worker nodes of cluster 'mycluster'
	if len(args) == 0 || len(args[0]) == 0 {
		log.Fatalln("No node name given")
	}

	return &k3d.Node{Name: args[0]}
}
