package main

import (
	"debug/dwarf"
	"debug/macho"
	"fmt"
)

// analyse tools:
//
//  (cd test/trap/macho && go tool nm __debug_bin_example | grep fnString)
//
// (cd test/trap/macho && go tool objdump -S __debug_bin_example | grep -A 10
// fnString)

// FuncArgInfo represents information about a function's arguments
type FuncArgInfo struct {
	Name     string
	ArgTypes []string
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
			if name != funcName {
				reader.SkipChildren()
				continue
			}

			fmt.Printf("Found function: %s\n", name) // Debug log
			// Found our function
			var argTypes []string

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
						argTypes = append(argTypes, param.String())
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
					typeStr := paramType.String()
					// Strip "struct " prefix if present for strings and slices
					if len(typeStr) > 7 && typeStr[:7] == "struct " {
						typeStr = typeStr[7:]
					}
					// Clean up struct field names by removing package prefix and fixing spaces
					if len(typeStr) > 8 && typeStr[:8] == "struct {" {
						// Find all occurrences of "main." and remove them, also normalize spaces
						var result []byte
						i := 0
						for i < len(typeStr) {
							if i+5 <= len(typeStr) && typeStr[i:i+5] == "main." {
								i += 5
							} else if typeStr[i] == '{' {
								result = append(result, '{')
								i++
								// Skip space after opening brace
								if i < len(typeStr) && typeStr[i] == ' ' {
									i++
								}
							} else if typeStr[i] == ';' {
								result = append(result, ';')
								result = append(result, ' ')
								i++
								// Skip any existing space after semicolon
								if i < len(typeStr) && typeStr[i] == ' ' {
									i++
								}
							} else if typeStr[i] == '}' {
								// Remove space before closing brace
								if len(result) > 0 && result[len(result)-1] == ' ' {
									result = result[:len(result)-1]
								}
								result = append(result, '}')
								i++
							} else {
								result = append(result, typeStr[i])
								i++
							}
						}
						typeStr = string(result)
					}
					fmt.Printf("Found parameter from child: %s\n", typeStr)
					argTypes = append(argTypes, typeStr)
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
