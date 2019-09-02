package run

/*
 * The functions in this file take care of spinning up the
 * k3s server and worker containers as well as deleting them.
 */

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type ClusterSpec struct {
	AgentArgs         []string
	APIPort           apiPort
	AutoRestart       bool
	ClusterName       string
	Env               []string
	Image             string
	NodeToPortSpecMap map[string][]string
	PortAutoOffset    int
	ServerArgs        []string
	Verbose           bool
	Volumes           []string
	AgentOnly         bool
	ServerURL         string
}

func startContainer(verbose bool, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (string, error) {
	ctx := context.Background()

	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("ERROR: couldn't create docker client\n%+v", err)
	}

	resp, err := docker.ContainerCreate(ctx, config, hostConfig, networkingConfig, containerName)
	if client.IsErrNotFound(err) {
		log.Printf("Pulling image %s...\n", config.Image)
		reader, err := docker.ImagePull(ctx, config.Image, types.ImagePullOptions{})
		if err != nil {
			return "", fmt.Errorf("ERROR: couldn't pull image %s\n%+v", config.Image, err)
		}
		defer reader.Close()
		if verbose {
			_, err := io.Copy(os.Stdout, reader)
			if err != nil {
				log.Printf("WARNING: couldn't get docker output\n%+v", err)
			}
		} else {
			_, err := io.Copy(ioutil.Discard, reader)
			if err != nil {
				log.Printf("WARNING: couldn't get docker output\n%+v", err)
			}
		}
		resp, err = docker.ContainerCreate(ctx, config, hostConfig, networkingConfig, containerName)
		if err != nil {
			return "", fmt.Errorf("ERROR: couldn't create container after pull %s\n%+v", containerName, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("ERROR: couldn't create container %s\n%+v", containerName, err)
	}

	if err := docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func createServer(spec *ClusterSpec) (string, error) {
	log.Printf("Creating server using %s...\n", spec.Image)

	containerLabels := make(map[string]string)
	containerLabels["app"] = "k3d"
	containerLabels["component"] = "server"
	containerLabels["created"] = time.Now().Format("2006-01-02 15:04:05")
	containerLabels["cluster"] = spec.ClusterName

	containerName := GetContainerName("server", spec.ClusterName, -1)

	// ports to be assigned to the server belong to roles
	// all, server or <server-container-name>
	serverPorts, err := MergePortSpecs(spec.NodeToPortSpecMap, "server", containerName)
	if err != nil {
		return "", err
	}

	hostIP := "0.0.0.0"
	containerLabels["apihost"] = "localhost"
	if spec.APIPort.Host != "" {
		hostIP = spec.APIPort.HostIP
		containerLabels["apihost"] = spec.APIPort.Host
	}

	apiPortSpec := fmt.Sprintf("%s:%s:%s/tcp", hostIP, spec.APIPort.Port, spec.APIPort.Port)

	serverPorts = append(serverPorts, apiPortSpec)

	serverPublishedPorts, err := CreatePublishedPorts(serverPorts)
	if err != nil {
		log.Fatalf("Error: failed to parse port specs %+v \n%+v", serverPorts, err)
	}

	hostConfig := &container.HostConfig{
		PortBindings: serverPublishedPorts.PortBindings,
		Privileged:   true,
	}

	if spec.AutoRestart {
		hostConfig.RestartPolicy.Name = "unless-stopped"
	}

	if len(spec.Volumes) > 0 && spec.Volumes[0] != "" {
		hostConfig.Binds = spec.Volumes
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			k3dNetworkName(spec.ClusterName): {
				Aliases: []string{containerName},
			},
		},
	}

	config := &container.Config{
		Hostname:     containerName,
		Image:        spec.Image,
		Cmd:          append([]string{"server"}, spec.ServerArgs...),
		ExposedPorts: serverPublishedPorts.ExposedPorts,
		Env:          spec.Env,
		Labels:       containerLabels,
	}
	id, err := startContainer(spec.Verbose, config, hostConfig, networkingConfig, containerName)
	if err != nil {
		return "", fmt.Errorf("ERROR: couldn't create container %s\n%+v", containerName, err)
	}

	return id, nil
}

// createWorker creates/starts a k3s agent node that connects to the server
func createWorker(spec *ClusterSpec, postfix int) (string, error) {
	containerLabels := make(map[string]string)
	containerLabels["app"] = "k3d"
	containerLabels["component"] = "worker"
	containerLabels["created"] = time.Now().Format("2006-01-02 15:04:05")
	containerLabels["cluster"] = spec.ClusterName

	containerName := GetContainerName("worker", spec.ClusterName, postfix)
	env := spec.Env

	// ports to be assigned to the server belong to roles
	// all, server or <server-container-name>
	workerPorts, err := MergePortSpecs(spec.NodeToPortSpecMap, "worker", containerName)
	if err != nil {
		return "", err
	}
	workerPublishedPorts, err := CreatePublishedPorts(workerPorts)
	if err != nil {
		return "", err
	}
	if spec.PortAutoOffset > 0 {
		// TODO: add some checks before to print a meaningful log message saying that we cannot map multiple container ports
		// to the same host port without a offset
		workerPublishedPorts = workerPublishedPorts.Offset(postfix + spec.PortAutoOffset)
	}

	hostConfig := &container.HostConfig{
		Tmpfs: map[string]string{
			"/run":     "",
			"/var/run": "",
		},
		PortBindings: workerPublishedPorts.PortBindings,
		Privileged:   true,
	}

	if spec.AutoRestart {
		hostConfig.RestartPolicy.Name = "unless-stopped"
	}

	if len(spec.Volumes) > 0 && spec.Volumes[0] != "" {
		hostConfig.Binds = spec.Volumes
	}

	networkingConfig := &network.NetworkingConfig{}

	// if we only create agent(s) and have a server url provided for connection,
	// we try to connect to the server's docker network
	// TODO: distinguish between k3d and k3s server connection string
	if spec.AgentOnly && spec.ServerURL != "" {

		ctx := context.Background()
		docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return "", err
		}

		serverName := strings.Split(spec.ServerURL, ":")[0]
		inspectResponse, err := docker.ContainerInspect(ctx, serverName)
		if err != nil {
			return "", err
		}
		serverClusterName := inspectResponse.Config.Labels["cluster"]
		filters := filters.NewArgs()
		filters.Add("label", fmt.Sprintf("cluster=%s", serverClusterName))
		serverNetworks, err := docker.NetworkList(ctx, types.NetworkListOptions{
			Filters: filters,
		})
		if err != nil {
			return "", err
		}
		networkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			serverNetworks[0].Name: {},
		}
		log.Println("Connecting to", serverNetworks[0].Name)
	} else {
		env = append(env, fmt.Sprintf("K3S_URL=https://k3d-%s-server:%s", spec.ClusterName, spec.APIPort.Port))
		networkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			k3dNetworkName(spec.ClusterName): {
				Aliases: []string{containerName},
			},
		}
	}

	config := &container.Config{
		Hostname:     containerName,
		Image:        spec.Image,
		Env:          env,
		Cmd:          append([]string{"agent"}, spec.AgentArgs...),
		Labels:       containerLabels,
		ExposedPorts: workerPublishedPorts.ExposedPorts,
	}

	id, err := startContainer(spec.Verbose, config, hostConfig, networkingConfig, containerName)
	if err != nil {
		return "", fmt.Errorf("ERROR: couldn't start container %s\n%+v", containerName, err)
	}

	return id, nil
}

// removeContainer tries to rm a container, selected by Docker ID, and does a rm -f if it fails (e.g. if container is still running)
func removeContainer(ID string) error {
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("ERROR: couldn't create docker client\n%+v", err)
	}

	options := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := docker.ContainerRemove(ctx, ID, options); err != nil {
		return fmt.Errorf("ERROR: couldn't delete container [%s] -> %+v", ID, err)
	}
	return nil
}
