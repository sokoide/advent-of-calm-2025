package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

// addComposedOfToAST adds a new ComposedOf relationship to the AST.
func addComposedOfToAST(f *ast.File, id, containerID string, childNodeIDs []string) error {
	if id == "" {
		id = fmt.Sprintf("composed-%s", containerID)
	}

	fn := findFunctionInAST(f, []string{"wireComponents", "build"})
	if fn == nil {
		return fmt.Errorf("wireComponents or build function not found in AST")
	}

	receiverName := getReceiverName(fn, "a")

	// Build the node IDs slice: []string{"n1", "n2"}
	elts := make([]ast.Expr, len(childNodeIDs))
	for i, nid := range childNodeIDs {
		elts[i] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", nid)}
	}

	newStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(receiverName),
				Sel: ast.NewIdent("ComposedOf"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", id)},
				&ast.BasicLit{Kind: token.STRING, Value: "\"Container relationship\""},
				ast.NewIdent(containerID),
				&ast.CompositeLit{
					Type: &ast.ArrayType{Elt: ast.NewIdent("string")},
					Elts: elts,
				},
			},
		},
	}

	insertStmtBeforeReturn(fn, newStmt)

	log.Printf("📝 Adding ComposedOf: %s (container: %s)", id, containerID)
	return nil
}

// deleteComposedOfFromAST removes a ComposedOf call from the AST.
func deleteComposedOfFromAST(f *ast.File, composedOfID string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			if isComposedOfCallWithID(stmt, composedOfID) {
				found = true
				continue // Skip this statement to delete it
			}
			newList = append(newList, stmt)
		}

		if found {
			block.List = newList
		}

		return true
	})

	if !found {
		return fmt.Errorf("composed-of %q not found", composedOfID)
	}
	return nil
}

// isComposedOfCallWithID checks if a statement is a ComposedOf call with the given ID.
func isComposedOfCallWithID(stmt ast.Stmt, composedOfID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	// Check if it's a method call like a.ComposedOf(...)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "ComposedOf" {
		return false
	}

	if len(call.Args) < 1 {
		return false
	}

	// Check first argument (composed-of ID)
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}

	val := strings.Trim(lit.Value, "\"")
	return val == composedOfID
}

// updateComposedOfInAST updates a ComposedOf call in the AST (description and/or child nodes).
func updateComposedOfInAST(f *ast.File, composedOfID, newDesc string, newChildNodeIDs []string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's a ComposedOf call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if sel.Sel.Name != "ComposedOf" {
			return true
		}

		if len(call.Args) < 2 {
			return true
		}

		// Check first argument (ID)
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		val := strings.Trim(lit.Value, "\"")
		if val != composedOfID {
			return true
		}

		// Update description (second argument) if provided
		if newDesc != "" && len(call.Args) >= 2 {
			call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newDesc)}
		}

		// Update child nodes if provided
		if len(newChildNodeIDs) > 0 && len(call.Args) >= 4 {
			// Args[2] = container, Args[3] = nodes slice
			// Build new composite literal for nodes
			elts := make([]ast.Expr, len(newChildNodeIDs))
			for i, nodeID := range newChildNodeIDs {
				elts[i] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", nodeID)}
			}
			call.Args[3] = &ast.CompositeLit{
				Type: &ast.ArrayType{Elt: &ast.Ident{Name: "string"}},
				Elts: elts,
			}
		}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("composed-of %q not found", composedOfID)
	}
	return nil
}

// deleteComposedOfReferencingVariable removes or updates ComposedOf calls that reference the given variable or node ID
func deleteComposedOfReferencingVariable(f *ast.File, varName string, nodeID string) {
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		newList := make([]ast.Stmt, 0, len(fn.Body.List))
		for _, stmt := range fn.Body.List {
			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				newList = append(newList, stmt)
				continue
			}

			call, ok := exprStmt.X.(*ast.CallExpr)
			if !ok {
				newList = append(newList, stmt)
				continue
			}

			// Check if it's a ComposedOf call
			if isComposedOfCall(call) {
				// 1. Check if the container (parent) matches varName or nodeID -> Delete entire statement
				if composedOfContainerMatches(call, varName, nodeID) {
					continue // Delete statement
				}

				// 2. Check and filter children slice
				if removeComposedOfChild(call, varName, nodeID) {
					// Modified in place, keep the statement
				}
			}

			newList = append(newList, stmt)
		}
		fn.Body.List = newList
	}
}

// isComposedOfCall checks if a call is a ComposedOf call.
func isComposedOfCall(call *ast.CallExpr) bool {
	// Walk up chain to find ComposedOf
	curr := call
	for curr != nil {
		sel, ok := curr.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if sel.Sel.Name == "ComposedOf" {
			return true
		}
		// Move up
		if innerCall, ok := sel.X.(*ast.CallExpr); ok {
			curr = innerCall
		} else {
			break
		}
	}
	return false
}

// getComposedOfBaseCall gets the base ComposedOf call from a chain.
func getComposedOfBaseCall(call *ast.CallExpr) *ast.CallExpr {
	curr := call
	for curr != nil {
		sel, ok := curr.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil
		}
		if sel.Sel.Name == "ComposedOf" {
			return curr
		}
		if innerCall, ok := sel.X.(*ast.CallExpr); ok {
			curr = innerCall
		} else {
			break
		}
	}
	return nil
}

// composedOfContainerMatches checks if the container of a ComposedOf matches the given varName or nodeID.
func composedOfContainerMatches(call *ast.CallExpr, varName string, nodeID string) bool {
	baseCall := getComposedOfBaseCall(call)
	if baseCall == nil || len(baseCall.Args) < 3 {
		return false
	}
	// Container is 3rd argument
	return variableReferenceMatches(baseCall.Args[2], varName, nodeID)
}

// removeComposedOfChild removes the varName or nodeID from the children slice. Returns true if modified.
func removeComposedOfChild(call *ast.CallExpr, varName string, nodeID string) bool {
	baseCall := getComposedOfBaseCall(call)
	if baseCall == nil || len(baseCall.Args) < 4 {
		return false
	}

	comp, ok := baseCall.Args[3].(*ast.CompositeLit)
	if !ok {
		return false
	}

	newElts := make([]ast.Expr, 0, len(comp.Elts))
	modified := false
	for _, elt := range comp.Elts {
		if variableReferenceMatches(elt, varName, nodeID) {
			modified = true
			continue
		}
		newElts = append(newElts, elt)
	}

	if modified {
		comp.Elts = newElts
	}
	return modified
}
