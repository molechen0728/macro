package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"gss"
	"macro/macro"
)

func main() {

	f, err := parser.ParseFile(token.NewFileSet(), "test.go", nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		panic(err)
	}

	a := macro.AtVisitor{
		Smd:       make(map[string]gss.StructMetaData),
		StructMap: map[string]map[string]struct{}{},
	}

	ast.Walk(&a, f)
	fmt.Printf("%v\n", a)

	g := macro.GenErrorRetrunVisitor{}

	ast.Walk(&g, f)
	fmt.Println(g)

	g.Replace("test.go", "test.test.go")
}
