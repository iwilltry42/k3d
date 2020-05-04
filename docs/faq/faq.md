# FAQ / Nice to know

## Issues with BTRFS

- As [@jaredallard](https://github.com/jaredallard) [pointed out](https://github.com/rancher/k3d/pull/48), people running `k3d` on a system with **btrfs**, may need to mount `/dev/mapper` into the nodes for the setup to work.
  - This will do: `k3d create cluster CLUSTER_NAME -v /dev/mapper:/dev/mapper`

## Issues with ZFS

- k3s currently has [no support for ZFS](ttps://github.com/rancher/k3s/issues/66) and thus, creating multi-master setups (e.g. `k3d create cluster multimaster --masters 3`) fails, because the initializing master node (server flag `--cluster-init`) errors out with the following log:
  ```bash
  starting kubernetes: preparing server: start cluster and https: raft_init(): io: create I/O capabilities probe file: posix_allocate: operation not supported on socket
  ```

  - This issue can be worked around by providing docker with a different filesystem (that's also better for docker-in-docker stuff).
  - A possible solution can be found here: [https://github.com/rancher/k3s/issues/1688#issuecomment-619570374](https://github.com/rancher/k3s/issues/1688#issuecomment-619570374)

## Pods evicted due to lack of disk space

- Pods go to evicted state after doing X
  - Related issues: [#133 - Pods evicted due to `NodeHasDiskPressure`](https://github.com/rancher/k3d/issues/133) (collection of #119 and #130)
  - Background: somehow docker runs out of space for the k3d node containers, which triggers a hard eviction in the kubelet
  - Possible [fix/workaround by @zer0def](https://github.com/rancher/k3d/issues/133#issuecomment-549065666):
    - use a docker storage driver which cleans up properly (e.g. overlay2)
    - clean up or expand docker root filesystem
    - change the kubelet's eviction thresholds upon cluster creation: `k3d create cluster --k3s-agent-arg '--kubelet-arg=eviction-hard=imagefs.available<1%,nodefs.available<1%' --k3s-agent-arg '--kubelet-arg=eviction-minimum-reclaim=imagefs.available=1%,nodefs.available=1%'`
