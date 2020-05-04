#!/bin/bash

CURR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
[ -d "$CURR_DIR" ] || { echo "FATAL: no current dir (maybe running in zsh?)";  exit 1; }

# shellcheck source=./common.sh
source "$CURR_DIR/common.sh"

#########################################################################################

[ -n "$EXE" ] || abort "no EXE provided"

info "Preparing filesystem and environment..."

mkdir -p /root/.kube

info "Starting e2e tests..."

for i in $CURR_DIR/test_*.sh ; do
  base=$(basename $i .sh)
  highlight "***** Running $base *****"
  $i || abort "test $base failed"
done

exit 0

