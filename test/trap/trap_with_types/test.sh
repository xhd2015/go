#!/usr/bin/env bash

set -eo pipefail

# DO NOT MODIFY THIS FILE

project_root=$(git rev-parse --show-toplevel)

cd "$project_root/test/trap/trap_with_types"

../with-go-devel.sh go test -c -o __debug_bin_test -gcflags="all=-N -l" ./

args=("$@")
n=${#args[@]}
for((i=0;i<n;i++)); do
    # if starts with -, then replace with -test.
    if [[ ${args[$i]} == -* ]]; then
        args[$i]="-test.${args[$i]:1}"
    fi
done

./__debug_bin_test "${args[@]}"