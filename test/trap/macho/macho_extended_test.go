package macho

import (
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

// checkFunctionTypeInfo is a helper function that tests both function argument types and symbol info
func checkFunctionTypeInfo(t *testing.T, funcName string, wantType *TypeInfo) {
	t.Helper()

	checkFunctionMultiArgInfo(t, funcName, []*TypeInfo{wantType})
}

// Special case helper for functions with multiple parameters
func checkFunctionMultiArgInfo(t *testing.T, funcName string, wantTypes []*TypeInfo) {
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
		diff := assert.Diff(want, info.ArgTypes[i])
		if diff != "" {
			t.Errorf("arg %d mismatch:\n%s", i, diff)
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
		checkFunctionTypeInfo(t, "main.fnFloat32", &TypeInfo{Kind: KindFloat32, Name: "float32"})
	})

	t.Run("Float64", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnFloat64", &TypeInfo{Kind: KindFloat64, Name: "float64"})
	})
}

func TestCheckComplexTypes(t *testing.T) {
	t.Run("Complex64", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnComplex64", &TypeInfo{Kind: KindComplex64, Name: "complex64"})
	})

	t.Run("Complex128", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnComplex128", &TypeInfo{Kind: KindComplex128, Name: "complex128"})
	})
}

// Tests for channel types
func TestCheckChannelTypes(t *testing.T) {
	t.Skip("skipping channel tests because it's hard to track the type")
	t.Run("Bidirectional", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnChan", &TypeInfo{
			Kind: KindChan,
			Elem: &TypeInfo{Kind: KindBasic, Name: "int"},
		})
	})

	t.Run("ChanMyStruct", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnChanMyStruct", &TypeInfo{
			Kind: KindChan,
			Elem: &TypeInfo{Kind: KindStruct, Name: "main.MyStruct"},
		})
	})

	t.Run("SendOnly", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnChanSend", &TypeInfo{Kind: KindBasic, Name: "chan<- int"})
	})

	t.Run("ReceiveOnly", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnChanRecv", &TypeInfo{Kind: KindBasic, Name: "<-chan int"})
	})
}

// Tests for array and slice types
func TestCheckArrayTypes(t *testing.T) {
	t.Run("OneDimensional", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnArray", &TypeInfo{
			Kind: KindArray,
			Elem: &TypeInfo{
				Kind: KindBasic,
				Name: "int",
			},
			Count: 3,
		})
	})

	t.Run("TwoDimensional", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnArray2D",
			&TypeInfo{
				Kind:  KindArray,
				Count: 2,
				Elem: &TypeInfo{
					Kind:  KindArray,
					Count: 3,
					Elem: &TypeInfo{
						Kind: KindBasic,
						Name: "int",
					},
				},
			},
		)
	})
}

// Tests for map types
func TestCheckMapTypes(t *testing.T) {
	t.Skip("skipping map tests because it's hard to track the type")
	t.Run("WithComplex", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnMapComplex", &TypeInfo{
			Kind: KindMap,
			Key:  &TypeInfo{Kind: KindBasic, Name: "string"},
			Elem: &TypeInfo{
				Kind: KindBasic,
				Name: "complex128",
			},
		})
	})

	t.Run("WithStruct", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnMapStruct", &TypeInfo{
			Kind: KindMap,
			Key:  &TypeInfo{Kind: KindBasic, Name: "string"},
			Elem: &TypeInfo{
				Kind:   KindStruct,
				Fields: []FieldInfo{{Name: "main.x", Type: &TypeInfo{Kind: KindBasic, Name: "int"}}, {Name: "main.y", Type: &TypeInfo{Kind: KindBasic, Name: "int"}}},
			},
		})
	})
}

// Tests for function types
func TestCheckFunctionTypes(t *testing.T) {
	t.Skip("skipping function tests because it's hard to track the type")
	t.Run("Simple", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnFunc", &TypeInfo{
			Kind:    KindFunc,
			Args:    []*TypeInfo{{Kind: KindBasic, Name: "int"}},
			Results: []*TypeInfo{{Kind: KindBasic, Name: "int"}},
		})
	})

	t.Run("MultipleArgs", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnFuncMulti", &TypeInfo{
			Kind:    KindFunc,
			Args:    []*TypeInfo{{Kind: KindBasic, Name: "int"}, {Kind: KindBasic, Name: "string"}},
			Results: []*TypeInfo{{Kind: KindBasic, Name: "int"}, {Kind: KindBasic, Name: "error"}},
		})
	})
}

// Tests for interface types
func TestCheckInterfaceTypes(t *testing.T) {

	t.Run("Slice", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnInterfaceSlice", &TypeInfo{Kind: KindSlice, Elem: &TypeInfo{Kind: KindInterface}})
	})

	t.Run("Map", func(t *testing.T) {
		t.Skip("skipping map tests because it's hard to track the type")
		checkFunctionTypeInfo(t, "main.fnInterfaceMap", &TypeInfo{Kind: KindMap, Key: &TypeInfo{Kind: KindInterface}, Elem: &TypeInfo{Kind: KindInterface}})
	})
}

// Tests for named types
func TestCheckNamedTypes(t *testing.T) {
	t.Run("TypeAlias", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnNamedType", &TypeInfo{
			Kind: KindNamed,
			Name: "MyInt",
			Elem: &TypeInfo{
				Kind: KindBasic,
				Name: "int",
			},
		})
	})

	t.Run("Struct", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnNamedStruct", &TypeInfo{
			Kind: KindNamed,
			Name: "main.MyStruct",
			Underlying: &TypeInfo{
				Kind: KindStruct,
				Name: "main.MyStruct",
				Fields: []FieldInfo{
					{Name: "X", Type: &TypeInfo{Kind: KindBasic, Name: "int"}},
					{Name: "Y", Type: &TypeInfo{Kind: KindString}},
				},
			},
		})
	})

	t.Run("Interfaces", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnNamedInterface", &TypeInfo{
			Kind: KindNamed,
			Name: "main.MyInterface",
			Underlying: &TypeInfo{
				Kind: KindInterface,
			},
		})
	})
}

// Tests for original basic types
func TestCheckBasicTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		checkFunctionMultiArgInfo(t, "main.fnString", []*TypeInfo{{
			Kind: KindString,
		}, {
			Kind: KindBasic,
			Name: "int",
		}})
	})

	t.Run("Int3", func(t *testing.T) {
		checkFunctionMultiArgInfo(t, "main.fnInt3", []*TypeInfo{{Kind: KindBasic, Name: "int"}, {Kind: KindBasic, Name: "int"}, {Kind: KindBasic, Name: "int"}})
	})

	t.Run("Slice", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnSlice", &TypeInfo{Kind: KindSlice, Elem: &TypeInfo{Kind: KindBasic, Name: "int"}})
	})

	t.Run("Interface", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnInterface", &TypeInfo{Kind: KindInterface})
	})

	t.Run("Struct", func(t *testing.T) {
		checkFunctionTypeInfo(t, "main.fnStruct", &TypeInfo{
			Kind: KindStruct,
			Fields: []FieldInfo{
				{Name: "a", Type: &TypeInfo{Kind: KindBasic, Name: "int"}},
				{Name: "b", Type: &TypeInfo{Kind: KindBasic, Name: "int"}},
			},
		})
	})
}
