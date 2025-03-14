#!/usr/bin/env bash

set -eo pipefail

# DO NOT MODIFY THIS FILE

project_root=$(git rev-parse --show-toplevel)

cd "$project_root/test/trap/macho"


go build -o __debug_bin_example -gcflags="all=-N -l" ./example

go test -v "$@"