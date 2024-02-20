#!/usr/bin/env bash

set -e

rebuild=false
verbose=false
args=()
gcflags=()
output=
cwd=
mod=
debug=
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
      -mod|-mod=*)
      if [[ $1 = '-mod='* ]];then
         mod=${1#'-mod='}
         shift
      else
         mod=$2
         shift 2
      fi
      ;;
      --debug|--debug=*)
      if [[ $1 = '-debug='* ]];then
         debug=${1#'-debug='}
         shift
      else
         debug=$2
         shift 2
      fi
      ;;
     -w|--cwd)
        cwd=$2
        shift 2
        ;;
      -o)
        output=$2
        shift 2
      ;;
      --help|-h)
      cat <<'EOF'
Usage: debug.sh <CMD> [OPTIONS]
Available commands: build-compiler,build,debug,gen-runtime-type
Options:
   --help,-h       help
   -w, --cwd DIR   working directory
   --verbose,-v    show verbose log

Cmd build,debug:
   -a               go build -a
   -gcflags  FLAGS  
   --gcflags FLAGS  go build -gcflags
   -o file          go build -o
   -mod MOD         go build -mod
   --debug  PKG     debug pkg
  
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

if [[ -n $cwd ]];then
    cd "$cwd"
fi

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
      if [[ -n $mod ]];then
        build_flags=("${build_flags[@]}" -mod "$mod")
      fi
      (
        cd "$shdir"
        for log in *.log;do
          echo "$(date_log) >>>>>>>BEGIN<<<<<<<<" >> "$log"
        done
      )

      if [[ $verbose = true ]];then
          tail -fn1 "$shdir/compile.log" &
          trap "kill -9 $!" EXIT
      fi
      PATH=$goroot/bin:$PATH GOROOT=$goroot DEBUG_PKG=$debug go build -toolexec="$shdir/exce_tool $cmd" "${build_flags[@]}" "$@"
      ;;
    build-compiler)
      (
        cd "$shdir/.."
        PATH=$goroot/bin:$PATH GOROOT=$goroot go build -gcflags="all=-N -l" -o "$shdir/test/compile-devel" ./compile
      )
    ;;
    gen-runtime-type)
    (
      cd "$shdir/.."
      ./test/with-go-devel.sh go generate ./compile/internal/typecheck
    )
     
     ;;
    *)
      echo "unknow command: $cmd" >&2
      exit 1
    ;;
esac