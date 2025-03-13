#!/usr/bin/env bash

project_root=$(git rev-parse --show-toplevel)

export GOROOT=$project_root
export PATH=$GOROOT/bin:$PATH

"$@"