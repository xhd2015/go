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

# ./with-go-devel.sh go build -o __debug_bin_demo -gcflags="all=-N -l" ./

./with-go-devel.sh go test -gcflags="all=-N -l" -v "$@"