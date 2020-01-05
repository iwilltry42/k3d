# Thoughts

## commands

### `create`

```shell
k3d
  |- create
  |       |- cluster [NAME ] [flags]
  |       |- node [NAME ] [flags]
  |
  |- delete
  |       |- cluster [NAME ] [flags]
  |       |- node [NAME ] [flags]
  |- get
  |    |- cluster
  |    |- node
  |- start
  |      |- cluster
  |      |- node
  |- stop
  |     |- cluster
  |     |- node
```

## Overview

- `cmd/`: everything around the CLI of k3d = human interface, printed output (e.g. list of clusters)
- `pkg/`: everything else, can be used as a module from other Go projects
  - `cluster/`: everything around managing cluster components
  - `runtimes/`: translate k3d types (node, cluster, etc.) to container runtime specific types and manage them
  - `types/`: collection of types (structs) and constants used by k3d
  - `util/`: utilities, that could be used for everything, not directly related to the project

## k3d <-> runtime

k3d _should_ work with more than one runtime, if we can implement the Runtime interface for it.
Here's how k3d types should translate to a runtime type:

- `cluster` = set of _containers_ running in the same _network_, maybe mounting the same _volume(s)_
- `node` = _container_ with _exposed ports_ and _volume mounts_

### docker

#### node -> container

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

## expose ports / volumes => DONE

- `--port [host:]port[:containerPort][/protocol][@group_identifier[[index] | @node_identifier]`
  - Examples:
    - `--port 0.0.0.0:8080:8081/tcp@workers` -> whole group
    - `--port 80@workers[0]` -> single instance of group by list index
    - `--port 80@workers[0,2-3]` -> multiple instances of a group by index lists and ranges
    - `--port 80@k3d-test-worker-0` -> single instance by specific node identifier
    - `--port 80@k3d-test-master-0@workers[1-5]` -> multiple instances by combination of node and group identifiers

- analogous for volumes

## multi master setup => WIP

- if `--masters` > 1 deploy a load-balancer in front of them as an extra container
  - consider that in the kubeconfig file and `--tls-san`
  - make this the default, but provide a `--no-lb` flag

## Store additional created stuff in labels => DONE

- when creating a cluster, usually, you also create a new docker network (and maybe other resources)
  - store a reference to those in the container labels of cluster nodes
  - when deleting the cluster, parse the labels, deduplicate the results and delete the additional resources
  - DONE for network
    - new labels `k3d.cluster.network=<ID>` and `k3d.cluster.network.external=<true|false>` (determine whether to try to delete it when you delete a cluster, since network may have been created manually)


# Comparison to k3d v1

- k3d
  - check-tools
  - shell
    - --name
    - --command
    - --shell
      - auto, bash, zsh
  - create          -> `k3d create cluster CLUSTERNAME`
    - --name          -> y
    - --volume        -> y
    - --port          -> y
    - --api-port      -> y
    - --wait
    - --image         -> y
    - --server-arg    -> y
    - --agent-arg     -> y
    - --env
    - --workers       -> y
    - --auto-restart
  - (add-node)      -> `k3d create node NODENAME`
    - --role
    - --name
    - --count
    - --image
    - --arg
    - --env
    - --volume
    - --k3s
    - --k3s-secret
    - --k3s-token
  - delete          -> `k3d delete cluster CLUSTERNAME`
    - --name
    - --all
  - stop            -> `k3d stop cluster CLUSTERNAME`
    - --name
    - --all
  - start           -> `k3d start cluster CLUSTERNAME`
    - --name
    - --all
  - list
  - get-kubeconfig  -> `k3d get kubeconfig CLUSTERNAME`
    - --name        -> y
    - --all
  - import-images   -> `k3d load image [--cluster CLUSTERNAME] [--keep] IMAGES`
    - --name        -> y
    - --no-remove   -> y

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

## tools

- maybe rename `k3d load` to `k3d tools` and add tool cmds there?
  - e.g. `k3d tools import-images`
  - let's you set tools container version
    - `k3d tools --image k3d-tools:v2 import-images`
- add `k3d create --image-vol NAME` flag to re-use existing image volume
  - will add `k3d.volumes.imagevolume.external: true` label to nodes
    - should not be deleted with cluster
  - possibly add `k3d create volume` and `k3d create network` to create external networks?

## extra commands

- `k3d prune` to prune all dangling resources
  - nodes, volumes, networks

## use OCI

- [https://github.com/opencontainers/runtime-spec/blob/master/specs-go/config.go](https://github.com/opencontainers/runtime-spec/blob/master/specs-go/config.go)
- move node -> container translation out of runtime

## node configuration comparison

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

- remove/add nodes -> needs to remove line in `/var/lib/rancher/k3s/server/cred/node-passwd` for the deleted node
