#!/bin/bash

K3D_EXE=${EXE:-/bin/k3d}
K3D_IMAGE_TAG=$1

# define E2E_KEEP to non-empty for keeping the e2e runner container after running the tests
E2E_KEEP=${E2E_KEEP:-}

####################################################################################

TIMESTAMP=$(date "+%m%d%H%M%S")

k3de2e=$(docker run -d \
          -v "$(pwd)"/tests:/tests \
          --privileged \
          -e EXE="$K3D_EXE" \
          -e CI="true" \
          -e LOG_LEVEL="$LOG_LEVEL" \
          --name "k3d-e2e-runner-$TIMESTAMP" \
          k3d:$K3D_IMAGE_TAG)

sleep 5 # wait 5 seconds for docker to start

# Execute tests
finish() {
  docker stop "$k3de2e" || /bin/true
  if [ -z "$E2E_KEEP" ] ; then
    docker rm "$k3de2e" || /bin/true
  fi
}
trap finish EXIT

docker exec "$k3de2e" /tests/runner.sh
