package main

import (
	"fmt"
	"runtime"
)

func main() {
	val := add(1, 2, 3)
	fmt.Printf("add(1,2) = %d\n", val)
}

func add(a, b, c int) int {
	// ignore the lint error
	args := runtime.TrapCallerArgs()
	fmt.Printf("args: %v\n", args)
	return a + b
}
