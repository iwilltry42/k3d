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
package types

import (
	"fmt"
)

// DefaultClusterName specifies the default name used for newly created clusters
const DefaultClusterName = "k3s-default"

// DefaultClusterNameMaxLength specifies the maximal length of a passed in cluster name
// This restriction allows us to construct an name consisting of
// <DefaultObjectNamePrefix[3]>-<ClusterName>-<TypeSuffix[5-10]>-<Counter[1-3]>
// ... and still stay within the 64 character limit (e.g. of docker)
const DefaultClusterNameMaxLength = 32

// DefaultK3sImageRepo specifies the default image repository for the used k3s image
const DefaultK3sImageRepo = "docker.io/rancher/k3s"

// DefaultLBImage defines the default cluster load balancer image
const DefaultLBImage = "docker.io/library/traefik:v2.0"

// DefaultObjectNamePrefix defines the name prefix for every object created by k3d
const DefaultObjectNamePrefix = "k3d"

// Role defines a k3d node role
type Role string

// existing k3d node roles
const (
	MasterRole Role = "master"
	WorkerRole Role = "worker"
	NoRole     Role = "nope"
)

// DefaultK3dRoles defines the roles available for nodes
var DefaultK3dRoles = map[string]Role{
	string(MasterRole): MasterRole,
	string(WorkerRole): WorkerRole,
}

// DefaultObjectLabels specifies a set of labels that will be attached to k3d objects by default
var DefaultObjectLabels = map[string]string{
	"app": "k3d",
}

// DefaultTmpfsMounts specifies tmpfs mounts that are required for all k3d nodes
var DefaultTmpfsMounts = []string{
	"/run",
	"/var/run",
}

// DefaultNodeEnv defines some default environment variables that should be set on every node
var DefaultNodeEnv = []string{
	"K3S_KUBECONFIG_OUTPUT=/output/kubeconfig.yaml",
}

// DefaultToolsContainerImage defines the default image used for the tools container
const DefaultToolsContainerImage = "docker.io/iwilltry42/k3d-tools:v0.0.2" // TODO: get version dynamically or at build time

// DefaultImageVolumeMountPath defines the mount path inside k3d nodes where we will mount the shared image volume by default
const DefaultImageVolumeMountPath = "/k3d/images"

// DefaultConfigDirName defines the name of the config directory (where we'll e.g. put the kubeconfigs)
const DefaultConfigDirName = ".k3d" // should end up in $HOME/

// DefaultKubeconfigPrefix defines the default prefix for kubeconfig files
const DefaultKubeconfigPrefix = DefaultObjectNamePrefix + "-kubeconfig"

// CreateClusterOpts describe a set of options one can set when creating a cluster
type CreateClusterOpts struct {
	DisableImageVolume  bool
	DisableLoadbalancer bool
	WaitForMaster       int
	K3sServerArgs       []string
	K3sAgentArgs        []string
}

// ClusterNetwork describes a network which a cluster is running in
type ClusterNetwork struct {
	Name     string `yaml:"name" json:"name,omitempty"`
	External bool   `yaml:"external" json:"isExternal,omitempty"`
}

// Cluster describes a k3d cluster
type Cluster struct {
	Name               string               `yaml:"name" json:"name,omitempty"`
	Network            ClusterNetwork       `yaml:"network" json:"network,omitempty"`
	Secret             string               `yaml:"cluster_secret" json:"clusterSecret,omitempty"`
	Nodes              []*Node              `yaml:"nodes" json:"nodes,omitempty"`
	InitNode           *Node                // init master node
	MasterLoadBalancer *ClusterLoadbalancer `yaml:"master_loadbalancer" json:"masterLoadBalancer,omitempty"`
	ExternalDatastore  ExternalDatastore    `yaml:"external_datastore" json:"externalDatastore,omitempty"`
	CreateClusterOpts  *CreateClusterOpts   `yaml:"options" json:"options,omitempty"`
}

// Node describes a k3d node
type Node struct {
	Name       string            `yaml:"name" json:"name,omitempty"`
	Role       Role              `yaml:"role" json:"role,omitempty"`
	Image      string            `yaml:"image" json:"image,omitempty"`
	Volumes    []string          `yaml:"volumes" json:"volumes,omitempty"`
	Env        []string          `yaml:"env" json:"env,omitempty"`
	Cmd        []string          // filled automatically based on role
	Args       []string          `yaml:"extra_args" json:"extraArgs,omitempty"`
	Ports      []string          `yaml:"port_mappings" json:"portMappings,omitempty"`
	Restart    bool              `yaml:"restart" json:"restart,omitempty"`
	Labels     map[string]string // filled automatically
	Network    string            // filled automatically
	MasterOpts MasterOpts        `yaml:"master_opts" json:"masterOpts,omitempty"`
	WorkerOpts WorkerOpts        `yaml:"worker_opts" json:"workerOpts,omitempty"`
}

// MasterOpts describes some additional master role specific opts
type MasterOpts struct {
	ExposeAPI ExposeAPI `yaml:"expose_api" json:"exposeAPI,omitempty"`
	IsInit    bool      `yaml:"is_initializing_master" json:"isInitializingMaster,omitempty"`
}

// ExternalDatastore describes an external datastore used for HA/multi-master clusters
type ExternalDatastore struct {
	Endpoint string `yaml:"endpoint" json:"endpoint,omitempty"`
	CAFile   string `yaml:"ca_file" json:"caFile,omitempty"`
	CertFile string `yaml:"cert_file" json:"certFile,omitempty"`
	KeyFile  string `yaml:"key_file" json:"keyFile,omitempty"`
	Network  string `yaml:"network" json:"network,omitempty"`
}

// ExposeAPI describes specs needed to expose the API-Server
type ExposeAPI struct {
	Host   string `yaml:"host" json:"host,omitempty"`
	HostIP string `yaml:"host_ip" json:"hostIP,omitempty"`
	Port   string `yaml:"port" json:"port"`
}

// WorkerOpts describes some additional worker role specific opts
type WorkerOpts struct{}

// GetDefaultObjectName prefixes the passed name with the default prefix
func GetDefaultObjectName(name string) string {
	return fmt.Sprintf("%s-%s", DefaultObjectNamePrefix, name)
}

// ClusterLoadbalancer describes a loadbalancer deployed in front of a multi-master cluster
type ClusterLoadbalancer struct {
	Image       string
	ExposedPort string `yaml:"exposed_port" json:"exposedPort,omitempty"`
}
