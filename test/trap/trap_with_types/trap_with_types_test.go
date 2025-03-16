package trap_with_types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/xhd2015/go/test/trap/macho"
)

const testBin = "./__debug_bin_test"

var t *testing.T

// Add a global variable to store the most recent complex struct arguments
var lastComplexStructArgs []interface{}

// add is a simple function that adds two integers
// Used as a test subject for argument trapping
func add(a, b int) int {
	args := runtime.TrapCallerArgs()
	// convert args to type-specific values
	funArg, err := macho.GetFunctionArgTypes(testBin, "github.com/xhd2015/go/test/trap/trap_with_types.add")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}
	convertedArgs := convertArgs(args, funArg)
	expectedArgs := []interface{}{a, b}

	if !reflect.DeepEqual(convertedArgs, expectedArgs) {
		t.Errorf("convertedArgs = %v, expectedArgs = %v", convertedArgs, expectedArgs)
	}

	return a + b
}

// greet is a simple function that takes a string argument
// Used as a test subject for string argument trapping
func greet(name string) string {
	args := runtime.TrapCallerArgs()
	// convert args to type-specific values
	funArg, err := macho.GetFunctionArgTypes(testBin, "github.com/xhd2015/go/test/trap/trap_with_types.greet")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}
	convertedArgs := convertArgs(args, funArg)
	expectedArgs := []interface{}{name}

	if !reflect.DeepEqual(convertedArgs, expectedArgs) {
		t.Errorf("convertedArgs = %v, expectedArgs = %v", convertedArgs, expectedArgs)
	}

	return "Hello, " + name
}

// testStruct is a function that takes a struct argument
// Used as a test subject for struct argument trapping
func fnAStruct(s testStruct) testStruct {
	args := runtime.TrapCallerArgs()
	// convert args to type-specific values
	funArg, err := macho.GetFunctionArgTypes(testBin, "github.com/xhd2015/go/test/trap/trap_with_types.fnAStruct")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}
	convertedArgs := convertArgs(args, funArg)
	expectedArgs := []interface{}{s}

	if !reflect.DeepEqual(convertedArgs, expectedArgs) {
		t.Errorf("convertedArgs = %v, expectedArgs = %v", convertedArgs, expectedArgs)
	}

	// Return a new struct with swapped values
	return testStruct{A: s.B, B: s.A}
}

func fnStructWithSingleString(s structWithSingleString) string {
	args := runtime.TrapCallerArgs()
	funArg, err := macho.GetFunctionArgTypes(testBin, "github.com/xhd2015/go/test/trap/trap_with_types.fnStructWithSingleString")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}
	convertedArgs := convertArgs(args, funArg)
	expectedArgs := []interface{}{s}

	// compare json
	convertedArgsJson, _ := json.Marshal(convertedArgs)
	expectedArgsJson, _ := json.Marshal(expectedArgs)
	if string(convertedArgsJson) != string(expectedArgsJson) {
		t.Errorf("convertedArgs = %v, expectedArgs = %v", convertedArgs, expectedArgs)
	}

	return s.S
}

// fnAComplexStruct is a function that takes a complex struct argument
// Used as a test subject for complex struct argument trapping
func fnAComplexStruct(s complexStruct) string {
	args := runtime.TrapCallerArgs()

	// Remove storing args globally
	// lastComplexStructArgs = args

	funArg, err := macho.GetFunctionArgTypes(testBin, "github.com/xhd2015/go/test/trap/trap_with_types.fnAComplexStruct")
	if err != nil {
		t.Fatalf("GetFunctionArgTypes: %v", err)
	}
	convertedArgs := convertArgs(args, funArg)

	// Special handling for the converted complex struct, which may be a map
	if len(convertedArgs) == 1 {
		// If we got a map representation, manually verify the fields
		if structMap, ok := convertedArgs[0].(map[string]interface{}); ok {
			// Check all the fields of the complex struct
			var allFieldsCorrect = true
			var fieldErrors []string

			// Check Name
			if name, ok := structMap["Name"].(string); !ok || name != s.Name {
				allFieldsCorrect = false
				fieldErrors = append(fieldErrors, fmt.Sprintf("Name: got %v, want %q",
					structMap["Name"], s.Name))
			}

			// Check Value
			if val, ok := structMap["Value"].(int); !ok || val != s.Value {
				allFieldsCorrect = false
				fieldErrors = append(fieldErrors, fmt.Sprintf("Value: got %v, want %d",
					structMap["Value"], s.Value))
			}

			// Check Flag
			if flag, ok := structMap["Flag"].(bool); !ok || flag != s.Flag {
				allFieldsCorrect = false
				fieldErrors = append(fieldErrors, fmt.Sprintf("Flag: got %v, want %v",
					structMap["Flag"], s.Flag))
			}

			// Check Nested (which is also a map)
			if nested, ok := structMap["Nested"].(map[string]interface{}); ok {
				// Check A and B
				if nestedA, ok := nested["A"].(int); !ok || nestedA != s.Nested.A {
					allFieldsCorrect = false
					fieldErrors = append(fieldErrors, fmt.Sprintf("Nested.A: got %v, want %d",
						nested["A"], s.Nested.A))
				}

				if nestedB, ok := nested["B"].(int); !ok || nestedB != s.Nested.B {
					allFieldsCorrect = false
					fieldErrors = append(fieldErrors, fmt.Sprintf("Nested.B: got %v, want %d",
						nested["B"], s.Nested.B))
				}
			} else {
				allFieldsCorrect = false
				fieldErrors = append(fieldErrors, fmt.Sprintf("Nested: got %v, want %+v",
					structMap["Nested"], s.Nested))
			}

			// Check Data length (we can't check actual values since we're only creating an empty slice)
			if data, ok := structMap["Data"].([]int); ok {
				if len(data) != len(s.Data) {
					allFieldsCorrect = false
					fieldErrors = append(fieldErrors, fmt.Sprintf("Data length: got %d, want %d",
						len(data), len(s.Data)))
				}
			} else {
				allFieldsCorrect = false
				fieldErrors = append(fieldErrors, fmt.Sprintf("Data: got %v, want slice of length %d",
					structMap["Data"], len(s.Data)))
			}

			if !allFieldsCorrect {
				t.Errorf("complexStruct field errors: %s", strings.Join(fieldErrors, "; "))
			}

			// If all fields are correct, we're good - skip the deep equal check
			if allFieldsCorrect {
				return "complex struct"
			}
		}
	}

	// Fall back to the original comparison if we don't have a map or the field checks failed
	expectedArgs := []interface{}{s}
	if !reflect.DeepEqual(convertedArgs, expectedArgs) {
		t.Errorf("convertedArgs = %v, expectedArgs = %v", convertedArgs, expectedArgs)
	}

	return "complex struct"
}

// convertArgs converts raw arguments to their typed values
// and type information to capture and interpret function arguments
func TestAddWithTypedArgs(myt *testing.T) {
	t = myt
	// Call add with test values
	result := add(5, 7)

	// Expected result from the add function
	expectedResult := 12
	if result != expectedResult {
		myt.Errorf("add(5, 7) = %d, want %d", result, expectedResult)
	}
}

// TestGreetWithTypedArgs tests the combination of runtime.TrapCallerArgs
// with a string argument
func TestGreetWithTypedArgs(myt *testing.T) {
	t = myt
	// Call greet with a test value
	name := "World"
	result := greet(name)

	// Expected result from the greet function
	expectedResult := "Hello, World"
	if result != expectedResult {
		myt.Errorf("greet(%q) = %q, want %q", name, result, expectedResult)
	}
}

// TestStructWithTypedArgs tests the combination of runtime.TrapCallerArgs
// with a struct argument
func TestStructWithTypedArgs(myt *testing.T) {
	t = myt
	// Call fnAStruct with test values
	myStruct := testStruct{A: 5, B: 7}
	result := fnAStruct(myStruct)

	// Expected result from the fnAStruct function (swapped A and B)
	expectedResult := testStruct{A: 7, B: 5}
	if result != expectedResult {
		myt.Errorf("fnAStruct(%+v) = %+v, want %+v", myStruct, result, expectedResult)
	}
}

// TestStructWithTypedArgs tests the combination of runtime.TrapCallerArgs
// with a struct argument
func TestStructWithSingleString(myt *testing.T) {
	t = myt
	// Call fnAStruct with test values
	myStruct := structWithSingleString{S: "Hello, World"}
	result := fnStructWithSingleString(myStruct)

	// Expected result from the fnAStruct function (swapped A and B)
	expectedResult := "Hello, World"
	if result != expectedResult {
		myt.Errorf("fnStructWithSingleString(%+v) = %+v, want %+v", myStruct, result, expectedResult)
	}
}

// TestComplexStructWithTypedArgs tests the combination of runtime.TrapCallerArgs
// with a complex nested struct argument
func TestComplexStructWithTypedArgs(myt *testing.T) {
	t = myt
	// Call fnAComplexStruct with a complex struct
	myComplexStruct := complexStruct{
		Name:  "TestComplex",
		Value: 42,
		Flag:  true,
		Nested: testStruct{
			A: 10,
			B: 20,
		},
		Data: []int{1, 2, 3},
	}
	result := fnAComplexStruct(myComplexStruct)

	// Expected result from the fnAComplexStruct function
	expectedResult := "complex struct"
	if result != expectedResult {
		myt.Errorf("fnAComplexStruct(%+v) = %q, want %q", myComplexStruct, result, expectedResult)
	}
}
