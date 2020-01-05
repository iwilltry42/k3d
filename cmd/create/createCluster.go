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
	"os"

	"github.com/spf13/cobra"

	cliutil "github.com/rancher/k3d/cmd/util"
	k3dCluster "github.com/rancher/k3d/pkg/cluster"
	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	"github.com/rancher/k3d/version"

	log "github.com/sirupsen/logrus"
)

// NewCmdCreateCluster returns a new cobra command
func NewCmdCreateCluster() *cobra.Command {

	// create new command
	cmd := &cobra.Command{
		Use:   "cluster NAME",
		Short: "Create a new k3s cluster in docker",
		Long:  `Create a new k3s cluster with containerized nodes (k3s in docker).`,
		Args:  cobra.ExactArgs(1), // exactly one cluster name can be set // TODO: if not specified, use k3d.DefaultClusterName
		Run: func(cmd *cobra.Command, args []string) {
			runtime, cluster := parseCreateClusterCmd(cmd, args)
			if err := k3dCluster.CreateCluster(cluster, runtime); err != nil {
				log.Fatalln(err)
			}
			log.Infof("Cluster '%s' created successfully. You can now use it like this:", cluster.Name)
			fmt.Printf("export KUBECONFIG=$(%s get kubeconfig %s)\n", os.Args[0], cluster.Name)
			fmt.Println("kubectl cluster-info")
		},
	}

	/*********
	 * Flags *
	 *********/
	cmd.Flags().StringArrayP("api-port", "a", []string{"6443"}, "Specify the Kubernetes API server port (Format: `--api-port [HOST:]HOSTPORT[@NODEFILTER]`\n - Example: `k3d create -m 3 -a 0.0.0.0:6550@master[0] -a 0.0.0.0:6551@master[1]` ")
	cmd.Flags().IntP("masters", "m", 1, "Specify how many masters you want to create")
	cmd.Flags().IntP("workers", "w", 0, "Specify how many workers you want to create")
	// cmd.Flags().String("config", "", "Specify a cluster configuration file")                                     // TODO: to implement
	cmd.Flags().String("image", fmt.Sprintf("%s:%s", k3d.DefaultK3sImageRepo, version.GetK3sVersion(false)), "Specify k3s image that you want to use for the nodes")
	cmd.Flags().String("network", "", "Join an existing network")
	cmd.Flags().String("secret", "", "Specify a cluster secret. By default, we generate one.")
	cmd.Flags().StringArrayP("volume", "v", nil, "Mount volumes into the nodes (Format: `--volume [SOURCE:]DEST[@NODEFILTER[;NODEFILTER...]]`\n - Example: `k3d create -w 2 -v /my/path@worker[0,1] -v /tmp/test:/tmp/other@master[0]`")
	cmd.Flags().StringArrayP("port", "p", nil, "Map ports from the node containers to the host (Format: `[HOST:][HOSTPORT:]CONTAINERPORT[/PROTOCOL][@NODEFILTER]`)\n - Example: `k3d create -w 2 -p 8080:80@worker[0] -p 8081@worker[1]`")
	cmd.Flags().Int("wait", -1, "Wait for a specified amount of time (seconds, >= 0, 0 = forever) for the master(s) to be ready or timeout and rollback before returning")

	/* Image Importing */
	cmd.Flags().Bool("no-image-volume", false, "Disable the creation of a volume for importing images")

	/* Multi Master Configuration */ // TODO: to implement (whole multi master thingy)
	// multi-master - general
	cmd.Flags().Bool("no-lb", false, "[WIP] Disable automatic deployment of a load balancer in Multi-Master setups")                         // TODO: to implement
	cmd.Flags().String("lb-port", "0.0.0.0:6443", "[WIP] Specify port to be exposed by the master load balancer (Format: `[HOST:]HOSTPORT)") // TODO: to implement

	// multi-master - datastore
	cmd.Flags().String("datastore-endpoint", "", "[WIP] Specify external datastore endpoint (e.g. for multi master clusters)")
	/* TODO: activate
	cmd.Flags().String("datastore-network", "", "Specify container network where we can find the datastore-endpoint (add a connection)")

	// TODO: set default paths and hint, that one should simply mount the files using --volume flag
	cmd.Flags().String("datastore-cafile", "", "Specify external datastore's TLS Certificate Authority (CA) file")
	cmd.Flags().String("datastore-certfile", "", "Specify external datastore's TLS certificate file'")
	cmd.Flags().String("datastore-keyfile", "", "Specify external datastore's TLS key file'")
	*/

	/* k3s */ // TODO: to implement extra args
	cmd.Flags().StringArray("k3s-server-arg", nil, "[WIP] Additional args passed to the `k3s server` command on master nodes")
	cmd.Flags().StringArray("k3s-agent-arg", nil, "[WIP] Additional args passed to the `k3s agent` command on worker nodes")

	/* Subcommands */

	// done
	return cmd
}

// parseCreateClusterCmd parses the command input into variables required to create a cluster
func parseCreateClusterCmd(cmd *cobra.Command, args []string) (runtimes.Runtime, *k3d.Cluster) {
	// --runtime
	rt, err := cmd.Flags().GetString("runtime")
	if err != nil {
		log.Fatalln("No runtime specified")
	}
	runtime, err := runtimes.GetRuntime(rt)
	if err != nil {
		log.Fatalln(err)
	}

	// --image
	image, err := cmd.Flags().GetString("image")
	if err != nil {
		log.Errorln("No image specified")
		log.Fatalln(err)
	}
	if image == "latest" {
		image = version.GetK3sVersion(true)
	}

	// --masters
	masterCount, err := cmd.Flags().GetInt("masters")
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: allow more than one master
	if masterCount > 1 {
		log.Warnln("Multi-Master is setup not fully implemented/supported right now!")
	}

	// --workers
	workerCount, err := cmd.Flags().GetInt("workers")
	if err != nil {
		log.Fatalln(err)
	}

	// --network
	networkName, err := cmd.Flags().GetString("network")
	if err != nil {
		log.Fatalln(err)
	}
	network := k3d.ClusterNetwork{}
	if networkName != "" {
		network.Name = networkName
		network.External = true
	}

	// --secret
	secret, err := cmd.Flags().GetString("secret")
	if err != nil {
		log.Fatalln(err)
	}

	// --wait
	wait, err := cmd.Flags().GetInt("wait")
	if err != nil {
		log.Fatalln(err)
	}
	if cmd.Flags().Changed("wait") && wait < 0 {
		log.Fatalln("Value of '--wait' can't be less than 0")
	}
	if cmd.Flags().Changed("wait") {
		log.Warnln("--wait not implemented yet") // TODO:
	}

	// --api-port
	apiPortFlags, err := cmd.Flags().GetStringArray("api-port")
	if err != nil {
		log.Fatalln(err)
	}

	// error out if we have more api-ports than masters specified
	if len(apiPortFlags) > masterCount {
		log.Fatalf("Cannot expose more api-ports than master nodes exist (%d > %d)", len(apiPortFlags), masterCount)
	}

	ipPortCombinations := map[string]struct{}{} // only for finding duplicates
	apiPortFilters := map[string]struct{}{}     // only for deduplication
	exposeAPIToFiltersMap := map[k3d.ExposeAPI][]string{}
	for _, apiPortFlag := range apiPortFlags {

		// split the flag value from the node filter
		apiPortString, filters, err := cliutil.SplitFiltersFromFlag(apiPortFlag)
		if err != nil {
			log.Fatalln(err)
		}

		// if there's only one master node, we don't need a node filter, but if there's more than one, we need exactly one node filter per api-port flag
		if len(filters) > 1 || (len(filters) == 0 && masterCount > 1) {
			log.Fatalf("Exactly one node filter required per '--api-port' flag, but got %d on flag %s", len(filters), apiPortFlag)
		}

		// add default, if no filter was set and we only have a single master node
		if len(filters) == 0 && masterCount == 1 {
			filters = []string{"master[0]"}
		}

		// only one api-port mapping allowed per master node
		if _, exists := apiPortFilters[filters[0]]; exists {
			log.Fatalf("Cannot assign multiple api-port mappings to the same node: duplicate '%s'", filters[0])
		}
		apiPortFilters[filters[0]] = struct{}{}

		// parse the port mapping
		exposeAPI, err := cliutil.ParseAPIPort(apiPortString)
		if err != nil {
			log.Fatalln(err)
		}

		// error out on duplicates
		ipPort := fmt.Sprintf("%s:%s", exposeAPI.HostIP, exposeAPI.Port)
		if _, exists := ipPortCombinations[ipPort]; exists {
			log.Fatalf("Duplicate IP:PORT combination '%s' for the Api Port is not allowed", ipPort)
		}
		ipPortCombinations[ipPort] = struct{}{}

		// add to map
		exposeAPIToFiltersMap[exposeAPI] = filters
	}

	// --no-lb
	noLB, err := cmd.Flags().GetBool("no-lb")
	if err != nil {
		log.Fatalln(err)
	}

	// --lb-port
	lbPort, err := cmd.Flags().GetString("lb-port")
	if err != nil {
		log.Fatalln(err)
	}

	// --datastore-endpoint
	datastoreEndpoint, err := cmd.Flags().GetString("datastore-endpoint")
	if err != nil {
		log.Fatalln(err)
	}
	if datastoreEndpoint != "" {
		log.Fatalln("Using an external datastore for HA clusters is not yet supported.")
	}

	// --volume
	volumeFlags, err := cmd.Flags().GetStringArray("volume")
	if err != nil {
		log.Fatalln(err)
	}

	// volumeFilterMap will map volume mounts to applied node filters
	volumeFilterMap := make(map[string][]string, 1)
	for _, volumeFlag := range volumeFlags {

		// split node filter from the specified volume
		volume, filters, err := cliutil.SplitFiltersFromFlag(volumeFlag)
		if err != nil {
			log.Fatalln(err)
		}

		// validate the specified volume mount and return it in SRC:DEST format
		volume, err = cliutil.ValidateVolumeMount(volume)
		if err != nil {
			log.Fatalln(err)
		}

		// create new entry or append filter to existing entry
		if _, exists := volumeFilterMap[volume]; exists {
			volumeFilterMap[volume] = append(volumeFilterMap[volume], filters...)
		} else {
			volumeFilterMap[volume] = filters
		}
	}

	// --port
	portFlags, err := cmd.Flags().GetStringArray("port")
	if err != nil {
		log.Fatalln(err)
	}
	portFilterMap := make(map[string][]string, 1)
	for _, portFlag := range portFlags {
		// split node filter from the specified volume
		portmap, filters, err := cliutil.SplitFiltersFromFlag(portFlag)
		if err != nil {
			log.Fatalln(err)
		}

		if len(filters) > 1 {
			log.Fatalln("Can only apply a Portmap to one node")
		}

		// the same portmapping can't be applied to multiple nodes

		// validate the specified volume mount and return it in SRC:DEST format
		portmap, err = cliutil.ValidatePortMap(portmap)
		if err != nil {
			log.Fatalln(err)
		}

		// create new entry or append filter to existing entry
		if _, exists := portFilterMap[portmap]; exists {
			log.Fatalln("Same Portmapping can not be used for multiple nodes")
		} else {
			portFilterMap[portmap] = filters
		}
	}

	// --no-image-volume
	noImageVolume, err := cmd.Flags().GetBool("no-image-volume")
	if err != nil {
		log.Fatalln(err)
	}

	/*******************
	* generate cluster *
	********************/
	cluster := &k3d.Cluster{
		Name:    args[0], // TODO: validate name0
		Network: network,
		Secret:  secret,
		ClusterCreationOpts: &k3d.ClusterCreationOpts{
			DisableImageVolume: noImageVolume,
			WaitForMaster:      wait,
		},
	}

	// generate list of nodes
	cluster.Nodes = []*k3d.Node{}

	// -> master nodes
	for i := 0; i < masterCount; i++ {
		node := k3d.Node{
			Role:       k3d.MasterRole,
			Image:      image,
			MasterOpts: k3d.MasterOpts{},
		}

		// TODO: by default, we don't expose an PI port, even if we only have a single master: should we change that?
		// -> if we want to change that, simply add the exposeAPI struct here

		// first master node will be init node if we have more than one master specified but no external datastore
		if i == 0 && masterCount > 1 && datastoreEndpoint == "" {
			node.MasterOpts.IsInit = true
			cluster.InitNode = &node
		} // TODO: enable external datastore as well

		// append node to list
		cluster.Nodes = append(cluster.Nodes, &node)
	}

	// -> worker nodes
	for i := 0; i < workerCount; i++ {
		node := k3d.Node{
			Role:  k3d.WorkerRole,
			Image: image,
		}

		cluster.Nodes = append(cluster.Nodes, &node)
	}

	// add masterOpts
	for exposeAPI, filters := range exposeAPIToFiltersMap {
		nodes, err := cliutil.FilterNodes(cluster.Nodes, filters)
		if err != nil {
			log.Fatalln(err)
		}
		for _, node := range nodes {
			if node.Role != k3d.MasterRole {
				log.Fatalf("Node returned by filters '%+v' for exposing the API is not a master node", filters)
			}
			node.MasterOpts.ExposeAPI = exposeAPI
		}
	}

	// append volumes
	for volume, filters := range volumeFilterMap {
		nodes, err := cliutil.FilterNodes(cluster.Nodes, filters)
		if err != nil {
			log.Fatalln(err)
		}
		for _, node := range nodes {
			node.Volumes = append(node.Volumes, volume)
		}
	}

	// append ports
	for portmap, filters := range portFilterMap {
		nodes, err := cliutil.FilterNodes(cluster.Nodes, filters)
		if err != nil {
			log.Fatalln(err)
		}
		for _, node := range nodes {
			node.Ports = append(node.Ports, portmap)
		}
	}

	// TODO: create load balancer and other util containers // TODO: for now, this will only work with the docker provider (?) -> can replace dynamic docker lookup with static traefik config (?)
	if masterCount > 1 && !noLB { // TODO: add traefik to the same network and add traefik labels to the master node containers
		log.Debugln("Creating LB in front of master nodes")
		cluster.MasterLoadBalancer = &k3d.ClusterLoadbalancer{
			Image:       k3d.DefaultLBImage,
			ExposedPort: lbPort,
		}
	}

	return runtime, cluster
}
