package macho

import (
	"debug/dwarf"
	"fmt"
	"strings"
)

// parseType converts a dwarf.Type to our TypeInfo representation
func parseType(dw *dwarf.Data, t dwarf.Type, typeCache map[dwarf.Type]*TypeInfo) (*TypeInfo, error) {
	fmt.Printf("parseType %T\n", t)
	// Use type's string representation as cache key
	if cached, ok := typeCache[t]; ok {
		return cached, nil
	}

	// Create a new TypeInfo
	info := &TypeInfo{}
	// Store in cache
	typeCache[t] = info

	switch t := t.(type) {
	case *dwarf.BasicType:
		// Handle basic types with more precision
		info.Kind = KindBasic

		// Check for specific basic types
		switch t.Name {
		case "float32":
			info.Kind = KindFloat32
			info.Name = "float32"
		case "float64":
			info.Kind = KindFloat64
			info.Name = "float64"
		case "complex64":
			info.Kind = KindComplex64
			info.Name = "complex64"
		case "complex128":
			info.Kind = KindComplex128
			info.Name = "complex128"
		default:
			// For other basic types, use the name from DWARF
			info.Name = t.String()
		}
		return info, nil

	case *dwarf.PtrType:
		info.Kind = KindPtr
		elem, err := parseType(dw, t.Type, typeCache)
		if err != nil {
			return nil, err
		}
		info.Elem = elem
		return info, nil

	case *dwarf.ArrayType:
		info.Kind = KindArray
		info.Count = t.Count
		elem, err := parseType(dw, t.Type, typeCache)
		if err != nil {
			return nil, err
		}
		info.Elem = elem
		return info, nil

	case *dwarf.StructType:
		name := t.StructName
		switch name {
		case "string":
			// stringHeader struct{
			//   	Data uintptr
			//   	Len  int
			// }
			// this is actually a string
			info.Kind = KindString
			return info, nil
		case "runtime.iface":
			// is an interface{}
			info.Kind = KindInterface
			// TODO: add fields
			return info, nil
		case "runtime.eface":
			info.Kind = KindInterface
			return info, nil
		}
		// special slice type
		if strings.HasPrefix(name, "[]") {
			if len(t.Field) == 3 &&
				t.Field[0].Name == "array" &&
				t.Field[1].Name == "len" &&
				t.Field[2].Name == "cap" {
				ptrType, ok := t.Field[0].Type.(*dwarf.PtrType)
				if !ok {
					panic(fmt.Sprintf("expected ptr type, got %T", t.Field[0].Type))
				}
				elem, err := parseType(dw, ptrType.Type, typeCache)
				if err != nil {
					return nil, err
				}
				info.Kind = KindSlice
				info.Elem = elem
				return info, nil
			}
		}

		info.Kind = KindStruct
		// the name is just the literal def name
		if !strings.HasPrefix(name, "struct {") {
			info.Name = name
		}

		// Process struct fields
		fields := make([]FieldInfo, 0, len(t.Field))
		for _, f := range t.Field {
			fieldType, err := parseType(dw, f.Type, typeCache)
			if err != nil {
				return nil, err
			}
			fmt.Printf("struct field: %s\n", f.Name)
			// For struct fields, we want to preserve the package prefix
			fields = append(fields, FieldInfo{
				Name: f.Name,
				Type: fieldType,
			})
		}
		info.Fields = fields
		return info, nil

	case *dwarf.TypedefType:
		// For Go types like string, slice, map which are typedefs to structs
		fmt.Printf("TypedefType: %s\n", t.Name)

		if t.Name == "interface {}" {
			// empty interface
			info.Kind = KindInterface
			return info, nil
		}
		if strings.HasPrefix(t.Name, "chan ") {
			// chan type
			ptrType, ok := t.Type.(*dwarf.PtrType)
			if !ok {
				panic(fmt.Sprintf("expected ptr type, got %T", t.Type))
			}
			fmt.Printf("chan type: %s\n", ptrType.Type.String())
			elemType, err := parseType(dw, ptrType.Type, typeCache)
			if err != nil {
				return nil, err
			}
			info.Kind = KindChan
			info.Elem = elemType
			return info, nil
		}
		if strings.HasPrefix(t.Name, "map[") {
			// TODO: map
			// map type
			info.Kind = KindMap
			return info, nil
		}
		if strings.HasPrefix(t.Name, "func(") {
			// TODO: func
			// func type
			// fmt.Printf("func type: %v\n", t.Type)
			elemType, err := parseType(dw, t.Type, typeCache)
			if err != nil {
				return nil, err
			}
			info.Kind = KindFunc
			info.Args = []*TypeInfo{elemType}
			return info, nil
		}

		// Try to identify special Go types
		// Regular named type - preserve the original name
		info.Kind = KindNamed
		info.Name = t.Name

		var underlying *TypeInfo
		if nest1, ok := t.Type.(*dwarf.TypedefType); ok {
			// fmt.Printf("TypedefType nest2: %s %T\n", n.Name, n.Type)
			if nest2, ok := nest1.Type.(*dwarf.TypedefType); ok {
				// fmt.Printf("TypedefType sub2: %s\n", sub.Name)
				switch nest2.Name {
				case "runtime.iface":
					underlying = &TypeInfo{
						Kind: KindInterface,
					}
				case "runtime.eface":
					underlying = &TypeInfo{
						Kind: KindInterface,
					}
				}
			}
		}

		// Parse the underlying type
		if underlying == nil {
			var err error
			underlying, err = parseType(dw, t.Type, typeCache)
			if err != nil {
				return nil, err
			}
		}

		// Regular named type
		info.Underlying = underlying

		return info, nil
	case *dwarf.QualType:
		// Qualified types like const, volatile
		return parseType(dw, t.Type, typeCache)

	case *dwarf.IntType:
		info.Kind = KindBasic
		info.Name = "int"
		return info, nil

	case *dwarf.UintType:
		info.Kind = KindBasic
		info.Name = "uint"
		return info, nil

	case *dwarf.FloatType:
		// Distinguish between float32 and float64
		if t.ByteSize == 4 {
			info.Kind = KindFloat32
			info.Name = "float32"
		} else {
			info.Kind = KindFloat64
			info.Name = "float64"
		}
		return info, nil

	case *dwarf.ComplexType:
		// Distinguish between complex64 and complex128
		if t.ByteSize == 8 {
			info.Kind = KindComplex64
			info.Name = "complex64"
		} else {
			info.Kind = KindComplex128
			info.Name = "complex128"
		}
		return info, nil

	case *dwarf.BoolType:
		info.Kind = KindBasic
		info.Name = "bool"
		return info, nil

	case *dwarf.FuncType:
		// go: func(int) int
		// dwarf: func(int, *int) void
		fmt.Printf("func type: %v\n", t.Name)
		for _, param := range t.ParamType {
			if ptr, ok := param.(*dwarf.PtrType); ok {
				fmt.Printf("param: %s %T %s\n", ptr.Name, ptr, ptr.Type.String())
			} else {
				fmt.Printf("param: %T %s\n", param, param.String())
			}
		}
		fmt.Printf("results: %v\n", t.ReturnType)
		info.Kind = KindFunc
		return info, nil

	case *dwarf.VoidType:
		info.Kind = KindVoid
		return info, nil

	case *dwarf.UnsupportedType:
		info.Kind = KindUnknown
		info.Name = t.String()
		return info, nil

	default:
		// Handle other types
		info.Kind = KindUnknown
		info.Name = t.String()
		return info, nil
	}
}
