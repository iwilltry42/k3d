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
    command -v $1 >/dev/null 2>&1
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

check_k3d_clusters() {
  [ -n "$EXE" ] || abort "EXE is not defined"
  for c in "c1" "c2" ; do
    kc=$($EXE get kubeconfig $c)
    [ -n "$kc" ] || abort "could not obtain a kubeconfig for $c"
    if kubectl --kubeconfig="$kc" cluster-info ; then
      passed "cluster $c is reachable (with kubeconfig=$kc)"
    else
      warn "could not obtain cluster info for $c (with kubeconfig=$kc). Contents:\n$(cat $kc)"
      return 1
    fi
  done
  return 0
}

check_registry() {
  check_url $REGISTRY/v2/_catalog
}

