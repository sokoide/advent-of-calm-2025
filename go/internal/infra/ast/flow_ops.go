package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

// deleteFlowFromAST finds a DefineFlow("id", ...) call and removes it.
func deleteFlowFromAST(f *ast.File, flowID string) error {
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
			if isDefineFlowCallWithID(stmt, flowID) {
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
		return fmt.Errorf("flow definition %q not found", flowID)
	}
	return nil
}

// isDefineFlowCallWithID checks if a statement is a DefineFlow call with the given ID.
func isDefineFlowCallWithID(stmt ast.Stmt, flowID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	// Traverse down the call chain to find the root DefineFlow() call
	// Example: a.DefineFlow("id", ...).Steps(...)
	curr := exprStmt.X
	for {
		call, ok := curr.(*ast.CallExpr)
		if !ok {
			return false
		}

		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "DefineFlow" {
				if len(call.Args) >= 1 {
					if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						val := strings.Trim(lit.Value, "\"")
						if val == flowID {
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

// addFlowToAST adds a new DefineFlow call to the AST.
func addFlowToAST(f *ast.File, flowID, name, desc string, steps []string) error {
	fn := findFunctionInAST(f, []string{"defineFlows", "build"})
	if fn == nil {
		return fmt.Errorf("defineFlows or build function not found in AST")
	}

	receiverName := getReceiverName(fn, "a")

	// 1. Create DefineFlow call
	defineFlowCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(receiverName),
			Sel: ast.NewIdent("DefineFlow"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", flowID)},
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", name)},
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
		},
	}

	// 2. Create Steps call
	stepSpecs := make([]ast.Expr, 0, len(steps))
	for _, stepID := range steps {
		stepSpecs = append(stepSpecs, &ast.CompositeLit{
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("domain"),
				Sel: ast.NewIdent("StepSpec"),
			},
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key:   ast.NewIdent("ID"),
					Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", stepID)},
				},
			},
		})
	}

	stepsCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   defineFlowCall,
			Sel: ast.NewIdent("Steps"),
		},
		Args: stepSpecs,
	}

	insertStmtBeforeReturn(fn, &ast.ExprStmt{X: stepsCall})

	log.Printf("📝 Adding Flow: %s", flowID)
	return nil
}

// updateFlowInAST updates an existing DefineFlow call.
func updateFlowInAST(f *ast.File, flowID, name, desc string, steps []string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		// Look for statement containing DefineFlow
		stmt, ok := n.(ast.Stmt)
		if !ok {
			return true
		}

		if !isDefineFlowCallWithID(stmt, flowID) {
			return true
		}

		// Found the statement. Now we need to update arguments of DefineFlow and Steps.
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return true
		}

		// Traverse chain to find DefineFlow and Steps calls
		var defineFlowCall *ast.CallExpr
		var stepsCall *ast.CallExpr

		curr := exprStmt.X
		for {
			call, ok := curr.(*ast.CallExpr)
			if !ok {
				break
			}

			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "DefineFlow" {
					defineFlowCall = call
				} else if sel.Sel.Name == "Steps" {
					stepsCall = call
				}
				curr = sel.X
			} else {
				break
			}
		}

		if defineFlowCall != nil {
			// Update Name (arg 1) and Desc (arg 2)
			// Arg 0 is ID
			if len(defineFlowCall.Args) >= 3 {
				defineFlowCall.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", name)}
				defineFlowCall.Args[2] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)}
			}
		}

		if stepsCall != nil {
			// Update Steps args
			newArgs := make([]ast.Expr, 0, len(steps))
			for _, stepID := range steps {
				// domain.StepSpec{ID: "stepID"}
				val := &ast.CompositeLit{
					Type: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "domain"},
						Sel: &ast.Ident{Name: "StepSpec"},
					},
					Elts: []ast.Expr{
						&ast.KeyValueExpr{
							Key:   &ast.Ident{Name: "ID"},
							Value: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", stepID)},
						},
						// We don't preserve description for now as it's hard to track
					},
				}
				newArgs = append(newArgs, val)
			}
			stepsCall.Args = newArgs
		} else {
			// If Steps() call is missing, we need to append it.
			// Ideally we would wrap the existing expression `X` with `CallExpr{Fun: SelectorExpr{X: X, Sel: "Steps"}, Args: ...}`
			// But since we can't easily modify the structure here without re-writing the whole statement,
			// we will log a warning.
			// TODO: Implement adding Steps() if missing.
			log.Printf("Warning: Steps() call not found for flow %s, skipping steps update", flowID)
		}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("flow %q not found", flowID)
	}
	return nil
}
