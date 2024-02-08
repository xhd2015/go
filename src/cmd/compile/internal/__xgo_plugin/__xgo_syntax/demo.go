package __xgo_syntax

import (
	"cmd/compile/internal/syntax"
	"os"
)

func demoteRewriteFunc(f *syntax.File) {
	if os.Getenv("COMPILER_ALLOW_SYNTAX_REWRITE") != "true" {
		return
	}
	if f.PkgName.Value != "main" {
		return
	}
	for _, dec := range f.DeclList {
		fn, ok := dec.(*syntax.FuncDecl)
		if !ok {
			continue
		}
		if fn.Body == nil {
			continue
		}
		if fn.Name.Value != "main" {
			continue
		}
		list := make([]syntax.Stmt, len(fn.Body.List)+1)
		list[0] = &syntax.ExprStmt{
			X: &syntax.CallExpr{
				Fun: &syntax.SelectorExpr{
					X: &syntax.Name{
						Value: "fmt",
					},
					Sel: &syntax.Name{
						Value: "Printf",
					},
				},
				ArgList: []syntax.Expr{
					&syntax.BasicLit{
						Value: "\"hello Syntax\\n\"",
						Kind:  syntax.StringLit,
					},
				},
			},
		}

		// if not set pos, link fails
		syntax.Inspect(list[0], func(n syntax.Node) bool {
			if n == nil {
				return false
			}
			n.SetPos(syntax.MakePos(fn.Body.Pos().Base(), 1, 1))
			return true
		})
		for i := 0; i < len(fn.Body.List); i++ {
			list[i+1] = fn.Body.List[i]
		}
		fn.Body.List = list
	}
}
