package main

import (
	"fmt"

	"github.com/xhd2015/go/cmd/test/core/trace"
	_ "github.com/xhd2015/go/cmd/test/core/trap/trap_impl"
	"github.com/xhd2015/go/cmd/test/pkg"
)

func init() {
	// mock.Use()
	trace.Use()
}
func main() {
	res := testArgs("a")
	fmt.Printf("res: %v\n", res)
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
