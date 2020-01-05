# Documentation

## Functionality

### Defaults

* multiple master nodes
  * by default, when `--master` > 1 and no `--datastore-x` option is set, the first master node (master-0) will be the initializing master node
    * the initializing master node will have the `--cluster-init` flag appended
    * all other master nodes will refer to the initializing master node via `--server https://<init-node>:6443`
* API-Ports
  * by default, we don't expose any API-Port (no host port mapping)
