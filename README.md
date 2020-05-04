# k3d

[![Build Status](https://travis-ci.com/rancher/k3d.svg?branch=master)](https://travis-ci.com/rancher/k3d)
[![Go Report Card](https://goreportcard.com/badge/github.com/rancher/k3d)](https://goreportcard.com/report/github.com/rancher/k3d)

## k3s in docker

k3s is the lightweight Kubernetes distribution by Rancher: [rancher/k3s](https://github.com/rancher/k3s)

k3d creates containerized k3s clusters. This means, that you can spin up a multi-node k3s cluster on a single machine using docker.

This repository is based on [@zeerorg](https://github.com/zeerorg/)'s [zeerorg/k3s-in-docker](https://github.com/zeerorg/k3s-in-docker), reimplemented in Go by [@iwilltry42](https://github.com/iwilltry42/) in [iwilltry42/k3d](https://github.com/iwilltry42/k3d), which got adopted by Rancher in[rancher/k3d](https://github.com/rancher/k3d).

## Requirements

- [docker](https://docs.docker.com/install/)

## Get

You have several options there:

- via brew (homebrew/linuxbrew): `brew install k3d`
- use the install script to grab the latest release:
  - wget: `wget -q -O - https://raw.githubusercontent.com/rancher/k3d/master/install.sh | bash`
  - curl: `curl -s https://raw.githubusercontent.com/rancher/k3d/master/install.sh | bash`
- Grab a release from the [release tab](https://github.com/rancher/k3d/releases) and install it yourself.
- Via go: `go install github.com/rancher/k3d` (**Note**: this will give you unreleased/bleeding-edge changes)

or...

## Build

1. Clone this repo, e.g. via `go get -u github.com/rancher/k3d`
2. Inside the repo run
   - 'make install-tools' to make sure required go packages are installed
3. Inside the repo run one of the following commands
   - `make build` to build for your current system
   - `go install` to install it to your `GOPATH` (**Note**: this will give you unreleased/bleeding-edge changes)
   - `make build-cross` to build for all systems

## Usage

Check out what you can do via `k3d help`

Example Workflow: Create a new cluster and use it with `kubectl`

1. `k3d create cluster CLUSTER_NAME` to create a new single-node cluster (= 1 container running k3s)
2. `export KUBECONFIG=$(k3d get kubeconfig CLUSTER_NAME)` to make `kubectl` to use the kubeconfig for that cluster
3. execute some commands like `kubectl get pods --all-namespaces`
4. `k3d delete cluster CLUSTER_NAME` to delete the default cluster

### Exposing Services

If you want to access your services from the outside (e.g. via Ingress), you need to map the ports (e.g. port 80 for Ingress) using the `--publish` flag (or aliases).
Check out the [examples here](docs/examples.md).

## What now?

### More Information

- [Command Tree](docs/commands.md)
- [Default Behavior](docs/defaults.md)
- [Usage examples](docs/examples.md)
- [Frequently asked questions and nice-to-know facts](docs/faq.md)
- [Design/Architecture Thoughts, Plans, Unsorted Ideas and more](thoughts.md)

### Connect

1. Join the Rancher community on slack via [slack.rancher.io](https://slack.rancher.io/)
2. Go to [rancher-users.slack.com](https://rancher-users.slack.com) and join our channel #k3d
3. Start chatting
