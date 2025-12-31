package ast

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
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
				nodeType = strings.Title(strings.ToLower(n.Type))
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

// AddNodeInAST appends a new DefineNode call to the build or defineNodes function in the AST.
func AddNodeInAST(f *ast.File, nodeID, nodeType, name, desc string) error {
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Target Build... or defineNodes functions
		nameLower := strings.ToLower(fn.Name.Name)
		if !strings.Contains(nameLower, "build") && !strings.Contains(nameLower, "definenodes") {
			continue
		}

		// Determine receiver name (e.g., 'arch' or 'a')
		receiverName := "arch"
		if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
			for _, p := range fn.Type.Params.List {
				if len(p.Names) > 0 {
					receiverName = p.Names[0].Name
					break
				}
			}
		}

		// Create: <receiver>.DefineNode("id", domain.<Type>, "name", "desc")
		newStmt := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(receiverName),
					Sel: ast.NewIdent("DefineNode"),
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", nodeID)},
					&ast.SelectorExpr{
						X:   ast.NewIdent("domain"),
						Sel: ast.NewIdent(nodeType),
					},
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", name)},
					&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
				},
			},
		}

		// Insert before return if exists, otherwise append
		inserted := false
		for i, stmt := range fn.Body.List {
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				// Insert before return
				newList := make([]ast.Stmt, 0, len(fn.Body.List)+1)
				newList = append(newList, fn.Body.List[:i]...)
				newList = append(newList, newStmt)
				newList = append(newList, fn.Body.List[i:]...)
				fn.Body.List = newList
				inserted = true
				break
			}
		}
		if !inserted {
			fn.Body.List = append(fn.Body.List, newStmt)
		}
		return nil
	}

	return fmt.Errorf("build function not found in AST")
}

// UpdateNodePropertyInAST updates a specific property (name, description, owner, etc.) of a node.
func UpdateNodePropertyInAST(f *ast.File, nodeID, property, value string) error {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "DefineNode" {
			return true
		}

		if len(call.Args) < 1 {
			return true
		}
		idLit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || strings.Trim(idLit.Value, "\"") != nodeID {
			return true
		}

		// Handle specific properties
		switch property {
		case "name":
			if len(call.Args) >= 3 {
				call.Args[2] = &ast.BasicLit{Kind: idLit.Kind, Value: fmt.Sprintf("%q", value)}
				found = true
			}
		case "description":
			if len(call.Args) >= 4 {
				call.Args[3] = &ast.BasicLit{Kind: idLit.Kind, Value: fmt.Sprintf("%q", value)}
				found = true
			}
		case "owner":
			// owner is usually in WithOwner("owner", "cc") option
			for _, arg := range call.Args[4:] {
				optCall, ok := arg.(*ast.CallExpr)
				if !ok {
					continue
				}
				optSel, ok := optCall.Fun.(*ast.Ident)
				if ok && optSel.Name == "WithOwner" && len(optCall.Args) >= 1 {
					optCall.Args[0] = &ast.BasicLit{Kind: idLit.Kind, Value: fmt.Sprintf("%q", value)}
					found = true
				}
			}
		}

		return !found
	})

	if !found {
		return fmt.Errorf("property %q for node %q not updated in AST", property, nodeID)
	}
	return nil
}

// DeleteNodeInAST removes a DefineNode call with the given id from the AST.
func DeleteNodeInAST(f *ast.File, nodeID string) error {
	found := false
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		newList := make([]ast.Stmt, 0, len(fn.Body.List))
		for _, stmt := range fn.Body.List {
			shouldDelete := false
			
			// Check if stmt is arch.DefineNode("nodeID", ...)
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				if call, ok := exprStmt.X.(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "DefineNode" {
						if len(call.Args) > 0 {
							if idLit, ok := call.Args[0].(*ast.BasicLit); ok && idLit.Value == fmt.Sprintf("%q", nodeID) {
								shouldDelete = true
								found = true
							}
						}
					}
				}
			}

			if !shouldDelete {
				newList = append(newList, stmt)
			}
		}
		fn.Body.List = newList
	}

	if !found {
		return fmt.Errorf("node with id %q not found in AST", nodeID)
	}
	return nil
}
