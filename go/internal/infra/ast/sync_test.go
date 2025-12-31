package ast

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestUpdateNodeName(t *testing.T) {
	src := `package main
func build() {
	arch.DefineNode("node1", Service, "Old Name", "desc")
}`

	expected := `package main

func build() {
	arch.DefineNode("node1", Service, "New Name", "desc")
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: Implement UpdateNodeNameInAST in Green Phase
	err = UpdateNodeNameInAST(f, "node1", "New Name")
	if err != nil {
		t.Fatalf("failed to update node name: %v", err)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		t.Fatal(err)
	}

	actual := buf.String()
	if actual != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, actual)
	}
}

func TestAddNode(t *testing.T) {
	src := `package main
func BuildArchitecture() {
	arch.DefineNode("node1", Service, "Name 1", "desc")
}`

	expected := `package main

func BuildArchitecture() {
	arch.DefineNode("node1", Service, "Name 1", "desc")
	arch.DefineNode("node2", Service, "Name 2", "desc")
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	err = AddNodeInAST(f, "node2", "Service", "Name 2", "desc")
	if err != nil {
		t.Fatalf("failed to add node: %v", err)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		t.Fatal(err)
	}

	actual := buf.String()
	if actual != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, actual)
	}
}

func TestSyncArchitectureFromJSON(t *testing.T) {
	src := `package main
func build() {
	arch.DefineNode("node1", Service, "Old Name", "desc")
}`

	jsonInput := `{
		"unique-id": "test-arch",
		"nodes": [
			{
				"unique-id": "node1",
				"node-type": "service",
				"name": "Updated Name",
				"description": "updated desc"
			},
			{
				"unique-id": "node2",
				"node-type": "database",
				"name": "New DB",
				"description": "new desc"
			}
		]
	}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	err = SyncArchitectureFromJSON(f, jsonInput)
	if err != nil {
		t.Fatalf("failed to sync from JSON: %v", err)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		t.Fatal(err)
	}

	actual := buf.String()
	// Check if both nodes exist in the output
	if !strings.Contains(actual, `"Updated Name"`) {
		t.Errorf("expected updated name not found")
	}
	if !strings.Contains(actual, `"node2"`) {
		t.Errorf("expected new node not found")
	}
}
