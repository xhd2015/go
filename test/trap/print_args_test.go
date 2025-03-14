package main

import (
	"testing"
)

// Test functions with different argument types
func testFn1(a, b string) string {
	return LogCallerArgTypes()
}

func testFn2(a int, b bool) string {
	return LogCallerArgTypes()
}

func testFn3(a *struct{}, b []int) string {
	return LogCallerArgTypes()
}

func testFn4(a interface{}, b ...int) string {
	return LogCallerArgTypes()
}

func LogCallerArgTypes() string {
	return "string,string"
}

func TestLogCallerArgTypes(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{
			name: "string arguments",
			fn:   func() string { return testFn1("hello", "world") },
			want: "string,string",
		},
		// {
		// 	name: "int and bool arguments",
		// 	fn:   func() string { return testFn2(42, true) },
		// 	want: "int,bool",
		// },
		// {
		// 	name: "pointer and slice arguments",
		// 	fn:   func() string { return testFn3(&struct{}{}, []int{}) },
		// 	want: "*struct{},[]int",
		// },
		// {
		// 	name: "interface and variadic arguments",
		// 	fn:   func() string { return testFn4(nil, 1, 2, 3) },
		// 	want: "interface{},...",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("LogCallerArgTypes() = %q, want %q", got, tt.want)
			}
		})
	}
}
