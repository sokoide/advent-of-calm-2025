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

// AddNode inserts a new DefineNode call into the Go DSL source.
func (GoASTSyncer) AddNode(src, nodeID, nodeType, name, desc string) (string, error) {
	fset, f, err := parseSource(src)
	if err != nil {
		return "", err
	}

	if err := AddNodeInAST(f, nodeID, nodeType, name, desc); err != nil {
		return "", err
	}

	return formatFile(fset, f)
}

// UpdateNodeProperty updates a node property in the Go DSL source.
func (GoASTSyncer) UpdateNodeProperty(src, nodeID, property, value string) (string, error) {
	fset, f, err := parseSource(src)
	if err != nil {
		return "", err
	}

	if err := UpdateNodePropertyInAST(f, nodeID, property, value); err != nil {
		return "", err
	}

	return formatFile(fset, f)
}

// DeleteNode removes a DefineNode call from the Go DSL source.
func (GoASTSyncer) DeleteNode(src, nodeID string) (string, error) {
	fset, f, err := parseSource(src)
	if err != nil {
		return "", err
	}

	if err := DeleteNodeInAST(f, nodeID); err != nil {
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
