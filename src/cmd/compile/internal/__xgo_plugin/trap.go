package __xgo_plugin

import (
	"cmd/compile/internal/__xgo_plugin/__xgo_syntax"
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"fmt"
	"go/constant"
	"os"
	"strings"
)

var intfSlice *types.Type

func InsertTrapPoints() {
	dumpIR := os.Getenv("COMPILER_DEBUG_IR_DUMP_FUNCS")
	if dumpIR != "" && dumpIR != "false" {
		names := strings.Split(dumpIR, ",")
		for _, fn := range typecheck.Target.Funcs {
			// if strings.Contains(os.Getenv("COMPILER_DEBUG_IR_FUNC"), fn.Sym().Name) {
			for _, name := range names {
				if strings.Contains(fn.Sym().Name, name) {
					ir.Dump("debug:", fn)
					break
				}
			}

		}
		return
	}
	if os.Getenv("COMPILER_ALLOW_IR_REWRITE") != "true" {
		return
	}
	files := __xgo_syntax.GetFiles()
	__xgo_syntax.SetFiles(nil) // help GC

	// check if any file has __XGO_SKIP_TRAP
	var skipTrap bool
	for _, f := range files {
		for _, d := range f.DeclList {
			if d, ok := d.(*syntax.ConstDecl); ok && len(d.NameList) > 0 && d.NameList[0].Value == "__XGO_SKIP_TRAP" {
				skipTrap = true
				break
			}
		}
		if skipTrap {
			break
		}
	}

	// if true {
	// 	return
	// }
	linkMap := map[string]string{
		"__xgo_link_for_each_func": "__xgo_for_each_func",
		"__xgo_link_getcurg":       "__xgo_getcurg",
	}

	intf := types.Types[types.TINTER]
	intfSlice = types.NewSlice(intf)
	// printString := typecheck.LookupRuntime("printstring")
	trap := typecheck.LookupRuntime("__xgo_trap")
	for _, fn := range typecheck.Target.Funcs {
		fnName := fn.Sym().Name
		if fnName == "init" {
			// this init is package level auto generated init, so don't
			// trap this
			continue
		}
		// process link name
		// TODO: what about unnamed closure?
		linkName := linkMap[fnName]
		if linkName != "" {
			// ir.Dump("before:", fn)
			replaceWithRuntimeCall(fn, linkName)
			// ir.Dump("after:", fn)
			continue
		}
		if skipTrap || strings.HasSuffix(fnName, "_xgo_trap_skip") {
			continue
		}
		// fn.Body =
		t := fn.Type()

		afterV := typecheck.TempAt(base.AutogeneratedPos, fn, types.NewSignature(nil, nil, nil))
		stopV := typecheck.TempAt(base.AutogeneratedPos, fn, types.Types[types.TBOOL])

		callTrap := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, trap, []ir.Node{
			takeAddr(t.Recv()),
			takeAddrs(t.Params()),
			takeAddrs(t.Results()),
		})

		assignStmt := ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS2, []ir.Node{afterV, stopV}, []ir.Node{callTrap})
		assignStmt.Def = true

		callAfter := ir.NewIfStmt(base.AutogeneratedPos, ir.NewBinaryExpr(base.AutogeneratedPos, ir.ONE, afterV, ir.NewNilExpr(base.AutogeneratedPos, afterV.Type())), []ir.Node{
			ir.NewGoDeferStmt(base.AutogeneratedPos, ir.ODEFER, ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, afterV, nil)),
		}, nil)

		origBody := fn.Body
		newBody := make([]ir.Node, 1+len(origBody))
		newBody[0] = callAfter
		for i := 0; i < len(origBody); i++ {
			newBody[i+1] = origBody[i]
		}
		ifStmt := ir.NewIfStmt(base.AutogeneratedPos, stopV, nil, newBody)

		fn.Body = []ir.Node{assignStmt, typecheck.Stmt(ifStmt)}
		typecheck.Stmts(fn.Body)
	}

	// if false {
	regFuncs()
	// }
}

func regFuncs() {
	sym, ok := types.LocalPkg.Syms["__xgo_register_funcs"]
	if !ok {
		return
	}
	// TODO: check sym is func, and accepts the following param
	regFunc := typecheck.LookupRuntime("__xgo_register_func")
	node := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, sym.Def.(*ir.Name), []ir.Node{
		regFunc,
	})
	nodes := []ir.Node{node}
	typecheck.Stmts(nodes)
	prependInit(typecheck.Target, nodes)
}

func replaceWithRuntimeCall(fn *ir.Func, name string) {
	runtimeFunc := typecheck.LookupRuntime(name)
	params := fn.Type().Params()
	results := fn.Type().Results()
	paramNames := make([]ir.Node, 0, len(params))
	for _, p := range params {
		paramNames = append(paramNames, p.Nname.(*ir.Name))
	}
	resNames := make([]ir.Node, 0, len(results))
	for _, p := range results {
		resNames = append(resNames, p.Nname.(*ir.Name))
	}
	var callNode ir.Node
	callNode = ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, runtimeFunc, paramNames)
	if len(resNames) > 0 {
		// if len(resNames) == 1 {
		// 	callNode = ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS, resNames, []ir.Node{callNode})
		// } else {
		callNode = ir.NewReturnStmt(base.AutogeneratedPos, []ir.Node{callNode})
		// callNode = ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS2, resNames, []ir.Node{callNode})

		// callNode = ir.NewAssignListStmt(base.AutogeneratedPos, ir.OAS2, resNames, []ir.Node{callNode})
		// }
	}
	var node ir.Node
	node = ir.NewIfStmt(base.AutogeneratedPos,
		ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TBOOL], constant.MakeBool(true)),
		[]ir.Node{
			// ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, typecheck.LookupRuntime("printstring"), []ir.Node{
			// 	ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("debug getg")),
			// }),
			callNode,
		},
		fn.Body,
	)
	savedFunc := ir.CurFunc
	ir.CurFunc = fn
	node = typecheck.Stmt(node)
	ir.CurFunc = savedFunc

	fn.Body = []ir.Node{node}
	// .Prepend(node)
}

func regFuncsV1() {
	files := __xgo_syntax.GetFiles()
	__xgo_syntax.SetFiles(nil) // help GC

	type declName struct {
		name         string
		recvTypeName string
		recvPtr      bool
	}
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
			if fn.Recv != nil {
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
			})
		}
	}

	regFunc := typecheck.LookupRuntime("__xgo_register_func")
	regMethod := typecheck.LookupRuntime("__xgo_register_method")
	_ = regMethod

	var regNodes []ir.Node
	for _, declName := range declFuncNames {
		var valNode ir.Node
		fnSym, ok := types.LocalPkg.LookupOK(declName.name)
		if !ok {
			panic(fmt.Errorf("func name symbol not found: %s", declName.name))
		}
		if declName.recvTypeName != "" {
			typeSym, ok := types.LocalPkg.LookupOK(declName.recvTypeName)
			if !ok {
				panic(fmt.Errorf("type name symbol not found: %s", declName.recvTypeName))
			}
			var recvNode ir.Node
			if !declName.recvPtr {
				recvNode = typeSym.Def.(*ir.Name)
				// recvNode = ir.NewNameAt(base.AutogeneratedPos, typeSym, nil)
			} else {
				// types.TypeSymLookup are for things like "int","func(){...}"
				//
				// typeSym2 := types.TypeSymLookup(declName.recvTypeName)
				// if typeSym2 == nil {
				// 	panic("empty typeSym2")
				// }
				// types.TypeSym()
				recvNode = ir.TypeNode(typeSym.Def.(*ir.Name).Type())
			}
			valNode = ir.NewSelectorExpr(base.AutogeneratedPos, ir.OMETHEXPR, recvNode, fnSym)
			continue
		} else {
			valNode = fnSym.Def.(*ir.Name)
			// valNode = ir.NewNameAt(base.AutogeneratedPos, fnSym, fnSym.Def.Type())
			// continue
		}
		_ = valNode

		node := ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, regFunc, []ir.Node{
			// ir.NewNilExpr(base.AutogeneratedPos, types.AnyType),
			ir.NewConvExpr(base.AutogeneratedPos, ir.OCONV, types.Types[types.TINTER] /*types.AnyType*/, valNode),
			// ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
		})

		// ir.MethodExprFunc()
		regNodes = append(regNodes, node)
	}

	// this typecheck is required
	// to make subsequent steps work
	typecheck.Stmts(regNodes)

	// regFuncs.Body = []ir.Node{
	// 	ir.NewCallExpr(base.AutogeneratedPos, ir.OCALL, typecheck.LookupRuntime("printstring"), []ir.Node{
	// 		ir.NewBasicLit(base.AutogeneratedPos, types.Types[types.TSTRING], constant.MakeString("hello init\n")),
	// 	}),
	// }
	prependInit(typecheck.Target, regNodes)
}

// how to delcare a new function?
// init names are usually init.0, init.1, ...
//
// NOTE: when there is already an init function, declare new init function
// will give an error: main..inittask: relocation target main.init.1 not defined
func prependInit(target *ir.Package, body []ir.Node) {
	if len(target.Inits) > 0 {
		target.Inits[0].Body.Prepend(body...)
		return
	}

	sym := types.LocalPkg.Lookup(fmt.Sprintf("init.%d", len(target.Inits)))
	regFuncs := ir.NewFunc(base.AutogeneratedPos, base.AutogeneratedPos, sym, types.NewSignature(nil, nil, nil))
	regFuncs.Body = body

	target.Inits = append(target.Inits, regFuncs)
	target.Funcs = append(target.Funcs, regFuncs)
}

func takeAddr(recv *types.Field) ir.Expr {
	if recv == nil {
		return ir.NewNilExpr(base.AutogeneratedPos, types.Types[types.TINTER])
	}
	arg := ir.NewAddrExpr(base.AutogeneratedPos, recv.Nname.(*ir.Name))
	conv := ir.NewConvExpr(base.AutogeneratedPos, ir.OCONV, types.Types[types.TINTER], arg)
	conv.SetImplicit(true)
	return conv
}

// take address of all parameters
func takeAddrs(fields []*types.Field) ir.Expr {
	if len(fields) == 0 {
		return ir.NewNilExpr(base.AutogeneratedPos, intfSlice)
	}
	paramList := make([]ir.Node, len(fields))
	for i, f := range fields {
		paramList[i] = takeAddr(f)
	}
	return ir.NewCompLitExpr(base.AutogeneratedPos, ir.OCOMPLIT, intfSlice, paramList)
}
