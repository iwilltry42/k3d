#!/bin/bash

CURR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
[ -d "$CURR_DIR" ] || { echo "FATAL: no current dir (maybe running in zsh?)";  exit 1; }

# shellcheck source=./common.sh
source "$CURR_DIR/common.sh"

info "Creating cluster multimaster..."
$EXE create cluster "multimaster" --masters 3 --api-port 6443 --wait --timeout 360s || failed "could not create cluster multimaster"

info "Checking that we have access to the cluster..."
check_clusters "multimaster" || failed "error checking cluster"

info "Sleeping for 5 seconds to give the cluster enough time to get ready..."
sleep 5

info "Checking that we have 3 master nodes online..."
check_multi_node "multimaster" 3 || failed "failed to verify number of nodes"

info "Deleting cluster multimaster..."
$EXE delete cluster "multimaster" || failed "could not delete the cluster multimaster"

exit 0


