# Usage
Compiler entrance: [../compile/main.go](../compile/main.go)
```bash
./debug.sh build-compiler
./debug.sh exec
```

# Configure of git exclude
```
root=$(git rev-parse --show-toplevel)
mkdir -p "$root/.git/info"
cat >>"$root/.git/info/exclude" <<'EOF'
/src/cmd/test/*.log
/src/cmd/test/compile-devel
/src/cmd/test/*.bin
EOF
```