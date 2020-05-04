# Networking

- Related issues:
    - [rancher/k3d #220](https://github.com/rancher/k3d/issues/220)

## Introduction

By default, k3d creates a new (docker) network for every new cluster.
Using the `--network STRING` flag upon creation to connect to an existing network.
Existing networks won't be managed by k3d together with the cluster lifecycle.

## Connecting to docker "internal"/pre-defined networks

### `host` network

When using the `--network` flag to connect to the host network (i.e. `k3d create cluster --network host`),
you won't be able to create more than **one master node**.
An edge case would be one master node (with agent disabled) and one worker node.

### `bridge` network

By default, every network that k3d creates is working in `bridge` mode.
But when you try to use `--network bridge` to connect to docker's internal `bridge` network, you may
run into issues with grabbing certificates from the API-Server. Single-Node clusters should work though.

### `none` "network"

Well.. this doesn't really make sense for k3d anyway ¯\_(ツ)_/¯
