package main

import (
	"fmt"
	"log"
	"os"

	run "github.com/rancher/k3d/cli"
	"github.com/rancher/k3d/version"
	"github.com/urfave/cli"
)

// defaultK3sImage specifies the default image being used for server and workers
const defaultK3sImage = "docker.io/rancher/k3s"
const defaultK3sClusterName string = "k3s-default"

// main represents the CLI application
func main() {

	// App Details
	app := cli.NewApp()
	app.Name = "k3d"
	app.Usage = "Run k3s in Docker!"
	app.Version = version.GetVersion()
	app.Authors = []cli.Author{
		{
			Name:  "Thorsten Klein",
			Email: "iwilltry42@gmail.com",
		},
		{
			Name:  "Rishabh Gupta",
			Email: "r.g.gupta@outlook.com",
		},
		{
			Name: "Darren Shepherd",
		},
	}

	// commands that you can execute
	app.Commands = []cli.Command{
		{
			// check-tools verifies that docker is up and running
			Name:    "check-tools",
			Aliases: []string{"ct"},
			Usage:   "Check if docker is running",
			Action:  run.CheckTools,
		},
		{
			// shell starts a shell in the context of a running cluster
			Name:  "shell",
			Usage: "Start a subshell for a cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: defaultK3sClusterName,
					Usage: "Set a name for the cluster",
				},
				cli.StringFlag{
					Name:  "command, c",
					Usage: "Run a shell command in the context of the cluster",
				},
				cli.StringFlag{
					Name:  "shell, s",
					Value: "auto",
					Usage: "which shell to use. One of [auto, bash, zsh]",
				},
			},
			Action: run.Shell,
		},
		{
			// create creates a new k3s cluster in docker containers
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create a single- or multi-node k3s cluster in docker containers",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: defaultK3sClusterName,
					Usage: "Set a name for the cluster",
				},
				cli.StringSliceFlag{
					Name:  "volume, v",
					Usage: "Mount one or more volumes into every node of the cluster (Docker notation: `source:destination`)",
				},
				cli.StringSliceFlag{
					Name:  "publish, add-port",
					Usage: "Publish k3s node ports to the host (Format: `[ip:][host-port:]container-port[/protocol]@node-specifier`, use multiple options to expose more ports)",
				},
				cli.IntFlag{
					Name:  "port-auto-offset",
					Value: 0,
					Usage: "Automatically add an offset (* worker number) to the chosen host port when using `--publish` to map the same container-port from multiple k3d workers to the host",
				},
				cli.StringFlag{
					// TODO: to be deprecated
					Name:  "version",
					Usage: "Choose the k3s image version",
				},
				cli.StringFlag{
					// TODO: only --api-port, -a soon since we want to use --port, -p for the --publish/--add-port functionality
					Name:  "api-port, a, port, p",
					Value: "6443",
					Usage: "Specify the Kubernetes cluster API server port (Format: `[host:]port` (Note: --port/-p will be used for arbitrary port mapping as of v2.0.0, use --api-port/-a instead for setting the api port)",
				},
				cli.IntFlag{
					Name:  "wait, t",
					Value: 0, // timeout
					Usage: "Wait for the cluster to come up before returning until timoout (in seconds). Use --wait 0 to wait forever",
				},
				cli.StringFlag{
					Name:  "image, i",
					Usage: "Specify a k3s image (Format: <repo>/<image>:<tag>)",
					Value: fmt.Sprintf("%s:%s", defaultK3sImage, version.GetK3sVersion()),
				},
				cli.StringSliceFlag{
					Name:  "server-arg, x",
					Usage: "Pass an additional argument to k3s server (new flag per argument)",
				},
				cli.StringSliceFlag{
					Name:  "agent-arg",
					Usage: "Pass an additional argument to k3s agent (new flag per argument)",
				},
				cli.StringSliceFlag{
					Name:  "env, e",
					Usage: "Pass an additional environment variable (new flag per variable)",
				},
				cli.IntFlag{
					Name:  "workers, w",
					Value: 0,
					Usage: "Specify how many worker nodes you want to spawn",
				},
				cli.BoolFlag{
					Name:  "auto-restart",
					Usage: "Set docker's --restart=unless-stopped flag on the containers",
				},
				cli.BoolFlag{
					Name:  "agent-only",
					Usage: "Do not create the server",
				},
				cli.StringFlag{
					Name:  "server",
					Usage: "Server connection string for agent-only cluster",
				},
				cli.StringFlag{
					Name:  "secret",
					Usage: "Server connection secret for agent-only cluster",
				},
			},
			/*Subcommands: cli.Commands{
				cli.Command{
					Name:  "agent",
					Usage: "Create one or more k3s agent containers",
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "replicas",
							Value: 1,
							Usage: "Specify how many agents you want to create",
						},
					},
					Action: run.CreateAgent,
				},
			},*/
			Action: run.CreateCluster,
		},
		{
			// delete deletes an existing k3s cluster (remove container and cluster directory)
			Name:    "delete",
			Aliases: []string{"d", "del"},
			Usage:   "Delete cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: defaultK3sClusterName,
					Usage: "name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Delete all existing clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.DeleteCluster,
		},
		{
			// stop stopy a running cluster (its container) so it's restartable
			Name:  "stop",
			Usage: "Stop cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: defaultK3sClusterName,
					Usage: "Name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Stop all running clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.StopCluster,
		},
		{
			// start restarts a stopped cluster container
			Name:  "start",
			Usage: "Start a stopped cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: defaultK3sClusterName,
					Usage: "Name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Start all stopped clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.StartCluster,
		},
		{
			// list prints a list of created clusters
			Name:    "list",
			Aliases: []string{"ls", "l"},
			Usage:   "List all clusters",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Also show non-running clusters",
				},
			},
			Action: run.ListClusters,
		},
		{
			// get-kubeconfig grabs the kubeconfig from the cluster and prints the path to it
			Name:  "get-kubeconfig",
			Usage: "Get kubeconfig location for cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: defaultK3sClusterName,
					Usage: "Name of the cluster",
				},
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Get kubeconfig for all clusters (this ignores the --name/-n flag)",
				},
			},
			Action: run.GetKubeConfig,
		},
		{
			// get-kubeconfig grabs the kubeconfig from the cluster and prints the path to it
			Name:    "import-images",
			Aliases: []string{"i"},
			Usage:   "Import a comma- or space-separated list of container images from your local docker daemon into the cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n, cluster, c",
					Value: defaultK3sClusterName,
					Usage: "Name of the cluster",
				},
				cli.BoolFlag{
					Name:  "no-remove, no-rm, keep, k",
					Usage: "Disable automatic removal of the tarball",
				},
			},
			Action: run.ImportImage,
		},
	}

	// Global flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Enable verbose output",
		},
	}

	// run the whole thing
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
