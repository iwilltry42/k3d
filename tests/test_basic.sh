#!/bin/bash

CURR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
[ -d "$CURR_DIR" ] || { echo "FATAL: no current dir (maybe running in zsh?)";  exit 1; }

# shellcheck source=./common.sh
source "$CURR_DIR/common.sh"

info "Creating two clusters..."
$EXE --verbose create cluster c1 --wait 60 --api-port 6443 || failed "could not create cluster c1"
$EXE --verbose create cluster c2 --wait 60 --api-port 6444 || failed "could not create cluster c2"

info "Checking we have access to both clusters..."
check_k3d_clusters "c1" "c2" || failed "error checking cluster"

info "Deleting clusters..."
$EXE --verbose delete cluster c1 || failed "could not delete the cluster c1"
$EXE --verbose delete cluster c2 || failed "could not delete the cluster c2"

exit 0


