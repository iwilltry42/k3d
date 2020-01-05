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

	"github.com/rancher/k3d/pkg/cluster"
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// NewCmdGetKubeconfig returns a new cobra command
func NewCmdGetKubeconfig() *cobra.Command {

	// create new command
	cmd := &cobra.Command{
		Use:   "kubeconfig NAME", // TODO: enable putting more than one name or even --all
		Short: "Get kubeconfig",
		Long:  `Get kubeconfig.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugln("get kubeconfig called")
			rt, c, path := parseGetKubeconfigCmd(cmd, args)
			kubeconfigpath, err := cluster.GetKubeconfigPath(rt, c, path)
			if err != nil {
				log.Fatalln(err)
			}

			// only print kubeconfig file path if output is not stdout ("-")
			if path != "-" {
				fmt.Println(kubeconfigpath)
			}
		},
	}

	// add flags
	cmd.Flags().StringP("output", "o", "", "Define output [ - | <file> ]")
	if err := cmd.MarkFlagFilename("output"); err != nil {
		log.Fatalln("Failed to mark flag --output as filename")
	}
	// cmd.Flags().BoolP("all", "a", false, "Get kubeconfigs from all existing clusters") // TODO:

	// done
	return cmd
}

func parseGetKubeconfigCmd(cmd *cobra.Command, args []string) (runtimes.Runtime, *k3d.Cluster, string) {
	// --runtime
	rt, err := cmd.Flags().GetString("runtime")
	if err != nil {
		log.Fatalln("No runtime specified")
	}
	runtime, err := runtimes.GetRuntime(rt)
	if err != nil {
		log.Fatalln(err)
	}

	// --output
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		log.Fatalln("No output specified")
	}

	return runtime, &k3d.Cluster{Name: args[0]}, output // TODO: validate first
}
