package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

// addControlToAST adds a new AddControl call to the AST.
func addControlToAST(f *ast.File, controlID, desc string) error {
	fn := findFunctionInAST(f, []string{"addGlobalControls", "build", "defineNodes"})
	if fn == nil {
		return fmt.Errorf("suitable function for AddControl not found in AST")
	}

	receiverName := getReceiverName(fn, "arch")

	newStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(receiverName),
				Sel: ast.NewIdent("AddControl"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", controlID)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
			},
		},
	}

	insertStmtBeforeReturn(fn, newStmt)

	log.Printf("📝 Adding Control: %s", controlID)
	return nil
}

// deleteControlFromAST removes an AddControl call from the AST.
func deleteControlFromAST(f *ast.File, controlID string) error {
	found := false

	// Controls are typically added via arch.AddControl("id", "desc")
	// We need to find and remove such statements
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
			if isAddControlCallWithID(stmt, controlID) {
				found = true
				continue // Skip this statement to delete it
			}
			newList = append(newList, stmt)
		}

		if found {
			block.List = newList
			return false
		}

		return true
	})

	if !found {
		return fmt.Errorf("control %q not found", controlID)
	}
	return nil
}

// isAddControlCallWithID checks if a statement is an AddControl call with the given ID.
func isAddControlCallWithID(stmt ast.Stmt, controlID string) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	// Check if it's a method call like arch.AddControl(...)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "AddControl" {
		return false
	}

	if len(call.Args) < 1 {
		return false
	}

	// Check first argument (control ID)
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}

	val := strings.Trim(lit.Value, "\"")
	return val == controlID
}

// updateControlInAST updates the description of an AddControl call in the AST.
func updateControlInAST(f *ast.File, controlID, newDesc string) error {
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's an AddControl call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if sel.Sel.Name != "AddControl" {
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
		if val != controlID {
			return true
		}

		// Update description (second argument)
		call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newDesc)}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("control %q not found", controlID)
	}
	return nil
}
