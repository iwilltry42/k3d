#!/bin/bash

RED='\033[1;31m'
GRN='\033[1;32m'
YEL='\033[1;33m'
BLU='\033[1;34m'
WHT='\033[1;37m'
MGT='\033[1;95m'
CYA='\033[1;96m'
END='\033[0m'
BLOCK='\033[1;37m'

PATH=/usr/local/bin:$PATH
export PATH

log() { >&2 printf "${BLOCK}>>>${END} $1\n"; }

info() { log "${BLU}$1${END}"; }
highlight() { log "${MGT}$1${END}"; }

bye() {
  log "${BLU}$1... exiting${END}"
  exit 0
}

warn() { log "${RED}!!! WARNING !!! $1 ${END}"; }

abort() {
  log "${RED}FATAL: $1${END}"
  exit 1
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

failed() {
  if [ -z "$1" ] ; then
    log "${RED}failed!!!${END}"
  else
    log "${RED}$1${END}"
  fi
  abort "test failed"
}

passed() {
  if [ -z "$1" ] ; then
    log "${GRN}done!${END}"
  else
    log "${GRN}$1${END}"
  fi
}

# checks that a URL is available, with an optional error message
check_url() {
  command_exists curl || abort "curl is not installed"
  curl -L --silent -k --output /dev/null --fail "$1"
}

# check_clusters verifies that given clusters are reachable
check_clusters() {
  [ -n "$EXE" ] || abort "EXE is not defined"
  for c in "$@" ; do
    $EXE get kubeconfig "$c" --switch
    if kubectl cluster-info ; then
      passed "cluster $c is reachable"
    else
      warn "could not obtain cluster info for $c. Kubeconfig:\n$(kubectl config view)"
      docker ps -a
      return 1
    fi
  done
  return 0
}

# check_multi_node verifies that a cluster runs with an expected number of nodes
check_multi_node() {
  cluster=$1
  expectedNodeCount=$2
  $EXE get kubeconfig "$cluster" --switch
  nodeCount=$(kubectl get nodes -o=custom-columns=NAME:.metadata.name --no-headers | wc -l)
  if [[ $nodeCount == $expectedNodeCount ]]; then
    passed "cluster $cluster has $expectedNodeCount nodes, as expected"
  else
    warn "cluster $cluster has incorrect number of nodes: $nodeCount != $expectedNodeCount"
    kubectl get nodes -o=custom-columns=NAME:.metadata.name --no-headers
    docker ps -a
    return 1
  fi
  return 0
}

check_registry() {
  check_url "$REGISTRY/v2/_catalog"
}

check_volume_exists() {
  docker volume inspect "$1" >/dev/null 2>&1
}
