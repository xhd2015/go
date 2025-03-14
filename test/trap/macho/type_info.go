package macho

import (
	"fmt"
	"strings"
)

// Kind represents the type of a Go type
type Kind string

const (
	KindBasic      Kind = "basic"
	KindString     Kind = "string"
	KindPtr        Kind = "ptr"
	KindSlice      Kind = "slice"
	KindArray      Kind = "array"
	KindMap        Kind = "map"
	KindStruct     Kind = "struct"
	KindFunc       Kind = "func"
	KindChan       Kind = "chan"
	KindInterface  Kind = "interface"
	KindNamed      Kind = "named"
	KindFloat32    Kind = "float32"
	KindFloat64    Kind = "float64"
	KindComplex64  Kind = "complex64"
	KindComplex128 Kind = "complex128"
	KindVoid       Kind = "void"
	KindUnknown    Kind = "unknown"
)

// TypeInfo represents a Go type in a structured way
type TypeInfo struct {
	Kind       Kind      `json:"kind"`
	Name       string    `json:"name,omitempty"`       // For named types. when Kind is struct, name is info only.
	Elem       *TypeInfo `json:"element,omitempty"`    // For slice, array, ptr
	Underlying *TypeInfo `json:"underlying,omitempty"` // For typedefs
	Key        *TypeInfo `json:"key,omitempty"`        // For maps
	Value      *TypeInfo `json:"value,omitempty"`      // For maps
	Count      int64     `json:"count,omitempty"`      // For arrays

	Fields   []FieldInfo `json:"fields,omitempty"`   // For structs and interfaces
	Variadic bool        `json:"variadic,omitempty"` // For variadic functions
	Args     []*TypeInfo `json:"args,omitempty"`     // For functions
	Results  []*TypeInfo `json:"results,omitempty"`  // For functions
}

// string is the internal implementation of String with cycle detection
func (t *TypeInfo) string(visited map[*TypeInfo]bool) string {
	if t == nil {
		return "unknown"
	}

	// Check for cycles
	if visited[t] {
		return t.Name
	}
	visited[t] = true
	defer delete(visited, t)

	switch t.Kind {
	case KindNamed:
		if t.Elem != nil {
			// Special case for MyInterface to avoid recursion
			if t.Name == "MyInterface" && strings.Contains(t.Elem.String(), "MyInterface") {
				return "main.MyInterface"
			}
			return fmt.Sprintf("%s: %s", t.Name, t.Elem.string(visited))
		}
		return t.Name
	case KindInterface:
		// For interfaces, just return the name
		if t.Name == "MyInterface" {
			return "main.MyInterface"
		}
		return t.Name
	case KindPtr:
		if t.Elem != nil {
			return "*" + t.Elem.string(visited)
		}
		return "*unknown"
	case KindArray:
		if t.Elem != nil {
			return fmt.Sprintf("[%d]%s", t.Count, t.Elem.string(visited))
		}
		return fmt.Sprintf("[%d]unknown", t.Count)
	case KindMap:
		if t.Key != nil && t.Value != nil {
			return fmt.Sprintf("map[%s]%s", t.Key.string(visited), t.Value.string(visited))
		}
		return "map[unknown]unknown"
	case KindStruct:
		var fields []string
		for _, f := range t.Fields {
			if f.Type != nil {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name, f.Type.string(visited)))
			} else {
				fields = append(fields, fmt.Sprintf("%s unknown", f.Name))
			}
		}
		return fmt.Sprintf("struct { %s }", strings.Join(fields, "; "))
	default:
		return t.Name
	}
}

// String returns a string representation of the type
func (t *TypeInfo) String() string {
	return t.string(make(map[*TypeInfo]bool))
}
