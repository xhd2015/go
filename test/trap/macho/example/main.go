package main

import "fmt"

func main() {
	res1 := fnString("hello", 42)
	fmt.Println(res1)

	res2 := fnInt3(1, 2, 3)
	fmt.Println(res2)

	res3 := fnSlice([]int{1, 2, 3})
	fmt.Println(res3)

	res4 := fnInterface(1)
	fmt.Println(res4)

	res5 := fnStruct(struct {
		a int
		b int
	}{a: 1, b: 2})
	fmt.Println(res5)
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

func fnStruct(c struct {
	a int
	b int
}) int {
	return c.a + c.b
}
