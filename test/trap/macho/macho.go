package main

import (
	"debug/dwarf"
	"debug/macho"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// analyse tools:
//
//  (cd test/trap/macho && go tool nm __debug_bin_example | grep fnString)
//
// (cd test/trap/macho && go tool objdump -S __debug_bin_example | grep -A 10
// fnString)

type Kind string

const (
	KindBasic       Kind = "basic"
	KindBool        Kind = "bool"
	KindInt         Kind = "int"
	KindUint        Kind = "uint"
	KindInt8        Kind = "int8"
	KindInt16       Kind = "int16"
	KindInt32       Kind = "int32"
	KindInt64       Kind = "int64"
	KindUint8       Kind = "uint8"
	KindUint16      Kind = "uint16"
	KindUint32      Kind = "uint32"
	KindUint64      Kind = "uint64"
	KindString      Kind = "string"
	KindSlice       Kind = "slice"
	KindFloat32     Kind = "float32"
	KindFloat64     Kind = "float64"
	KindComplex64   Kind = "complex64"
	KindComplex128  Kind = "complex128"
	KindArray       Kind = "array"
	KindMap         Kind = "map"
	KindPtr         Kind = "ptr"
	KindStruct      Kind = "struct"
	KindFunc        Kind = "func"
	KindVoid        Kind = "void"
	KindUnsupported Kind = "unsupported"
	KindUnknown     Kind = "unknown"
	KindNamed       Kind = "named"
)

// TypeInfo represents a Go type in a structured way
type TypeInfo struct {
	Kind   Kind        `json:"kind"`
	Name   string      `json:"name,omitempty"`    // For named types
	Elem   *TypeInfo   `json:"element,omitempty"` // For slice, array, ptr
	Key    *TypeInfo   `json:"key,omitempty"`     // For maps
	Fields []FieldInfo `json:"fields,omitempty"`  // For structs
	Count  int64       `json:"count,omitempty"`   // For arrays
}

// FieldInfo represents a field in a struct
type FieldInfo struct {
	Name string    `json:"name"`
	Type *TypeInfo `json:"type"`
}

// FuncArgInfo represents information about a function's arguments
type FuncArgInfo struct {
	Name     string
	ArgTypes []TypeInfo
}

// String returns the string representation of TypeInfo
func (t TypeInfo) String() string {
	switch t.Kind {
	case KindBasic:
		return t.Name
	case KindString:
		return "string"
	case KindStruct:
		if t.Name != "" {
			return t.Name
		}
		fields := make([]string, 0, len(t.Fields))
		for _, f := range t.Fields {
			// Remove package prefix (like "main.") from field names
			fieldName := f.Name
			if dotIdx := strings.LastIndex(fieldName, "."); dotIdx != -1 {
				fieldName = fieldName[dotIdx+1:]
			}
			fields = append(fields, fieldName+" "+f.Type.String())
		}
		// Fix spacing in struct output to match expected format
		return "struct {" + strings.Join(fields, "; ") + "}"
	case KindFunc:
		return "func"
	case KindVoid:
		return "void"
	case KindUnsupported:
		return "unsupported"
	case KindUnknown:
		return "unknown"
	case KindNamed:
		return t.Name
	case KindPtr:
		return "*" + t.Elem.String()
	case KindSlice:
		return "[]" + t.Elem.String()
	case KindArray:
		return fmt.Sprintf("[%d]%s", t.Count, t.Elem.String())
	case KindMap:
		return fmt.Sprintf("map[%s]%s", t.Key.String(), t.Elem.String())
	case KindFloat32:
		return "float32"
	case KindFloat64:
		return "float64"
	case KindComplex64:
		return "complex64"
	case KindComplex128:
		return "complex128"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindUint:
		return "uint"
	default:
		return t.Name
	}
}

// parseType converts a dwarf.Type to our TypeInfo representation
func parseType(dw *dwarf.Data, t dwarf.Type, typeCache map[string]*TypeInfo) (*TypeInfo, error) {
	// Use type's string representation as cache key
	typeName := t.String()
	if cached, ok := typeCache[typeName]; ok {
		return cached, nil
	}

	// Create a new TypeInfo
	info := &TypeInfo{}
	// Store in cache
	typeCache[typeName] = info

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
		info.Kind = KindStruct
		info.Name = t.StructName

		// Process struct fields
		fields := make([]FieldInfo, 0, len(t.Field))
		for _, f := range t.Field {
			fieldType, err := parseType(dw, f.Type, typeCache)
			if err != nil {
				return nil, err
			}
			fields = append(fields, FieldInfo{
				Name: f.Name,
				Type: fieldType,
			})
		}
		info.Fields = fields
		return info, nil

	case *dwarf.TypedefType:
		// For Go types like string, slice, map which are typedefs to structs
		name := t.String()

		// Try to identify special Go types
		if name == "string" {
			info.Kind = KindString
			return info, nil
		} else if strings.HasPrefix(name, "[]") { // slice
			info.Kind = KindSlice
			elem, err := parseType(dw, t.Type, typeCache)
			if err != nil {
				return nil, err
			}
			// Special handling for slices: the actual element type is usually inside the struct
			if elem.Kind == "struct" && len(elem.Fields) >= 3 {
				// Try to extract the element type from the underlying struct
				// Typical slice struct has data, len, cap fields
				dataField := elem.Fields[0]
				if dataField.Type != nil && dataField.Type.Kind == "ptr" && dataField.Type.Elem != nil {
					info.Elem = dataField.Type.Elem
				} else {
					// Fallback
					info.Elem = &TypeInfo{Kind: "unknown"}
				}
			} else {
				info.Elem = elem
			}
			return info, nil
		} else if strings.HasPrefix(name, "map[") { // map
			info.Kind = KindMap

			// Maps are complex in DWARF, so we'll try to extract key and value types from the name
			mapStr := name[4:] // Skip "map["
			bracketIdx := strings.Index(mapStr, "]")
			if bracketIdx > 0 {
				keyStr := mapStr[:bracketIdx]
				valueStr := mapStr[bracketIdx+1:]

				info.Key = &TypeInfo{Kind: "basic", Name: keyStr}
				info.Elem = &TypeInfo{Kind: "basic", Name: valueStr}
			} else {
				// Fallback for complex map types
				info.Key = &TypeInfo{Kind: "unknown"}
				info.Elem = &TypeInfo{Kind: "unknown"}
			}
			return info, nil
		} else if name == "complex64" {
			info.Kind = KindComplex64
			info.Name = "complex64"
			return info, nil
		} else if name == "complex128" {
			info.Kind = KindComplex128
			info.Name = "complex128"
			return info, nil
		} else {
			// Regular named type - preserve the original name
			info.Kind = KindNamed
			info.Name = name

			// For named types like "main.MyInt", we want to preserve the full name
			// rather than just using the underlying type
			return info, nil
		}

	case *dwarf.QualType:
		// Qualified types like const, volatile
		return parseType(dw, t.Type, typeCache)

	case *dwarf.IntType:
		info.Kind = KindInt
		info.Name = t.String()
		return info, nil

	case *dwarf.UintType:
		info.Kind = KindUint
		info.Name = t.String()
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
		info.Kind = KindBool
		return info, nil

	case *dwarf.FuncType:
		info.Kind = KindFunc
		return info, nil

	case *dwarf.VoidType:
		info.Kind = KindVoid
		return info, nil

	case *dwarf.UnsupportedType:
		info.Kind = KindUnsupported
		info.Name = t.String()
		return info, nil

	default:
		// Handle other types
		info.Kind = KindUnknown
		info.Name = t.String()
		return info, nil
	}
}

// GetFunctionArgTypes reads function argument types from a Mach-O binary using DWARF debug info
func GetFunctionArgTypes(binaryPath string, funcName string) (*FuncArgInfo, error) {
	// Open the Mach-O file
	f, err := macho.Open(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("open macho: %v", err)
	}
	defer f.Close()

	// Get DWARF data
	dw, err := f.DWARF()
	if err != nil {
		return nil, fmt.Errorf("get dwarf: %v", err)
	}

	// Type cache to handle recursive types
	typeCache := make(map[string]*TypeInfo)

	// Extract the function name without package prefix if it has one
	shortFuncName := funcName
	if lastDot := strings.LastIndex(funcName, "."); lastDot != -1 {
		shortFuncName = funcName[lastDot+1:]
	}

	fmt.Printf("Searching for function: %s (short name: %s)\n", funcName, shortFuncName)

	// Read function info from DWARF
	reader := dw.Reader()
	for {
		entry, err := reader.Next()
		if err != nil {
			return nil, fmt.Errorf("read dwarf entry: %v", err)
		}
		if entry == nil {
			break
		}

		// Look for function entries
		if entry.Tag == dwarf.TagSubprogram {
			name, ok := entry.Val(dwarf.AttrName).(string)
			if !ok {
				continue
			}

			// Check for both the full name and the short name
			if name != funcName && name != shortFuncName {
				reader.SkipChildren()
				continue
			}

			fmt.Printf("Found matching function: %s\n", name) // Debug log
			// Found our function
			var argTypes []TypeInfo

			// First try to get parameters from function type
			if typeOffset, ok := entry.Val(dwarf.AttrType).(dwarf.Offset); ok {
				funcType, err := dw.Type(typeOffset)
				if err != nil {
					return nil, fmt.Errorf("get function type: %v", err)
				}

				// Type assert to FuncType
				if ft, ok := funcType.(*dwarf.FuncType); ok {
					fmt.Printf("Found function type with %d parameters\n", len(ft.ParamType))
					for i, param := range ft.ParamType {
						fmt.Printf("Parameter %d: %s\n", i, param.String())

						// Parse the parameter type
						typeInfo, err := parseType(dw, param, typeCache)
						if err != nil {
							return nil, fmt.Errorf("parse parameter type: %v", err)
						}
						argTypes = append(argTypes, *typeInfo)
					}
					return &FuncArgInfo{
						Name:     name,
						ArgTypes: argTypes,
					}, nil
				}
			}

			// If function type approach didn't work, try reading parameters from child entries
			fmt.Printf("Trying child entries approach\n")
			var foundReturnValue bool
			for {
				child, err := reader.Next()
				if err != nil {
					return nil, fmt.Errorf("read parameter entry: %v", err)
				}
				if child == nil || child.Tag == 0 {
					break
				}
				if child.Tag != dwarf.TagFormalParameter {
					continue
				}

				// Debug: print all attributes
				fmt.Printf("Parameter attributes:\n")
				for _, f := range child.Field {
					fmt.Printf("  %v: %v\n", f.Attr, f.Val)
				}

				// Check if this is a return value parameter
				isReturnValue := false
				if varParam, ok := child.Val(dwarf.AttrVarParam).(bool); ok && varParam {
					// Return values have VarParam set to true
					isReturnValue = true
				} else if name, ok := child.Val(dwarf.AttrName).(string); ok && name[0] == '~' {
					// Return values also have names starting with ~
					isReturnValue = true
				}

				if isReturnValue {
					fmt.Printf("Found return value parameter, stopping parameter collection\n")
					foundReturnValue = true
					continue
				}

				// Skip parameters after we've found a return value
				if foundReturnValue {
					continue
				}

				if typeOffset, ok := child.Val(dwarf.AttrType).(dwarf.Offset); ok {
					paramType, err := dw.Type(typeOffset)
					if err != nil {
						return nil, fmt.Errorf("get parameter type: %v", err)
					}

					// Parse the detailed type information
					typeInfo, err := parseType(dw, paramType, typeCache)
					if err != nil {
						return nil, fmt.Errorf("parse parameter type: %v", err)
					}

					fmt.Printf("Found parameter from child: %s\n", typeInfo.String())
					argTypes = append(argTypes, *typeInfo)
				}
			}

			if len(argTypes) > 0 {
				fmt.Printf("Found %d parameters from child entries\n", len(argTypes))
				return &FuncArgInfo{
					Name:     name,
					ArgTypes: argTypes,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("function %s not found", funcName)
}

// TODO: this is just for demo, don't call it
func listAllFunctions(dw *dwarf.Data) {
	// Try to list all available functions for debugging
	fmt.Println("Available functions in DWARF:")
	reader := dw.Reader()
	for {
		entry, err := reader.Next()
		if err != nil {
			break
		}
		if entry == nil {
			break
		}
		if entry.Tag == dwarf.TagSubprogram {
			name, ok := entry.Val(dwarf.AttrName).(string)
			if ok {
				fmt.Printf("  %s\n", name)
			}
		}
	}
}

// GetFunctionSymbolInfo gets symbol information for a function from the Mach-O binary
func GetFunctionSymbolInfo(binaryPath string, funcName string) (*macho.Symbol, error) {
	f, err := macho.Open(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("open macho: %v", err)
	}
	defer f.Close()

	// Check symbol table
	if f.Symtab == nil {
		return nil, fmt.Errorf("no symbol table found")
	}

	// Look for function in symbol table
	for _, sym := range f.Symtab.Syms {
		if sym.Name == funcName {
			return &sym, nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found", funcName)
}

// Example function with various parameter types for testing
func ExampleFunction(
	s string,
	i int,
	b bool,
	nums []int,
	dict map[string]int,
	data struct {
		Name string
		Age  int
	},
	ptr *string,
) {
	// Function body not important - this is just for DWARF info
	fmt.Println(s, i, b, nums, dict, data, ptr)
}

func macho_main() {
	// Example usage
	if len(os.Args) < 3 {
		fmt.Println("Usage: macho <binary_path> <function_name>")
		os.Exit(1)
	}

	binaryPath := os.Args[1]
	funcName := os.Args[2]

	// Get detailed function argument info
	funcInfo, err := GetFunctionArgTypes(binaryPath, funcName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print detailed information
	fmt.Printf("Function: %s\n", funcInfo.Name)
	fmt.Printf("Parameter types:\n")

	for i, argType := range funcInfo.ArgTypes {
		// Format as pretty JSON
		jsonBytes, err := json.MarshalIndent(argType, "", "  ")
		if err != nil {
			fmt.Printf("  Param %d: Error marshaling type\n", i)
			continue
		}
		fmt.Printf("  Param %d: %s\n", i, string(jsonBytes))
	}
}

func main() {
	macho_main()
}
