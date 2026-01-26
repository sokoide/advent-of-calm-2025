package ast

import (
	"fmt"
	"go/ast"
	"go/token"
)

// addInterfaceToAST finds the DefineNode call for the given nodeID and adds an Interface() call after it.
func addInterfaceToAST(f *ast.File, fset *token.FileSet, nodeID, interfaceID, protocol string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		// Look for the statement containing DefineNode with matching ID
		var insertIndex int = -1
		var varName string

		for i, stmt := range block.List {
			// Case 1: Assignment statement: varName := a.DefineNode(...)
			if assign, ok := stmt.(*ast.AssignStmt); ok {
				for j, expr := range assign.Rhs {
					if isDefineNodeCallWithID(expr, nodeID) {
						// Found it! Get LHS name
						if ident, ok := assign.Lhs[j].(*ast.Ident); ok {
							varName = ident.Name
							insertIndex = i + 1
							break
						}
					}
				}
			}

			// Case 2: Expression statement: a.DefineNode(...)
			// Convert to variable assignment and add Interface as separate statement
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				if isDefineNodeCallWithID(exprStmt.X, nodeID) {
					// Generate a variable name from nodeID
					varName = generateVarNameFromID(nodeID)

					// Convert expression statement to assignment
					assignStmt := &ast.AssignStmt{
						Lhs: []ast.Expr{ast.NewIdent(varName)},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{exprStmt.X},
					}

					// Replace the expression statement with assignment
					block.List[i] = assignStmt

					// Set insert index for the interface call
					insertIndex = i + 1
					break
				}
			}

			if insertIndex != -1 {
				break
			}
		}

		// Handle variable assignment case
		if insertIndex != -1 && varName != "" {
			// Create the new statement: varName.Interface("id", "proto")
			// We can't easily construct a full AST manually with positions, so we construct a call expression
			// and rely on go/printer or formatFile to handle it.

			// Construct: varName.Interface
			fun := &ast.SelectorExpr{
				X:   &ast.Ident{Name: varName},
				Sel: &ast.Ident{Name: "Interface"},
			}

			// Args: "id", "proto"
			args := []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", interfaceID)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", protocol)},
			}

			call := &ast.CallExpr{Fun: fun, Args: args}
			stmt := &ast.ExprStmt{X: call}

			// Insert into block
			newOrderedList := append([]ast.Stmt{}, block.List[:insertIndex]...)
			newOrderedList = append(newOrderedList, stmt)
			newOrderedList = append(newOrderedList, block.List[insertIndex:]...)
			block.List = newOrderedList

			found = true
			return false
		}

		return true
	})

	if !found {
		return fmt.Errorf("definition for node %s not found", nodeID)
	}
	return nil
}

// deleteInterfaceFromAST finds an Interface("id", ...) call and removes it.
func deleteInterfaceFromAST(f *ast.File, interfaceID string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newOrderedList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			if isInterfaceCallWithID(stmt, interfaceID) {
				found = true
				continue // Delete
			}
			newOrderedList = append(newOrderedList, stmt)
		}

		if found {
			block.List = newOrderedList
			return false
		}
		return true
	})

	if !found {
		return fmt.Errorf("interface call %q not found", interfaceID)
	}
	return nil
}

// isInterfaceCallWithID checks if a statement is an Interface call with the given ID.
func isInterfaceCallWithID(stmt ast.Stmt, interfaceID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	// Traverse down the call chain to find the root Interface() call
	// Example: node.Interface("id").SetName("name")
	// AST: Call(SetName) -> X: Call(Interface) -> X: node
	curr := exprStmt.X
	for {
		call, ok := curr.(*ast.CallExpr)
		if !ok {
			return false
		}

		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Interface" {
				if len(call.Args) >= 1 {
					if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						val := lit.Value
						if len(val) >= 2 && val[0] == '"' {
							val = val[1 : len(val)-1]
						}
						if val == interfaceID {
							return true
						}
					}
				}
			}
			// Move down to the receiver
			curr = sel.X
		} else {
			return false
		}
	}
}
