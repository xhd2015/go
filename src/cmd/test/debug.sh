#!/usr/bin/env bash

set -e

rebuild=false
verbose=false
args=()
gcflags=()
output=
while [[ $# -gt 0 ]];do
    case $1 in
      -a)
         rebuild=true
         shift
      ;;
      -v|--verbose)
        verbose=true
        shift
      ;;
      -gcflags|--gcflags)
      gcflags=("$1" "$2")
      shift 2
      ;;
      -o)
        output=$2
        shift 2
      ;;
      --help|-h)
      cat <<'EOF'
usage: debug.sh <CMD> [OPTIONS]
Options:
   --help,-h       help
   --verbose,-v    show verbose log

Cmd build,debug:
   -a               go build -a
   -gcflags  FLAGS  
   --gcflags FLAGS  go build -gcflags
   -o file          go build -o
            
  
Cmd build-compiler:

Example:
    $ debug.sh build-compiler
    $ debug.sh build
EOF
      shift
      exit
      ;;
      --)
      shift
      args=("${args[@]}" "$@")
      break
      ;;
      *)
      args=("${args[@]}" $1)
      shift
      ;;
    esac
done

shdir=$(cd "$(dirname "$0")" && pwd)
goroot=$(cd "$shdir/../../.." && pwd)
cd "$shdir"

set -- "${args[@]}"

cmd=$1
shift || true

if [[ -z $cmd ]];then
   echo "usage: $(basename "$0") CMD" >&2
   exit 1
fi


function date_log {
   date +"%Y-%m-%d %H:%M:%S"
}

case "$cmd" in
   debug|build)
      build_flags=("${gcflags[@]}" -o "${output:-main.bin}")
      if [[ $rebuild = true ]];then
           build_flags=("${build_flags[@]}" -a)
      fi
      for log in *.log;do
          echo "$(date_log) >>>>>>>BEGIN<<<<<<<<" >> "$log"
      done
      if [[ $verbose = true ]];then
          tail -fn1 compile.log &
          trap "kill -9 $!" EXIT
      fi
      PATH=$goroot/bin:$PATH GOROOT=$goroot go build -toolexec="$shdir/exce_tool $cmd" "${build_flags[@]}" "$@"
      ;;
    build-compiler)
      (
        cd ..
        PATH=$goroot/bin:$PATH GOROOT=$goroot go build -gcflags="all=-N -l" -o ./test/compile-devel ./compile
      )
    ;;
    *)
      echo "unknow command: $cmd" >&2
      exit 1
    ;;
esac