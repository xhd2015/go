package __xgo_syntax

import (
	"cmd/compile/internal/syntax"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var files []*syntax.File

func SetFiles(f []*syntax.File) {
	files = f
}

func GetFiles() []*syntax.File {
	return files
}

func AfterFilesParsed(fileList []*syntax.File, addFile func(name string, r io.Reader)) {
	if len(fileList) == 0 {
		return
	}
	files = fileList
	pkgName := fileList[0].PkgName.Value
	if pkgName == "runtime" {
		return
	}
	// cannot directly import the runtime package
	// but we can first:
	//  1.modify the importcfg
	//  2.do not import anything, rely on IR to finish remaining steps
	//
	// I feel the second is more proper as importcfg is an extra layer of
	// complexity, and runtime can be compiled or cached, we cannot locate
	// where its _pkg_.a is.
	body := getRegFuncsBody(fileList)
	autoGen :=
		"package " + pkgName + "\n" +
			"func __xgo_register_funcs(__xgo_reg_func func(fn interface{}, recvName string, argNames []string, resNames []string)){\n" +
			body +
			"\n}"
	// ioutil.WriteFile("test.log", []byte(autoGen), 0755)
	addFile("__xgo_autogen.go", strings.NewReader(autoGen))
}

type declName struct {
	name         string
	recvTypeName string
	recvPtr      bool

	// arg names
	recvName string
	argNames []string
	resNames []string
}

func (c *declName) RefName() string {
	if c.recvTypeName == "" {
		return c.name
	}
	if c.recvPtr {
		return fmt.Sprintf("(*%s).%s", c.recvTypeName, c.name)
	}
	return c.recvTypeName + "." + c.name
}
func getRegFuncsBody(files []*syntax.File) string {
	var declFuncNames []*declName
	for _, f := range files {
		for _, decl := range f.DeclList {
			fn, ok := decl.(*syntax.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Value == "init" {
				continue
			}
			var recvTypeName string
			var recvPtr bool
			var recvName string
			if fn.Recv != nil {
				recvName = "_"
				if fn.Recv.Name != nil {
					recvName = fn.Recv.Name.Value
				}
				if starExpr, ok := fn.Recv.Type.(*syntax.Operation); ok && starExpr.Op == syntax.Mul {
					recvTypeName = starExpr.X.(*syntax.Name).Value
					recvPtr = true
				} else {
					recvTypeName = fn.Recv.Type.(*syntax.Name).Value
				}
			}

			declFuncNames = append(declFuncNames, &declName{
				name:         fn.Name.Value,
				recvTypeName: recvTypeName,
				recvPtr:      recvPtr,
				recvName:     recvName,
				argNames:     getFieldNames(fn.Type.ParamList),
				resNames:     getFieldNames(fn.Type.ResultList),
			})
		}
	}

	stmts := make([]string, 0, len(declFuncNames))
	for _, declName := range declFuncNames {
		stmts = append(stmts, fmt.Sprintf("__xgo_reg_func(%s,%s,%s,%s)", declName.RefName(), strconv.Quote(declName.recvName), quoteNamesExpr(declName.argNames), quoteNamesExpr(declName.resNames)))
	}
	return strings.Join(stmts, "\n")
}

func getFieldNames(x []*syntax.Field) []string {
	names := make([]string, 0, len(x))
	for _, p := range x {
		var name string
		if p.Name != nil {
			name = p.Name.Value
		}
		names = append(names, name)
	}
	return names
}

func quoteNamesExpr(names []string) string {
	if len(names) == 0 {
		return "nil"
	}
	qNames := make([]string, 0, len(names))
	for _, name := range names {
		qNames = append(qNames, strconv.Quote(name))
	}
	return "[]string{" + strings.Join(qNames, ",") + "}"
}
