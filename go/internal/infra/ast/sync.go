package ast

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"log"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SyncArchitectureFromJSON synchronizes the AST with the provided CALM JSON.
func SyncArchitectureFromJSON(f *ast.File, jsonStr string) error {
	var archData struct {
		Nodes []struct {
			ID   string `json:"unique-id"`
			Type string `json:"node-type"`
			Name string `json:"name"`
			Desc string `json:"description"`
		} `json:"nodes"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &archData); err != nil {
		return fmt.Errorf("failed to parse CALM JSON: %w", err)
	}

	// 1. Collect existing node IDs from AST
	existingNodeIDs := make(map[string]bool)
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "DefineNode" {
			return true
		}
		if len(call.Args) > 0 {
			if idLit, ok := call.Args[0].(*ast.BasicLit); ok {
				id := strings.Trim(idLit.Value, "\"")
				existingNodeIDs[id] = true
			}
		}
		return true
	})

	// 2. Update or Add nodes from JSON
	for _, n := range archData.Nodes {
		if n.ID == "" {
			continue
		}
		if existingNodeIDs[n.ID] {
			// Update properties
			if n.Name != "" {
				if err := UpdateNodePropertyInAST(f, n.ID, "name", n.Name); err != nil {
					log.Printf("Warning: failed to update name for %s: %v", n.ID, err)
				}
			}
			if n.Desc != "" {
				if err := UpdateNodePropertyInAST(f, n.ID, "description", n.Desc); err != nil {
					log.Printf("Warning: failed to update description for %s: %v", n.ID, err)
				}
			}
			delete(existingNodeIDs, n.ID)
		} else {
			// Add new node
			nodeType := "Service"
			if n.Type != "" {
				// Convert "service" -> "Service"
				caser := cases.Title(language.English)
				nodeType = caser.String(strings.ToLower(n.Type))
			}
			if err := AddNodeInAST(f, n.ID, nodeType, n.Name, n.Desc); err != nil {
				return err
			}
		}
	}

	// 3. Delete nodes that are no longer in JSON
	for id := range existingNodeIDs {
		if err := DeleteNodeInAST(f, id); err != nil {
			log.Printf("Warning: failed to delete node %s: %v", id, err)
		}
	}

	return nil
}

// UpdateNodeNameInAST finds a DefineNode call with the given id and updates its name argument.
func UpdateNodeNameInAST(f *ast.File, nodeID, newName string) error {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's arch.DefineNode(...)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "DefineNode" {
			return true
		}

		// Check first argument (unique-id)
		if len(call.Args) < 1 {
			return true
		}
		idLit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || idLit.Value != fmt.Sprintf("%q", nodeID) {
			return true
		}

		// Found it! Third argument is Name
		if len(call.Args) >= 3 {
			call.Args[2] = &ast.BasicLit{
				Kind:  idLit.Kind,
				Value: fmt.Sprintf("%q", newName),
			}
			found = true
			return false // stop inspection
		}

		return true
	})

	if !found {
		return fmt.Errorf("node with id %q not found in AST", nodeID)
	}
	return nil
}
