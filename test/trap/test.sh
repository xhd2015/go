#!/usr/bin/env bash

set -eo pipefail

# DO NOT MODIFY THIS FILE

project_root=$(git rev-parse --show-toplevel)

function build_go {
    cd "$project_root/src"
    
    bash ./make.bash
}
# not need to rebuild everytime
# (build_go)

cd "$project_root/test/trap"

./with-go-devel.sh go test -c -o __debug_bin_demo -gcflags="all=-N -l" ./

args=("$@")
n=${#args[@]}
for((i=0;i<n;i++)); do
    # if starts with -, then replace with -test.
    if [[ ${args[$i]} == -* ]]; then
        args[$i]="-test.${args[$i]:1}"
    fi
done

./with-go-devel.sh go test -gcflags="all=-N -l" -v "${args[@]}"