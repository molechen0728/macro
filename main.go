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

	n := macro.Node{
		Smd:       make(map[string]gss.StructMetaData),
		StructMap: map[string]map[string]struct{}{},
	}

	ast.Walk(&n, f)
	fmt.Printf("%v\n", n)

}
