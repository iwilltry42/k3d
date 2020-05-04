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

	"github.com/rancher/k3d/pkg/cluster"
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/liggitt/tabwriter"
)

// NewCmdGetCluster returns a new cobra command
func NewCmdGetCluster() *cobra.Command {

	// create new command
	cmd := &cobra.Command{
		Use:     "cluster [NAME [NAME...]]",
		Aliases: []string{"clusters"},
		Short:   "Get cluster",
		Long:    `Get cluster.`,
		Args:    cobra.MinimumNArgs(0), // 0 or more; 0 = all
		Run: func(cmd *cobra.Command, args []string) {
			log.Debugln("get cluster called")
			clusters, headersOff := parseGetClusterCmd(cmd, args)
			var existingClusters []*k3d.Cluster
			if clusters == nil { // Option a)  no cluster name specified -> get all clusters
				found, err := cluster.GetClusters(runtimes.SelectedRuntime)
				if err != nil {
					log.Fatalln(err)
				}
				existingClusters = append(existingClusters, found...)
			} else { // Option b) cluster name specified -> get specific cluster
				found, err := cluster.GetCluster(clusters, runtimes.SelectedRuntime)
				if err != nil {
					log.Fatalln(err)
				}
				existingClusters = append(existingClusters, found)
			}
			// print existing clusters
			printClusters(existingClusters, headersOff)
		},
	}

	// add flags
	cmd.Flags().Bool("no-headers", false, "Disable headers")

	// add subcommands

	// done
	return cmd
}

func parseGetClusterCmd(cmd *cobra.Command, args []string) (*k3d.Cluster, bool) {

	// --no-headers
	headersOff, err := cmd.Flags().GetBool("no-headers")
	if err != nil {
		log.Fatalln(err)
	}

	// Args = cluster name
	if len(args) == 0 {
		return nil, headersOff
	}

	cluster := &k3d.Cluster{Name: args[0]}

	return cluster, headersOff
}

func printClusters(clusters []*k3d.Cluster, headersOff bool) {

	tabwriter := tabwriter.NewWriter(os.Stdout, 6, 4, 3, ' ', tabwriter.RememberWidths)
	defer tabwriter.Flush()

	if !headersOff {
		headers := []string{"NAME", "MASTERS", "WORKERS"} // TODO: getCluster: add status column
		_, err := fmt.Fprintf(tabwriter, "%s\n", strings.Join(headers, "\t"))
		if err != nil {
			log.Fatalln("Failed to print headers")
		}
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})

	for _, cluster := range clusters {
		masterCount := 0
		workerCount := 0
		for _, node := range cluster.Nodes {
			if node.Role == k3d.MasterRole {
				masterCount++
			} else if node.Role == k3d.WorkerRole {
				workerCount++
			}
		}
		fmt.Fprintf(tabwriter, "%s\t%d\t%d\n", cluster.Name, masterCount, workerCount)
	}
}
