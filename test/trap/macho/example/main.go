package main

import "fmt"

// ExampleFunction has various parameter types for testing
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
	fmt.Println(s, i, b, nums, dict, data, ptr)
}

// Named types for testing
type MyInt int
type MyStruct struct {
	X int
	Y string
}
type MyInterface interface {
	DoSomething() int
}

// Implement MyInterface
type MyInterfaceImpl struct {
	Value int
}

func (m MyInterfaceImpl) DoSomething() int {
	return m.Value
}

func main() {
	str := "hello"
	ExampleFunction(
		"test string",
		42,
		true,
		[]int{1, 2, 3},
		map[string]int{"a": 1, "b": 2},
		struct {
			Name string
			Age  int
		}{
			Name: "John",
			Age:  30,
		},
		&str,
	)

	// Call all test functions to ensure they are included in the binary
	fmt.Println(fnString("test", 123))
	fmt.Println(fnInt3(1, 2, 3))
	fmt.Println(fnSlice([]int{1, 2, 3}))
	fmt.Println(fnInterface("test"))
	fmt.Println(fnNamedEmptyInterface(MyEmptyInterface(1)))
	fmt.Println(fnStruct(struct {
		a int
		b int
	}{a: 1, b: 2}))

	// Call extended test functions
	fmt.Println(fnFloat32(1.5))
	fmt.Println(fnFloat64(2.5))
	fmt.Println(fnComplex64(1 + 2i))
	fmt.Println(fnComplex128(3 + 4i))

	ch := make(chan int)
	fmt.Println(fnChan(ch))
	chMyStruct := make(chan MyStruct)
	fmt.Println(fnChanMyStruct(chMyStruct))
	fmt.Println(fnChanSend(ch))
	fmt.Println(fnChanRecv(ch))

	arr := [3]int{1, 2, 3}
	fmt.Println(fnArray(arr))
	arr2d := [2][3]int{{1, 2, 3}, {4, 5, 6}}
	fmt.Println(fnArray2D(arr2d))

	m := map[string]complex128{"a": 1 + 2i}
	fmt.Println(fnMapComplex(m))

	ms := map[string]struct{ x, y int }{"a": {1, 2}}
	fmt.Println(fnMapStruct(ms))

	f := func(x int) int { return x * 2 }
	fmt.Println(fnFunc(f))

	f2 := func(x int, s string) (int, error) { return len(s), nil }
	fmt.Println(fnFuncMulti(f2))

	is := []interface{}{1, "test", true}
	fmt.Println(fnInterfaceSlice(is))

	im := map[interface{}]interface{}{1: "one", "two": 2}
	fmt.Println(fnInterfaceMap(im))

	mi := MyInt(42)
	fmt.Println(fnNamedType(mi))

	ms2 := MyStruct{X: 1, Y: "test"}
	fmt.Println(fnNamedStruct(ms2))

	// Create and use an implementation of MyInterface
	myImpl := MyInterfaceImpl{Value: 42}
	fmt.Println(fnNamedInterface(myImpl))
}

func fnString(a string, b int) string {
	return fmt.Sprintf("%s %d", a, b)
}

func fnInt3(a, b, c int) int {
	return a + b + c
}

func fnSlice(a []int) int {
	return len(a)
}

func fnInterface(a interface{}) int {
	return 0
}

type MyEmptyInterface interface{}

func fnNamedEmptyInterface(a MyEmptyInterface) int {
	return 0
}

func fnStruct(c struct {
	a int
	b int
}) int {
	return c.a + c.b
}

// Extended test functions with various types
func fnFloat32(x float32) float32 {
	return x * 2
}

func fnFloat64(x float64) float64 {
	return x * 2
}

func fnComplex64(x complex64) complex64 {
	return x * 2
}

func fnComplex128(x complex128) complex128 {
	return x * 2
}

func fnChan(x chan int) chan int {
	return x
}

func fnChanMyStruct(x chan MyStruct) chan MyStruct {
	return x
}

func fnChanSend(x chan<- int) chan<- int {
	return x
}

func fnChanRecv(x <-chan int) <-chan int {
	return x
}

func fnArray(x [3]int) [3]int {
	return x
}

func fnArray2D(x [2][3]int) [2][3]int {
	return x
}

func fnMapComplex(x map[string]complex128) map[string]complex128 {
	return x
}

func fnMapStruct(x map[string]struct{ x, y int }) map[string]struct{ x, y int } {
	return x
}

func fnFunc(x func(int) int) func(int) int {
	return x
}

func fnFuncMulti(x func(int, string) (int, error)) func(int, string) (int, error) {
	return x
}

func fnInterfaceSlice(x []interface{}) []interface{} {
	return x
}

func fnInterfaceMap(x map[interface{}]interface{}) map[interface{}]interface{} {
	return x
}

func fnNamedType(x MyInt) MyInt {
	return x
}

func fnNamedStruct(x MyStruct) MyStruct {
	return x
}

func fnNamedInterface(x MyInterface) MyInterface {
	return x
}
