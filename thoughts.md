# Thoughts

## Command Tree

- k3d
  - create
    - cluster NAME
      - --api-port
      - --datastore-cafile
      - --datastore-certfile
      - --datastore-endpoint
      - --datastore-keyfile
      - --datastore-network
      - --image
      - --k3s-agent-arg
      - --k3s-server-arg
      - --lb-port
      - --masters
      - --network
      - --no-lb
      - --port
      - --secret
      - --volume
      - --workers
    - node NAME
      - --cluster
      - --image
      - --replicas
      - --role
  - delete
    - cluster NAME
      - --all
    - node NAME
      - --all
  - get
    - cluster NAME
      - --no-headers
    - node NAME
      - --no-headers
    - kubeconfig NAME
      - --output
  - start
    - cluster NAME
      - --all
    - node NAME
  - stop
    - cluster NAME
      - --all
    - node NAME

## Feature Comparison to k3d v1

### v1.x feature -> implementation in v3

- k3d
  - check-tools                   -> won't do
  - shell                         -> planned: `k3d shell CLUSTER`
    - --name                      -> planned: drop (now as arg)
    - --command                   -> planned: keep
    - --shell                     -> planned: keep (or second arg)
      - auto, bash, zsh
  - create                        -> `k3d create cluster CLUSTERNAME`
    - --name                      -> dropped, implemented via arg
    - --volume                    -> implemented
    - --port                      -> implemented
    - --port-auto-offset          -> TBD
    - --api-port                  -> implemented
    - --wait                      -> implemented
    - --image                     -> implemented
    - --server-arg                -> implemented as `--k3s-server-arg`
    - --agent-arg                 -> implemented as `--k3s-agent-arg`
    - --env                       -> planned
    - --label                     -> planned
    - --workers                   -> implemented
    - --auto-restart              -> planned
    - --enable-registry           -> planned (possible consolidation into less registry-related commands?)
    - --registry-name             -> TBD
    - --registry-port             -> TBD
    - --registry-volume           -> TBD
    - --registries-file           -> TBD
    - --enable-registry-cache     -> TBD
  - (add-node)                    -> `k3d create node NODENAME`
    - --role                      -> implemented
    - --name                      -> dropped, implemented as arg
    - --count                     -> implemented as `--replicas`
    - --image                     -> implemented
    - --arg                       -> planned
    - --env                       -> planned
    - --volume                    -> planned
    - --k3s                       -> TBD
    - --k3s-secret                -> TBD
    - --k3s-token                 -> TBD
  - delete                        -> `k3d delete cluster CLUSTERNAME`
    - --name                      -> dropped, implemented as arg
    - --all                       -> implemented
    - --prune                     -> TBD
    - --keep-registry-volume      -> TBD
  - stop                          -> `k3d stop cluster CLUSTERNAME`
    - --name                      -> dropped, implemented as arg
    - --all                       -> implemented
  - start                         -> `k3d start cluster CLUSTERNAME`
    - --name                      -> dropped, implemented as arg
    - --all                       -> implemented
  - list                          -> dropped, implemented as `k3d get clusters`
  - get-kubeconfig                -> `k3d get kubeconfig CLUSTERNAME`
    - --name                      -> dropped, implemented as arg
    - --all                       -> planned
    - --overwrite                 -> planned
  - import-images                 -> `k3d load image [--cluster CLUSTERNAME] [--keep] IMAGES`
    - --name                      -> implemented as `--cluster`
    - --no-remove                 -> implemented as `--keep`

## Repository/Package Overview

- `cmd/`: everything around the CLI of k3d = human interface, printed output (e.g. list of clusters)
- `pkg/`: everything else, can be used as a module from other Go projects
  - `cluster/`: everything around managing cluster components
  - `runtimes/`: translate k3d types (node, cluster, etc.) to container runtime specific types and manage them
  - `types/`: collection of types (structs) and constants used by k3d
  - `util/`: utilities, that could be used for everything, not directly related to the project

## k3d types <-> runtime translation

k3d _should_ work with more than one runtime, if we can implement the Runtime interface for it.
Here's how k3d types should translate to a runtime type:

- `cluster` = set of _containers_ running in the same _network_, maybe mounting the same _volume(s)_
- `node` = _container_ with _exposed ports_ and _volume mounts_

### Docker

#### Node to Container translation

`container = "github.com/docker/docker/api/types/container"`
`network = "github.com/docker/docker/api/types/network"`

- Name    -> container.Hostname = node.Name
- Role    -> container.Labels["k3d.role"] = node.Role
- Image   -> container.Image = node.Image
- Volumes -> container.HostConfig.PortBindings
- Env     -> 
- Args    -> 
- Ports   -> 
- Restart -> 
- Labels  -> container.Labels

## Node Configuration

- master node(s)
  - ENV
    - `K3S_CLUSTER_INIT`
      - if num_masters > 1 && no external datastore configured
    - `K3S_KUBECONFIG_OUTPUT`
      - k3d default -> `/output/kubeconfig.yaml`
  - CMD/ARGS
    - `--https-listen-port`
      - can/should be left default (unset = 6443), since we handle it via port mapping
      - `--tls-san=<some-ip-or-hostname>`
        - get from `--api-port` k3d flag and/or from docker machine
  - Runtime Configuration
    - nothing special
- all nodes
  - ENV
    - `K3S_TOKEN` for node authentication
  - CMD/ARGS
    - nothing special
  - Runtime Configuration
    - Volumes
      - shared image volume
        - cluster-specific (create cluster) or inherit from existing (create node)
      - tmpfs for k3s to work properly
        - `/run`
        - `/var/run`
    - Capabilities/Security Context
      - `privileged`
    - Network
      - cluster network or external/inherited
- worker nodes
  - ENV
    - `K3S_URL` to connect to master node
      - server hostname + port (6443)
      - cluster-specific or inherited
  - CMD/ARGS
    - nothing special
  - Runtime Configuration
    - nothing special

## Features

## [DONE] Node Filters

- `--port [host:]port[:containerPort][/protocol][@group_identifier[[index] | @node_identifier]`
  - Examples:
    - `--port 0.0.0.0:8080:8081/tcp@workers` -> whole group
    - `--port 80@workers[0]` -> single instance of group by list index
    - `--port 80@workers[0,2-3]` -> multiple instances of a group by index lists and ranges
    - `--port 80@k3d-test-worker-0` -> single instance by specific node identifier
    - `--port 80@k3d-test-master-0@workers[1-5]` -> multiple instances by combination of node and group identifiers

- analogous for volumes

## [WIP] Multi-Master Setup

- if `--masters` > 1 deploy a load-balancer in front of them as an extra container
  - consider that in the kubeconfig file and `--tls-san`
  - make this the default, but provide a `--no-lb` flag

## [DONE] Keep State in Docker Labels

- when creating a cluster, usually, you also create a new docker network (and maybe other resources)
  - store a reference to those in the container labels of cluster nodes
  - when deleting the cluster, parse the labels, deduplicate the results and delete the additional resources
  - DONE for network
    - new labels `k3d.cluster.network=<ID>` and `k3d.cluster.network.external=<true|false>` (determine whether to try to delete it when you delete a cluster, since network may have been created manually)

## Bonus Ideas

### Tools

- maybe rename `k3d load` to `k3d tools` and add tool cmds there?
  - e.g. `k3d tools import-images`
  - let's you set tools container version
    - `k3d tools --image k3d-tools:v2 import-images`
- add `k3d create --image-vol NAME` flag to re-use existing image volume
  - will add `k3d.volumes.imagevolume.external: true` label to nodes
    - should not be deleted with cluster
  - possibly add `k3d create volume` and `k3d create network` to create external volumes/networks?

### Prune Command

- `k3d prune` to prune all dangling resources
  - nodes, volumes, networks

### Use Open Standards (OCI, CRI, ...)

- [https://github.com/opencontainers/runtime-spec/blob/master/specs-go/config.go](https://github.com/opencontainers/runtime-spec/blob/master/specs-go/config.go)
- move node -> container translation out of runtime

### Private registry

- create a private registry to be used by k3d clusters
  - similar to [https://github.com/rancher/k3d/pull/161](https://github.com/rancher/k3d/pull/161)
- add `k3d create registry` command to create external registry (maybe instead of flags as in PR #161?)

### Syntactical shortcuts for k3d v1 backwards compatibility

- e.g. `k3d create` -> `k3d create cluster k3s-default`

### Unsorted Ideas

- Integrate build tool (e.g. buildkit, buildah, ...)
- use `tools.go` to keep tools (like `golangci-lint` and `gox`) dependencies
  - see e.g. [https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
  - see e.g. [https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)

### Possible Enhancements

- [!] remove/add nodes -> needs to remove line in `/var/lib/rancher/k3s/server/cred/node-passwd` for the deleted node
- directly use cluster struct in `createClusterCmd` filled by flag values
