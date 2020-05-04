# rancher/k3d

k3d is a lightweight wrapper to run [k3s](https://github.com/rancher/k3s) (Rancher Lab's minimal Kubernetes distribution) in docker.

k3d makes it very easy to create single- and multi-node [k3s](https://github.com/rancher/k3s) clusters in docker, e.g. for local development on Kubernetes.

## Quick Start

Create a cluster named `mycluster` with just a single master node:

```bash
k3d create cluster mycluster
```

Get the new cluster's connection details merged into your default kubeconfig (usually specified using the `KUBECONFIG` environment variable or the default path `$HOME/.kube/config`) and directly switch to the new context:

```bash
k3d get kubeconfig mycluster --switch
```

Use the new cluster with [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/), e.g.:

```bash
kubectl get nodes
```
