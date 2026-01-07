package ast

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// GoASTSyncer implements the domain.ASTSyncer port.
type GoASTSyncer struct{}

// SyncFromJSON applies a JSON model to the Go DSL source.
func (GoASTSyncer) SyncFromJSON(src, jsonStr string) (string, error) {
	fset, f, err := parseSource(src)
	if err != nil {
		return "", err
	}

	if err := SyncArchitectureFromJSON(f, jsonStr); err != nil {
		return "", err
	}

	return formatFile(fset, f)
}

func parseSource(src string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "dsl.go", src, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, f, nil
}

func formatFile(fset *token.FileSet, f *ast.File) (string, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return "", err
	}
	return buf.String(), nil
}
