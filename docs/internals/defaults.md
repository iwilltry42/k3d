# Defaults

* multiple master nodes
  * by default, when `--master` > 1 and no `--datastore-x` option is set, the first master node (master-0) will be the initializing master node
    * the initializing master node will have the `--cluster-init` flag appended
    * all other master nodes will refer to the initializing master node via `--server https://<init-node>:6443`
* API-Ports
  * by default, we don't expose any API-Port (no host port mapping)
* kubeconfig
  * if no output is set explicitly (via the `--output` flag), we use the default loading rules to get the default kubeconfig:
    * First: kubeconfig specified via the KUBECONFIG environment variable (error out if multiple are specified)
    * Second: default kubeconfig in home directory (e.g. `$HOME/.kube/config`)
