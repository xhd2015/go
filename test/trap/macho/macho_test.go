package main

import (
	"testing"
)

const testBin = "./__debug_bin_example"

func TestCheckFnString(t *testing.T) {
	checkFunctionMultiArgInfo(t, "main.fnString", []string{"string", "int"})
}

func TestCheckFnInt3(t *testing.T) {
	checkFunctionMultiArgInfo(t, "main.fnInt3", []string{"int", "int", "int"})
}

func TestCheckFnSlice(t *testing.T) {
	checkFunctionTypeInfo(t, "main.fnSlice", "[]int")
}

func TestCheckFnInterface(t *testing.T) {
	checkFunctionTypeInfo(t, "main.fnInterface", "interface {}")
}

func TestCheckFnStruct(t *testing.T) {
	checkFunctionTypeInfo(t, "main.fnStruct", "struct { main.a int; main.b int }")
}
