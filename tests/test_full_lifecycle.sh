#!/bin/bash

CURR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
[ -d "$CURR_DIR" ] || { echo "FATAL: no current dir (maybe running in zsh?)";  exit 1; }

# shellcheck source=./common.sh
source "$CURR_DIR/common.sh"

clustername="lifecycletest"

info "Creating cluster $clustername..."
$EXE create cluster "$clustername" --workers 1 --api-port 6443 --wait --timeout 360s || failed "could not create cluster $clustername"

info "Sleeping for 5 seconds to give the cluster enough time to get ready..."
sleep 5

# 1. check initial access to the cluster
info "Checking that we have access to the cluster..."
check_clusters "$clustername" || failed "error checking cluster"

info "Checking that we have 2 nodes online..."
check_multi_node "$clustername" 2 || failed "failed to verify number of nodes"

# 2. stop the cluster
info "Stopping cluster..."
$EXE stop cluster "$clustername"

info "Checking that cluster was stopped"
check_clusters "$clustername" && failed "cluster was not stopped, since we still have access"

# 3. start the cluster
info "Starting cluster..."
$EXE start cluster "$clustername"

info "Sleeping for 5 seconds to give the cluster enough time to get ready..."
sleep 5

info "Checking that we have access to the cluster..."
check_clusters "$clustername" || failed "error checking cluster"

info "Checking that we have 2 nodes online..."
check_multi_node "$clustername" 2 || failed "failed to verify number of nodes"

# 4. load an image into the cluster
info "Loading an image into the cluster..."
docker pull nginx:latest > /dev/null
docker tag nginx:latest nginx:local > /dev/null
$EXE load image nginx:local -c $clustername || failed "could not import image in $clustername"

# Cleanup

info "Deleting cluster $clustername..."
$EXE delete cluster "$clustername" || failed "could not delete the cluster $clustername"

exit 0


