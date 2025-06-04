package mainosexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer defines a static analyzer that prohibits direct calls to os.Exit in the main function of the main package.
var Analyzer = &analysis.Analyzer{
	Name:     "mainosexit",
	Doc:      "checks for direct calls to os.Exit in the main function of the main package",
	Requires: []*analysis.Analyzer{
		// No additional dependencies required
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	for _, file := range pass.Files {
		// Skip non-Go files (from .cache/go-build)
		if !strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), ".go") {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if funcDecl.Name.Name != "main" {
					return true
				}
				ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
					if callExpr, ok := n.(*ast.CallExpr); ok {
						if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := selExpr.X.(*ast.Ident); ok {
								if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
									pass.Reportf(callExpr.Pos(), "direct call to os.Exit is prohibited in the main function of the main package")
								}
							}
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}
