// This test requires a modified Go compiler with runtime_trap experiment enabled.
// Run with: (cd test/trap && ./with-go-devel.sh go test -gcflags="all=-N -l" -v ./)

package main

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestInt3(t *testing.T) {
	// Test with three integer arguments
	got := fnWithInt3(1, 2, 3)
	fmt.Printf("TestTrapCallerArgs got: %v\n", got)
	want := []interface{}{uint64(1), uint64(2), uint64(3)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("fnWithInt3(1,2,3) captured args = %v, want %v", got, want)
	}
}

// are all uint64s
type stringHeader struct {
	Data uint64
	Len  uint64
}

func TestString(t *testing.T) {
	// Test with string arguments
	var hello = "hello"
	var world = "world"
	got := fnWithString(hello, world)

	// No longer using these
	helloHeader := str2StringHeader(hello)
	worldHeader := str2StringHeader(world)

	fmt.Printf("TestString got: %v\n", got)

	want := []interface{}{
		[]interface{}{helloHeader.Data, helloHeader.Len},
		[]interface{}{worldHeader.Data, worldHeader.Len},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("fnWithString(%q,%q) captured args = %v, want %v", hello, world, got, want)
	}
}

func str2StringHeader(s string) *stringHeader {
	return (*stringHeader)(unsafe.Pointer(&s))
}

// fnWithInt3 is a test function that captures its arguments
func fnWithInt3(a, b, c int) []interface{} {
	fmt.Printf("fnWithInt3 called with: a=%d, b=%d, c=%d\n", a, b, c)
	args := runtime.TrapCallerArgs()
	fmt.Printf("Raw captured args: %#v\n", args)

	return args
}

// fnWithString is a test function that captures string arguments
func fnWithString(a, b string) []interface{} {
	fmt.Printf("fnWithString called with: a=%q, b=%q\n", a, b)
	args := runtime.TrapCallerArgs()
	fmt.Printf("Raw captured string args: %#v\n", args)
	return args
}
