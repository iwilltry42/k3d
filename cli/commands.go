package run

/*
 * This file contains the "backend" functionality for the CLI commands (and flags)
 */

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/urfave/cli"
)

const (
	defaultRegistry    = "docker.io"
	defaultServerCount = 1
)

// CheckTools checks if the docker API server is responding
func CheckTools(c *cli.Context) error {
	log.Print("Checking docker...")
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	ping, err := docker.Ping(ctx)

	if err != nil {
		return fmt.Errorf("ERROR: checking docker failed\n%+v", err)
	}
	log.Printf("SUCCESS: Checking docker succeeded (API: v%s)\n", ping.APIVersion)
	return nil
}

// CreateCluster creates a new single-node cluster container and initializes the cluster directory
func CreateCluster(c *cli.Context) error {

	if err := CheckClusterName(c.String("name")); err != nil {
		return err
	}

	if cluster, err := getClusters(false, c.String("name")); err != nil {
		return err
	} else if len(cluster) != 0 {
		// A cluster exists with the same name. Return with an error.
		return fmt.Errorf("ERROR: Cluster %s already exists", c.String("name"))
	}

	// On Error delete the cluster.  If there createCluster() encounter any error,
	// call this function to remove all resources allocated for the cluster so far
	// so that they don't linger around.
	deleteCluster := func() {
		if err := DeleteCluster(c); err != nil {
			log.Printf("Error: Failed to delete cluster %s", c.String("name"))
		}
	}

	// define image
	image := c.String("image")
	if c.IsSet("version") {
		// TODO: --version to be deprecated
		log.Println("[WARNING] The `--version` flag will be deprecated soon, please use `--image rancher/k3s:<version>` instead")
		if c.IsSet("image") {
			// version specified, custom image = error (to push deprecation of version flag)
			log.Fatalln("[ERROR] Please use `--image <image>:<version>` instead of --image and --version")
		} else {
			// version specified, default image = ok (until deprecation of version flag)
			image = fmt.Sprintf("%s:%s", strings.Split(image, ":")[0], c.String("version"))
		}
	}
	if len(strings.Split(image, "/")) <= 2 {
		// fallback to default registry
		image = fmt.Sprintf("%s/%s", defaultRegistry, image)
	}

	// create cluster network
	networkID, err := createClusterNetwork(c.String("name"))
	if err != nil {
		return err
	}
	log.Printf("Created cluster network with ID %s", networkID)

	// environment variables
	env := []string{"K3S_KUBECONFIG_OUTPUT=/output/kubeconfig.yaml"}
	env = append(env, c.StringSlice("env")...)
	if !c.Bool("agent-only") {
		env = append(env, fmt.Sprintf("K3S_CLUSTER_SECRET=%s", GenerateRandomString(20)))
	} else {
		env = append(env, fmt.Sprintf("K3S_CLUSTER_SECRET=%s", c.String("secret")), fmt.Sprintf("K3S_URL=https://%s", c.String("server")))
	}

	// k3s server arguments
	// TODO: --port will soon be --api-port since we want to re-use --port for arbitrary port mappings
	if c.IsSet("port") {
		log.Println("INFO: As of v2.0.0 --port will be used for arbitrary port mapping. Please use --api-port/-a instead for configuring the Api Port")
	}
	apiPort, err := parseAPIPort(c.String("api-port"))
	if err != nil {
		return err
	}

	k3sServerArgs := []string{"--https-listen-port", apiPort.Port}

	// When the 'host' is not provided by --api-port, try to fill it using Docker Machine's IP address.
	if apiPort.Host == "" {
		apiPort.Host, err = getDockerMachineIp()
		// IP address is the same as the host
		apiPort.HostIP = apiPort.Host
		// In case of error, Log a warning message, and continue on. Since it more likely caused by a miss configured
		// DOCKER_MACHINE_NAME environment variable.
		if err != nil {
			log.Printf("WARNING: Failed to get docker machine IP address, ignoring the DOCKER_MACHINE_NAME environment variable setting.\n")
		}
	}

	if apiPort.Host != "" {
		// Add TLS SAN for non default host name
		log.Printf("Add TLS SAN for %s", apiPort.Host)
		k3sServerArgs = append(k3sServerArgs, "--tls-san", apiPort.Host)
	}

	if c.IsSet("server-arg") || c.IsSet("x") {
		k3sServerArgs = append(k3sServerArgs, c.StringSlice("server-arg")...)
	}

	k3AgentArgs := []string{}
	if c.IsSet("agent-arg") {
		k3AgentArgs = append(k3AgentArgs, c.StringSlice("agent-arg")...)
	}

	// new port map
	portmap, err := mapNodesToPortSpecs(c.StringSlice("publish"), GetAllContainerNames(c.String("name"), defaultServerCount, c.Int("workers")))
	if err != nil {
		log.Fatal(err)
	}

	// create a docker volume for sharing image tarballs with the cluster
	imageVolume, err := createImageVolume(c.String("name"))
	log.Println("Created docker volume ", imageVolume.Name)
	if err != nil {
		return err
	}
	volumes := c.StringSlice("volume")
	volumes = append(volumes, fmt.Sprintf("%s:/images", imageVolume.Name))

	clusterSpec := &ClusterSpec{
		AgentArgs:         k3AgentArgs,
		APIPort:           *apiPort,
		AutoRestart:       c.Bool("auto-restart"),
		ClusterName:       c.String("name"),
		Env:               env,
		Image:             image,
		NodeToPortSpecMap: portmap,
		PortAutoOffset:    c.Int("port-auto-offset"),
		ServerArgs:        k3sServerArgs,
		Verbose:           c.GlobalBool("verbose"),
		Volumes:           volumes,
		AgentOnly:         c.Bool("agent-only"),
		ServerURL:         c.String("server"),
	}

	// create the server
	log.Printf("Creating cluster [%s]", c.String("name"))

	// create the directory where we will put the kubeconfig file by default (when running `k3d get-config`)
	createClusterDir(c.String("name"))

	dockerID := ""
	if !c.Bool("agent-only") {
		dockerID, err = createServer(clusterSpec)
		if err != nil {
			deleteCluster()
			return err
		}
	}

	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Wait for k3s to be up and running if wanted.
	// We're simply scanning the container logs for a line that tells us that everything's up and running
	// TODO: also wait for worker nodes
	start := time.Now()
	timeout := time.Duration(c.Int("wait")) * time.Second
	for c.IsSet("wait") && !c.Bool("agent-only") {
		// not running after timeout exceeded? Rollback and delete everything.
		if timeout != 0 && time.Now().After(start.Add(timeout)) {
			deleteCluster()
			return errors.New("Cluster creation exceeded specified timeout")
		}

		// scan container logs for a line that tells us that the required services are up and running
		out, err := docker.ContainerLogs(ctx, dockerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			out.Close()
			return fmt.Errorf("ERROR: couldn't get docker logs for %s\n%+v", c.String("name"), err)
		}
		buf := new(bytes.Buffer)
		nRead, _ := buf.ReadFrom(out)
		out.Close()
		output := buf.String()
		if nRead > 0 && strings.Contains(string(output), "Running kubelet") {
			break
		}

		time.Sleep(1 * time.Second)
	}

	// spin up the worker nodes
	// TODO: do this concurrently in different goroutines
	if c.Int("workers") > 0 {
		log.Printf("Booting %s workers for cluster %s", strconv.Itoa(c.Int("workers")), c.String("name"))
		for i := 0; i < c.Int("workers"); i++ {
			workerID, err := createWorker(clusterSpec, i)
			if err != nil {
				deleteCluster()
				return err
			}
			log.Printf("Created worker with ID %s\n", workerID)
		}
	}

	log.Printf("SUCCESS: created cluster [%s]", c.String("name"))
	log.Printf(`You can now use the cluster with:

export KUBECONFIG="$(%s get-kubeconfig --name='%s')"
kubectl cluster-info`, os.Args[0], c.String("name"))

	return nil
}

// DeleteCluster removes the containers belonging to a cluster and its local directory
func DeleteCluster(c *cli.Context) error {
	clusters, err := getClusters(c.Bool("all"), c.String("name"))

	if err != nil {
		return err
	}

	// remove clusters one by one instead of appending all names to the docker command
	// this allows for more granular error handling and logging
	for _, cluster := range clusters {
		log.Printf("Removing cluster [%s]", cluster.name)
		if len(cluster.workers) > 0 {
			// TODO: this could be done in goroutines
			log.Printf("...Removing %d workers\n", len(cluster.workers))
			for _, worker := range cluster.workers {
				if err := removeContainer(worker.ID); err != nil {
					log.Println(err)
					continue
				}
			}
		}
		deleteClusterDir(cluster.name)
		log.Println("...Removing server")
		if err := removeContainer(cluster.server.ID); err != nil {
			return fmt.Errorf("ERROR: Couldn't remove server for cluster %s\n%+v", cluster.name, err)
		}

		if err := deleteClusterNetwork(cluster.name); err != nil {
			log.Printf("WARNING: couldn't delete cluster network for cluster %s\n%+v", cluster.name, err)
		}

		log.Println("...Removing docker image volume")
		if err := deleteImageVolume(cluster.name); err != nil {
			log.Printf("WARNING: couldn't delete image docker volume for cluster %s\n%+v", cluster.name, err)
		}

		log.Printf("SUCCESS: removed cluster [%s]", cluster.name)
	}

	return nil
}

// StopCluster stops a running cluster container (restartable)
func StopCluster(c *cli.Context) error {
	clusters, err := getClusters(c.Bool("all"), c.String("name"))

	if err != nil {
		return err
	}

	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("ERROR: couldn't create docker client\n%+v", err)
	}

	// remove clusters one by one instead of appending all names to the docker command
	// this allows for more granular error handling and logging
	for _, cluster := range clusters {
		log.Printf("Stopping cluster [%s]", cluster.name)
		if len(cluster.workers) > 0 {
			log.Printf("...Stopping %d workers\n", len(cluster.workers))
			for _, worker := range cluster.workers {
				if err := docker.ContainerStop(ctx, worker.ID, nil); err != nil {
					log.Println(err)
					continue
				}
			}
		}
		log.Println("...Stopping server")
		if err := docker.ContainerStop(ctx, cluster.server.ID, nil); err != nil {
			return fmt.Errorf("ERROR: Couldn't stop server for cluster %s\n%+v", cluster.name, err)
		}

		log.Printf("SUCCESS: Stopped cluster [%s]", cluster.name)
	}

	return nil
}

// StartCluster starts a stopped cluster container
func StartCluster(c *cli.Context) error {
	clusters, err := getClusters(c.Bool("all"), c.String("name"))

	if err != nil {
		return err
	}

	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("ERROR: couldn't create docker client\n%+v", err)
	}

	// remove clusters one by one instead of appending all names to the docker command
	// this allows for more granular error handling and logging
	for _, cluster := range clusters {
		log.Printf("Starting cluster [%s]", cluster.name)

		log.Println("...Starting server")
		if err := docker.ContainerStart(ctx, cluster.server.ID, types.ContainerStartOptions{}); err != nil {
			return fmt.Errorf("ERROR: Couldn't start server for cluster %s\n%+v", cluster.name, err)
		}

		if len(cluster.workers) > 0 {
			log.Printf("...Starting %d workers\n", len(cluster.workers))
			for _, worker := range cluster.workers {
				if err := docker.ContainerStart(ctx, worker.ID, types.ContainerStartOptions{}); err != nil {
					log.Println(err)
					continue
				}
			}
		}

		log.Printf("SUCCESS: Started cluster [%s]", cluster.name)
	}

	return nil
}

// ListClusters prints a list of created clusters
func ListClusters(c *cli.Context) error {
	if c.IsSet("all") {
		log.Println("INFO: --all is on by default, thus no longer required. This option will be removed in v2.0.0")

	}
	printClusters()
	return nil
}

// GetKubeConfig grabs the kubeconfig from the running cluster and prints the path to stdout
func GetKubeConfig(c *cli.Context) error {
	cluster := c.String("name")
	kubeConfigPath, err := getKubeConfig(cluster)
	if err != nil {
		return err
	}

	// output kubeconfig file path to stdout
	fmt.Println(kubeConfigPath)
	return nil
}

// Shell starts a new subshell with the KUBECONFIG pointing to the selected cluster
func Shell(c *cli.Context) error {
	return subShell(c.String("name"), c.String("shell"), c.String("command"))
}

// ImportImage saves an image locally and imports it into the k3d containers
func ImportImage(c *cli.Context) error {
	images := make([]string, 0)
	if strings.Contains(c.Args().First(), ",") {
		images = append(images, strings.Split(c.Args().First(), ",")...)
	} else {
		images = append(images, c.Args()...)
	}
	return importImage(c.String("name"), images, c.Bool("no-remove"))
}
