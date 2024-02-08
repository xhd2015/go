package main

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/xhd2015/go/cmd/test/core/trace"
	_ "github.com/xhd2015/go/cmd/test/core/trap/trap_impl"
	"github.com/xhd2015/go/cmd/test/pkg"
	"github.com/xhd2015/go/cmd/test/test/test_trap/mock"
)

func init() {
	trace.Use()

	// trap.AddInterceptor(trap.Interceptor{
	// 	Pre: func(ctx context.Context, f *trap.FuncInfo, args *trap.FuncArgs) (data interface{}, err error) {
	// 		if strings.Contains(f.FullName, "testReflect") {
	// 			return nil, nil
	// 		}
	// 		return nil, nil
	// 	},
	// })
}

// can break some
func regTest() {
}

func main() {
	mock.CheckSSA()
	// v := reflect.ValueOf(0)
	// v.Call()
	runtime.TestModuleData_Requires_Xgo()
	res := testArgs("a")
	fmt.Printf("res: %v\n", res)
}

// GOSSAFUNC=main.checkSSA ./with-go-devel.sh go build -gcflags="all=-N -l" ./test/test_trap
func checkSSA() {
	var v interface{} = testReflect
	_ = v
	// fmt.Println(testReflect)
}

func getReflectWord(i interface{}) uintptr {
	type IHeader struct {
		typ  uintptr
		word uintptr
	}

	return (*IHeader)(unsafe.Pointer(&i)).word
}

func testReflect() {
	pc := runtime.Getcallerpc()
	entryPC := runtime.GetcallerFuncPC()

	fmt.Printf("testReflect caller pc: %x\n", pc)
	fmt.Printf("testReflect caller entry pc: %x\n", entryPC)
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
