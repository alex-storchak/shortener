package main

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "panicexitcheck",
	Doc:  "check for using function panic() or log.Fatal()/os.Exit() outside main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			checkPanic(call, pass)
			checkLogFatal(call, file, pass)
			checkOsExit(call, file, pass)

			return true
		})
	}
	return nil, nil
}

func checkPanic(call *ast.CallExpr, pass *analysis.Pass) {
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
		pass.Reportf(call.Pos(), "panic function used in expression")
	}
}

func checkLogFatal(call *ast.CallExpr, file *ast.File, pass *analysis.Pass) {
	p, f := calledPkgAndFuncName(call)
	if p != "log" || f != "Fatal" {
		return
	}
	if !isAllowed(call.Pos(), file, pass) {
		pass.Reportf(call.Pos(), "log.Fatal used outside main function of main package")
	}
}

func checkOsExit(call *ast.CallExpr, file *ast.File, pass *analysis.Pass) {
	p, f := calledPkgAndFuncName(call)
	if p != "os" || f != "Exit" {
		return
	}
	if !isAllowed(call.Pos(), file, pass) {
		pass.Reportf(call.Pos(), "os.Exit used outside main function of main package")
	}
}

func calledPkgAndFuncName(call *ast.CallExpr) (string, string) {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if p, ok := sel.X.(*ast.Ident); ok {
			return p.Name, sel.Sel.Name
		}
	}
	return "", ""
}

func isAllowed(pos token.Pos, f *ast.File, pass *analysis.Pass) bool {
	if ok := isMainPackage(pass); !ok {
		return false
	}
	if ok := isMainFunc(pos, f); !ok {
		return false
	}
	return true
}

// isMainPackage определяет, является ли текущий пакет пакетом main.
func isMainPackage(pass *analysis.Pass) bool {
	if pass.Pkg != nil && pass.Pkg.Name() == "main" {
		return true
	}
	return false
}

func isMainFunc(pos token.Pos, f *ast.File) bool {
	enclosing := enclosingFuncDecl(f, pos)
	if enclosing == nil {
		return false
	}
	if enclosing.Name == nil || enclosing.Name.Name != "main" {
		return false
	}
	return true
}

func enclosingFuncDecl(file *ast.File, pos token.Pos) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if fn.Body.Pos() <= pos && pos <= fn.Body.End() {
			return fn
		}
	}
	return nil
}
