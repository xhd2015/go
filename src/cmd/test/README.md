# Play with go compiler
This directory demonstrates how to play with the go compiler.
It modifies the compiler so that it inserts two extra print:
 - by modifying AST: hello Syntax
 - by modifying IR: hello IR
 - by modifying IR and inserting __x_trap to runtime: hello Trap

```sh
$ ./debug.sh build-comipler
$ ./debug.sh build
$ ./main.bin 
hello IR
hello Trap
hello Syntax
hello world
```

# Debug the compiler
Compiler entrance: [../compile/main.go](../compile/main.go)

```sh
./debug.sh build-compiler
./debug.sh debug # this will hang the terminal, and you can copy the output configuration to .vscode
```

# How to add customized function to runtime?
1.Edit [../compile/internal/typecheck/_builtin/runtime.go](../compile/internal/typecheck/_builtin/runtime.go) to add function declaration,
2.Execute go generate
```sh
./with-go-devel.sh go generate ../compile/internal/typecheck
```

# Check runtime symbols
```sh
./with-go-devel.sh go tool nm runtime.a
```

# Configure git exclude
```
root=$(git rev-parse --show-toplevel)
mkdir -p "$root/.git/info"
cat >>"$root/.git/info/exclude" <<'EOF'
/src/cmd/test/*.log
/src/cmd/test/compile-devel
/src/cmd/test/*.bin
/src/cmd/test/*.a
EOF
```