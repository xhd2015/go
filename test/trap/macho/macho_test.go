package main

import (
	"testing"
)

const testBin = "./__debug_bin_example"

func TestCheckFnString(t *testing.T) {
	// Test GetFunctionArgTypes
	info, err := GetFunctionArgTypes(testBin, "main.fnString")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}

	// Verify function info
	if info.Name != "main.fnString" {
		t.Errorf("wrong function name: got %q, want %q", info.Name, "main.fnString")
	}

	wantTypes := []string{"string", "int"}
	if len(info.ArgTypes) != len(wantTypes) {
		t.Errorf("wrong number of args: got %d, want %d", len(info.ArgTypes), len(wantTypes))
	}
	for i, want := range wantTypes {
		if i >= len(info.ArgTypes) {
			break
		}
		if info.ArgTypes[i] != want {
			t.Errorf("arg %d: got %q, want %q", i, info.ArgTypes[i], want)
		}
	}

	// Test GetFunctionSymbolInfo
	sym, err := GetFunctionSymbolInfo(testBin, "main.fnString")
	if err != nil {
		t.Fatalf("GetFunctionSymbolInfo: %v", err)
	}

	if sym.Name != "main.fnString" {
		t.Errorf("wrong symbol name: got %q, want %q", sym.Name, "main.fnString")
	}
}

func TestCheckFnInt3(t *testing.T) {
	// Test GetFunctionArgTypes
	info, err := GetFunctionArgTypes(testBin, "main.fnInt3")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}

	// Verify function info
	if info.Name != "main.fnInt3" {
		t.Errorf("wrong function name: got %q, want %q", info.Name, "main.fnInt3")
	}

	wantTypes := []string{"int", "int", "int"}
	if len(info.ArgTypes) != len(wantTypes) {
		t.Errorf("wrong number of args: got %d, want %d", len(info.ArgTypes), len(wantTypes))
	}
	for i, want := range wantTypes {
		if i >= len(info.ArgTypes) {
			break
		}
		if info.ArgTypes[i] != want {
			t.Errorf("arg %d: got %q, want %q", i, info.ArgTypes[i], want)
		}
	}

	// Test GetFunctionSymbolInfo
	sym, err := GetFunctionSymbolInfo(testBin, "main.fnInt3")
	if err != nil {
		t.Fatalf("GetFunctionSymbolInfo: %v", err)
	}

	if sym.Name != "main.fnInt3" {
		t.Errorf("wrong symbol name: got %q, want %q", sym.Name, "main.fnInt3")
	}
}

func TestCheckFnSlice(t *testing.T) {
	// Test GetFunctionArgTypes
	info, err := GetFunctionArgTypes(testBin, "main.fnSlice")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}

	// Verify function info
	if info.Name != "main.fnSlice" {
		t.Errorf("wrong function name: got %q, want %q", info.Name, "main.fnSlice")
	}

	wantTypes := []string{"[]int"}
	if len(info.ArgTypes) != len(wantTypes) {
		t.Errorf("wrong number of args: got %d, want %d", len(info.ArgTypes), len(wantTypes))
	}
	for i, want := range wantTypes {
		if i >= len(info.ArgTypes) {
			break
		}
		if info.ArgTypes[i] != want {
			t.Errorf("arg %d: got %q, want %q", i, info.ArgTypes[i], want)
		}
	}

	// Test GetFunctionSymbolInfo
	sym, err := GetFunctionSymbolInfo(testBin, "main.fnSlice")
	if err != nil {
		t.Fatalf("GetFunctionSymbolInfo: %v", err)
	}

	if sym.Name != "main.fnSlice" {
		t.Errorf("wrong symbol name: got %q, want %q", sym.Name, "main.fnSlice")
	}
}

func TestCheckFnInterface(t *testing.T) {
	// Test GetFunctionArgTypes
	info, err := GetFunctionArgTypes(testBin, "main.fnInterface")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}

	// Verify function info
	if info.Name != "main.fnInterface" {
		t.Errorf("wrong function name: got %q, want %q", info.Name, "main.fnInterface")
	}

	wantTypes := []string{"interface {}"}
	if len(info.ArgTypes) != len(wantTypes) {
		t.Errorf("wrong number of args: got %d, want %d", len(info.ArgTypes), len(wantTypes))
	}
	for i, want := range wantTypes {
		if i >= len(info.ArgTypes) {
			break
		}
		if info.ArgTypes[i] != want {
			t.Errorf("arg %d: got %q, want %q", i, info.ArgTypes[i], want)
		}
	}

	// Test GetFunctionSymbolInfo
	sym, err := GetFunctionSymbolInfo(testBin, "main.fnInterface")
	if err != nil {
		t.Fatalf("GetFunctionSymbolInfo: %v", err)
	}

	if sym.Name != "main.fnInterface" {
		t.Errorf("wrong symbol name: got %q, want %q", sym.Name, "main.fnInterface")
	}
}

func TestCheckFnStruct(t *testing.T) {
	// Test GetFunctionArgTypes
	info, err := GetFunctionArgTypes(testBin, "main.fnStruct")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}

	// Verify function info
	if info.Name != "main.fnStruct" {
		t.Errorf("wrong function name: got %q, want %q", info.Name, "main.fnStruct")
	}

	wantTypes := []string{"struct {a int; b int}"}
	if len(info.ArgTypes) != len(wantTypes) {
		t.Errorf("wrong number of args: got %d, want %d", len(info.ArgTypes), len(wantTypes))
	}
	for i, want := range wantTypes {
		if i >= len(info.ArgTypes) {
			break
		}
		if info.ArgTypes[i] != want {
			t.Errorf("arg %d: got %q, want %q", i, info.ArgTypes[i], want)
		}
	}

	// Test GetFunctionSymbolInfo
	sym, err := GetFunctionSymbolInfo(testBin, "main.fnStruct")
	if err != nil {
		t.Fatalf("GetFunctionSymbolInfo: %v", err)
	}

	if sym.Name != "main.fnStruct" {
		t.Errorf("wrong symbol name: got %q, want %q", sym.Name, "main.fnStruct")
	}
}
