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
package stop

import (
	"github.com/rancher/k3d/pkg/runtimes"
	"github.com/spf13/cobra"

	k3d "github.com/rancher/k3d/pkg/types"

	log "github.com/sirupsen/logrus"
)

// NewCmdStopNode returns a new cobra command
func NewCmdStopNode() *cobra.Command {

	// create new command
	cmd := &cobra.Command{
		Use:   "node NAME", // TODO: allow one or more names or --all",
		Short: "Stop an existing k3d node",
		Long:  `Stop an existing k3d node.`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugln("stop node called")
			runtime, node := parseStopNodeCmd(cmd, args)
			if err := runtime.StopNode(node); err != nil {
				log.Fatalln(err)
			}
		},
	}

	// done
	return cmd
}

// parseStopNodeCmd parses the command input into variables required to stop a node
func parseStopNodeCmd(cmd *cobra.Command, args []string) (runtimes.Runtime, *k3d.Node) {
	// --runtime
	rt, err := cmd.Flags().GetString("runtime")
	if err != nil {
		log.Fatalln("No runtime specified")
	}
	runtime, err := runtimes.GetRuntime(rt)
	if err != nil {
		log.Fatalln(err)
	}

	// node name // TODO: allow node filters, e.g. `k3d stop nodes mycluster@worker` to stop all worker nodes of cluster 'mycluster'
	if len(args) == 0 || len(args[0]) == 0 {
		log.Fatalln("No node name given")
	}

	return runtime, &k3d.Node{Name: args[0]} // TODO: validate and allow for more than one
}
