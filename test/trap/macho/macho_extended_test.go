package main

import (
	"testing"
)

// checkFunctionTypeInfo is a helper function that tests both function argument types and symbol info
func checkFunctionTypeInfo(t *testing.T, funcName string, wantType string) {
	t.Helper()

	// Get function argument types
	info, err := GetFunctionArgTypes(testBin, funcName)
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}

	// Check function name
	if info.Name != funcName {
		t.Errorf("wrong function name: got %q, want %q", info.Name, funcName)
	}

	// Check argument type
	if len(info.ArgTypes) != 1 {
		t.Errorf("wrong number of args: got %d, want 1", len(info.ArgTypes))
	} else if info.ArgTypes[0].String() != wantType {
		t.Errorf("arg 0: got %q, want %q", info.ArgTypes[0].String(), wantType)
	}

	// Check symbol info
	sym, err := GetFunctionSymbolInfo(testBin, funcName)
	if err != nil {
		t.Fatalf("GetFunctionSymbolInfo: %v", err)
	}

	if sym.Name != funcName {
		t.Errorf("wrong symbol name: got %q, want %q", sym.Name, funcName)
	}
}

// Special case helper for functions with multiple parameters
func checkFunctionMultiArgInfo(t *testing.T, funcName string, wantTypes []string) {
	t.Helper()

	// Get function argument types
	info, err := GetFunctionArgTypes(testBin, funcName)
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}

	// Check function name
	if info.Name != funcName {
		t.Errorf("wrong function name: got %q, want %q", info.Name, funcName)
	}

	// Check argument types
	if len(info.ArgTypes) != len(wantTypes) {
		t.Errorf("wrong number of args: got %d, want %d", len(info.ArgTypes), len(wantTypes))
	}
	for i, want := range wantTypes {
		if i >= len(info.ArgTypes) {
			break
		}
		if info.ArgTypes[i].String() != want {
			t.Errorf("arg %d: got %q, want %q", i, info.ArgTypes[i].String(), want)
		}
	}

	// Check symbol info
	sym, err := GetFunctionSymbolInfo(testBin, funcName)
	if err != nil {
		t.Fatalf("GetFunctionSymbolInfo: %v", err)
	}

	if sym.Name != funcName {
		t.Errorf("wrong symbol name: got %q, want %q", sym.Name, funcName)
	}
}

// Tests for numeric types
func TestCheckFloatTypes(t *testing.T) {
	t.Run("Float32", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnFloat32", "float32")
	})

	t.Run("Float64", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnFloat64", "float64")
	})
}

func TestCheckComplexTypes(t *testing.T) {
	t.Run("Complex64", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnComplex64", "complex64")
	})

	t.Run("Complex128", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnComplex128", "complex128")
	})
}

// Tests for channel types
func TestCheckChannelTypes(t *testing.T) {
	t.Run("Bidirectional", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnChan", "chan int")
	})

	t.Run("SendOnly", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnChanSend", "chan<- int")
	})

	t.Run("ReceiveOnly", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnChanRecv", "<-chan int")
	})
}

// Tests for array and slice types
func TestCheckArrayTypes(t *testing.T) {
	t.Run("OneDimensional", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnArray", "[3]int")
	})

	t.Run("TwoDimensional", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnArray2D", "[2][3]int")
	})
}

// Tests for map types
func TestCheckMapTypes(t *testing.T) {
	t.Run("WithComplex", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnMapComplex", "map[string]complex128")
	})

	t.Run("WithStruct", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnMapStruct", "map[string]struct { main.x int; main.y int }")
	})
}

// Tests for function types
func TestCheckFunctionTypes(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnFunc", "func(int) int")
	})

	t.Run("MultipleArgs", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnFuncMulti", "func(int, string) (int, error)")
	})
}

// Tests for interface types
func TestCheckInterfaceTypes(t *testing.T) {
	t.Run("Slice", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnInterfaceSlice", "[]interface {}")
	})

	t.Run("Map", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnInterfaceMap", "map[interface {}]interface {}")
	})
}

// Tests for named types
func TestCheckNamedTypes(t *testing.T) {
	t.Run("TypeAlias", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnNamedType", "int")
	})

	t.Run("Struct", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnNamedStruct", "main.MyStruct")
	})

	t.Run("Interface", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnNamedInterface", "main.MyInterface")
	})
}

// Tests for original basic types
func TestCheckBasicTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		checkFunctionMultiArgInfo(t, "main.fnString", []string{"string", "int"})
	})

	t.Run("Int3", func(t *testing.T) {
		checkFunctionMultiArgInfo(t, "main.fnInt3", []string{"int", "int", "int"})
	})

	t.Run("Slice", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnSlice", "[]int")
	})

	t.Run("Interface", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnInterface", "interface {}")
	})

	t.Run("Struct", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnStruct", "struct { main.a int; main.b int }")
	})
}
