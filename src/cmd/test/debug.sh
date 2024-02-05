#!/usr/bin/env bash

set -e

shdir=$(cd "$(dirname "$0")" && pwd)
cd "$shdir"

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
      for log in *.log;do
          echo "$(date_log) >>>>>>>BEGIN<<<<<<<<" >> "$log"
      done
      tail -fn1 compile.log &
      trap "kill -9 $!" EXIT
      with-go-devel go build -toolexec="$PWD/exce_tool $cmd" -a -o main.bin ./
      ;;
    build-compiler)
      (
        cd ..
        with-go-devel go build -gcflags="all=-N -l" -o ./test/compile-devel ./compile
      )
    ;;
    *)
      echo "unknow command: $cmd" >&2
      exit 1
    ;;
esac