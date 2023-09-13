package utils

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func StringToAst(file string) (*ast.File, *token.FileSet, error) {

	fs := token.NewFileSet()
	p, err := parser.ParseFile(fs, file, nil, parser.AllErrors)
	if err != nil {
		return nil, nil, err
	}

	return p, fs, nil
}
