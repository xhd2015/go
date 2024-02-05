# Play with go compiler
This directory demonstrates how to play with the go compiler.
It modifies the compiler so that it inserts two extra print:
 - by modifying AST:  hello Syntax
 - by modifying IR:   hello IR 

```sh
$ ./debug.sh build-comipler
$ ./debug.sh build
$ ./main.bin 
hello IR
hello Syntax
hello world
```

# Debug the compiler
Compiler entrance: [../compile/main.go](../compile/main.go)

```bash
./debug.sh build-compiler
./debug.sh debug # this will hang the terminal, and you can copy the output configuration to .vscode
```

# Configure git exclude
```
root=$(git rev-parse --show-toplevel)
mkdir -p "$root/.git/info"
cat >>"$root/.git/info/exclude" <<'EOF'
/src/cmd/test/*.log
/src/cmd/test/compile-devel
/src/cmd/test/*.bin
EOF
```