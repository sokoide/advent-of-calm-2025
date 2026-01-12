package ast

import (
	"go/ast"
	"testing"
)

func TestFindFunctionInAST_Priority(t *testing.T) {
	src := `package main
func Build() {
	// ...
}
func wireComponents() {
	// ...
}
`
	_, f, err := parseSource(src)
	if err != nil {
		t.Fatalf("parseSource failed: %v", err)
	}

	// Case 1: Prefer wireComponents over Build
	nameParts := []string{"wireComponents", "build"}
	got := findFunctionInAST(f, nameParts)
	if got == nil {
		t.Fatal("findFunctionInAST returned nil")
	}
	if got.Name.Name != "wireComponents" {
		t.Errorf("Expected wireComponents, got %s. Priority was NOT respected because Build comes first in the file.", got.Name.Name)
	}
}

func TestGetReceiverName(t *testing.T) {
	t.Run("Picks architecture parameter", func(t *testing.T) {
		src := `package main
func wire(n *nodes, a *domain.Architecture) {
}`
		_, f, _ := parseSource(src)
		var fn *ast.FuncDecl
		for _, d := range f.Decls {
			if fnd, ok := d.(*ast.FuncDecl); ok && fnd.Name.Name == "wire" {
				fn = fnd
				break
			}
		}
		got := getReceiverName(fn, "default")
		if got != "a" {
			t.Errorf("Expected a, got %s. (Likely picked the first param 'n')", got)
		}
	})

	t.Run("Picks local architecture variable in method", func(t *testing.T) {
		src := `package main
func (b Builder) Build() *domain.Architecture {
	arch := domain.NewArchitecture("id", "name", "desc")
	return arch
}`
		_, f, _ := parseSource(src)
		var fn *ast.FuncDecl
		for _, d := range f.Decls {
			if fnd, ok := d.(*ast.FuncDecl); ok && fnd.Name.Name == "Build" {
				fn = fnd
				break
			}
		}
		got := getReceiverName(fn, "default")
		if got != "arch" {
			t.Errorf("Expected arch, got %s", got)
		}
	})
}
