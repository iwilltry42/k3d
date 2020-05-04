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
package get

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/liggitt/tabwriter"
	"github.com/rancher/k3d/pkg/cluster"
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// NewCmdGetNode returns a new cobra command
func NewCmdGetNode() *cobra.Command {

	// create new command
	cmd := &cobra.Command{
		Use:     "node NAME", // TODO: getNode: allow one or more names or --all flag
		Short:   "Get node",
		Aliases: []string{"nodes"},
		Long:    `Get node.`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugln("get node called")
			node, headersOff := parseGetNodeCmd(cmd, args)
			var existingNodes []*k3d.Node
			if node == nil { // Option a)  no name specified -> get all nodes
				found, err := cluster.GetNodes(runtimes.SelectedRuntime)
				if err != nil {
					log.Fatalln(err)
				}
				existingNodes = append(existingNodes, found...)
			} else { // Option b) cluster name specified -> get specific cluster
				found, err := cluster.GetNode(node, runtimes.SelectedRuntime)
				if err != nil {
					log.Fatalln(err)
				}
				existingNodes = append(existingNodes, found)
			}
			// print existing clusters
			printNodes(existingNodes, headersOff)
		},
	}

	// add flags
	cmd.Flags().Bool("no-headers", false, "Disable headers")

	// add subcommands

	// done
	return cmd
}

func parseGetNodeCmd(cmd *cobra.Command, args []string) (*k3d.Node, bool) {
	// --no-headers
	headersOff, err := cmd.Flags().GetBool("no-headers")
	if err != nil {
		log.Fatalln(err)
	}

	// Args = node name
	if len(args) == 0 {
		return nil, headersOff
	}

	node := &k3d.Node{Name: args[0]} // TODO: validate name first?

	return node, headersOff
}

func printNodes(nodes []*k3d.Node, headersOff bool) {

	tabwriter := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', tabwriter.RememberWidths)
	defer tabwriter.Flush()

	if !headersOff {
		headers := []string{"NAME", "ROLE", "CLUSTER"} // TODO: add status
		_, err := fmt.Fprintf(tabwriter, "%s\n", strings.Join(headers, "\t"))
		if err != nil {
			log.Fatalln("Failed to print headers")
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	for _, node := range nodes {
		fmt.Fprintf(tabwriter, "%s\t%s\t%s\n", strings.TrimPrefix(node.Name, "/"), string(node.Role), node.Labels["k3d.cluster"])
	}
}
