package main

import (
	"reflect"
	"runtime"
	"testing"
)

func testType() {

}

func TestGetFuncName(t *testing.T) {
	pc := reflect.ValueOf(testType).Pointer()

	fn := runtime.FuncForPC(pc)
	t.Logf("fn: %v", fn.Name())
}

func TestGetPCType(t *testing.T) {

	reflect.TypeOf(testType)
}
