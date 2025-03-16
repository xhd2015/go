package macho

import (
	"testing"
)

const testBin = "./__debug_bin_example"

func TestCheckFnString(t *testing.T) {
	checkFunctionMultiArgInfo(t, "main.fnString", []*TypeInfo{
		{Kind: KindString},
		{Kind: KindBasic, Name: "int"},
	})
}

func TestCheckFnInt3(t *testing.T) {
	checkFunctionMultiArgInfo(t, "main.fnInt3", []*TypeInfo{
		{Kind: KindBasic, Name: "int"},
		{Kind: KindBasic, Name: "int"},
		{Kind: KindBasic, Name: "int"},
	})
}

func TestCheckFnSlice(t *testing.T) {
	checkFunctionTypeInfo(t, "main.fnSlice", &TypeInfo{
		Kind: KindSlice,
		Elem: &TypeInfo{
			Kind: KindBasic,
			Name: "int",
		},
	})
}

func TestCheckFnInterface(t *testing.T) {
	checkFunctionTypeInfo(t, "main.fnInterface", &TypeInfo{
		Kind: KindInterface,
	})
}
func TestCheckFnNamedEmptyInterface(t *testing.T) {
	checkFunctionTypeInfo(t, "main.fnNamedEmptyInterface", &TypeInfo{
		Kind:       KindNamed,
		Name:       "main.MyEmptyInterface",
		Underlying: &TypeInfo{Kind: KindInterface},
	})
}

func TestCheckFnUnamedStruct(t *testing.T) {
	checkFunctionTypeInfo(t, "main.fnStruct", &TypeInfo{
		Kind: KindStruct,
		Fields: []FieldInfo{
			{Name: "a", Type: &TypeInfo{Kind: KindBasic, Name: "int"}},
			{Name: "b", Type: &TypeInfo{Kind: KindBasic, Name: "int"}},
		},
	})
}
