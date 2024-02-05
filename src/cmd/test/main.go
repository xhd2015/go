package main

import (
	"fmt"

	"github.com/xhd2015/go/cmd/test/pkg"
)

func main() {
	res := testArgs("a")
	fmt.Printf("res: %v\n", res)
}

func test_DumpIR() {
	after, stop := trap()
	if !stop {
		if after != nil {
			defer after()
		}
		fmt.Printf("hello IR\n")
	}
}

func testArgs(s string) int {
	fmt.Printf("testArgs: %s\n", s)

	num(1).add(2)
	return 1
}

type num int

func (c num) add(b int) {
	fmt.Printf("%d+%d=%d\n", c, b, int(c)+b)
	pkg.Hello("pkg")
}

func trap() (after func(), stop bool) {
	return nil, false
}
