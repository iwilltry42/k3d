# Command Tree

```bash
k3d
  --runtime  # choose the container runtime (default: docker)
  --verbose  # enable verbose (debug) logging (default: false)
  create
    cluster [CLUSTERNAME]  # default cluster name is 'k3s-default'
      -a, --api-port  # specify the port on which the cluster will be accessible (e.g. via kubectl)
      -i, --image  # specify which k3s image should be used for the nodes
      --k3s-agent-arg  # add additional arguments to the k3s agent (see https://rancher.com/docs/k3s/latest/en/installation/install-options/agent-config/#k3s-agent-cli-help)
      --k3s-server-arg  # add additional arguments to the k3s server (see https://rancher.com/docs/k3s/latest/en/installation/install-options/server-config/#k3s-server-cli-help)
      -m, --masters  # specify how many master nodes you want to create
      --network  # specify a network you want to connect to
      --no-image-volume  # disable the creation of a volume for storing images (used for the 'k3d load image' command)
      -p, --port  # add some more port mappings
      --secret  # specify a cluster secret (default: auto-generated)
      --timeout  # specify a timeout, after which the cluster creation will be interrupted and changes rolled back
      --update-kubeconfig  # enable the automated update of the default kubeconfig with the details of the newly created cluster (also sets '--wait=true')
      -v, --volume  # specify additional bind-mounts
      --wait  # enable waiting for all master nodes to be ready before returning
      -w, --workers  # specify how many worker nodes you want to create
    node NODENAME  # Create new nodes (and add them to existing clusters)
      -c, --cluster  # specify the cluster that the node shall connect to
      -i, --image  # specify which k3s image should be used for the node(s)
          --replicas  # specify how many replicas you want to create with this spec
          --role  # specify the node role
  delete
    cluster CLUSTERNAME  # delete an existing cluster
      -a, --all  # delete all existing clusters
    node NODENAME  # delete an existing node
      -a, --all  # delete all existing nodes
  start
    cluster CLUSTERNAME  # start a (stopped) cluster
      -a, --all  # start all clusters
    node NODENAME  # start a (stopped) node
  stop
    cluster CLUSTERNAME  # stop a cluster
      -a, --all  # stop all clusters
    node  # stop a node
  get
    cluster [CLUSTERNAME [CLUSTERNAME ...]]
      --no-headers
    node
    kubeconfig CLUSTERNAME
  load
    image  [IMAGE [IMAGE ...]]  # Load one or more images from the local runtime environment into k3d clusters
      -c, --cluster
      -k, --keep-tarball, -k
      -t, --tar, -t
  completion SHELL  # Generate completion scripts
  version  # show k3d build version
  help [COMMAND]  # show help text for any command
```
