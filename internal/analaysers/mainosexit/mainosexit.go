// Analyzer checks that there are
// no calls of os.Exit in package main.
package mainosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NewCheckAnalayser - return a new object instance.
func NewCheckAnalayser() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "mainosexit",
		Doc:  "calling os.Exit in main package is not allowed",
		Run:  run,
	}
}

//nolint:nilnil
func run(pass *analysis.Pass) (interface{}, error) {
	pname, fname := "main", "main"
	if pass.Pkg.Name() != pname {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fun, ok := decl.(*ast.FuncDecl)
			if !ok || fun.Name.Name != fname {
				continue
			}

			astInspect(pass, fun)
		}
	}

	return nil, nil
}

func astInspect(
	pass *analysis.Pass,
	decl *ast.FuncDecl,
) {
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		call, isok := n.(*ast.CallExpr)
		if !isok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok {
			ident, ok := sel.X.(*ast.Ident)
			if ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(),
					"calling os.Exit in main package is not allowed")
			}
		}

		return true
	})
}
