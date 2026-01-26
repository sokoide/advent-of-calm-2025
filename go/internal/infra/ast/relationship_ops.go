package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strconv"
	"strings"
)

// deleteRelationshipByID finds a Connect() call with the specified relationship ID and removes it.
func deleteRelationshipByID(f *ast.File, relationshipID string) error {
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
			// Check if this statement contains a Connect() or Interacts() call with matching ID
			if containsRelationshipWithID(stmt, relationshipID) {
				found = true
				continue // Skip this statement (delete)
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
		return fmt.Errorf("relationship with ID %q not found", relationshipID)
	}
	return nil
}

// containsRelationshipWithID checks if a statement contains a Connect() or Interacts() call with the given ID.
// Handles frontend-generated IDs with -N suffix (e.g., "customer-interacts-lb-0" -> "customer-interacts-lb").
func containsRelationshipWithID(stmt ast.Stmt, relationshipID string) bool {
	// Strip -N suffix from relationshipID if present (frontend adds -0, -1, etc.)
	baseID := relationshipID
	if idx := strings.LastIndex(relationshipID, "-"); idx > 0 {
		suffix := relationshipID[idx+1:]
		// Check if suffix is numeric
		if _, err := strconv.Atoi(suffix); err == nil {
			baseID = relationshipID[:idx]
		}
	}

	found := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's a Connect(), Interacts(), or ComposedOf() call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		funcName := sel.Sel.Name
		if funcName != "Connect" && funcName != "Interacts" && funcName != "ComposedOf" {
			return true
		}

		// Check first argument is the relationship ID
		if len(call.Args) < 1 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		// Remove quotes from literal value
		idValue := lit.Value
		if len(idValue) >= 2 && idValue[0] == '"' && idValue[len(idValue)-1] == '"' {
			idValue = idValue[1 : len(idValue)-1]
		}

		// Match against both full ID and base ID
		if idValue == relationshipID || idValue == baseID {
			found = true
			return false
		}
		return true
	})
	return found
}

// addToArrayVariable finds a variable definition (varName := []string{...}) and adds a new element.
func addToArrayVariable(f *ast.File, varName, newValue string) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		// Look for assignment: varName := []string{...} or varName = append(...)
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		if len(assign.Lhs) == 0 || len(assign.Rhs) == 0 {
			return true
		}
		// Check if LHS is our variable
		lhsIdent, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhsIdent.Name != varName {
			return true
		}
		// Check if RHS is a composite literal (array)
		compLit, ok := assign.Rhs[0].(*ast.CompositeLit)
		if !ok {
			return true
		}
		// Add new element
		newElem := &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newValue)}
		compLit.Elts = append(compLit.Elts, newElem)
		found = true
		return false
	})
	return found
}

// updateRelationshipInAST finds a Connect() or Interacts() call with the specified ID and updates its properties.
func updateRelationshipInAST(f *ast.File, relationshipID, property string, value interface{}) error {
	found := false
	valStr := fmt.Sprintf("%v", value)

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		stmt, ok := n.(ast.Stmt)
		if !ok {
			return true
		}

		if !containsRelationshipWithID(stmt, relationshipID) {
			return true
		}

		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return true
		}

		// Traverse call chain from outermost to innermost
		// e.g., Connect(id, desc, ...).Protocol(proto)
		// Call(Protocol) -> X: Call(Connect)
		curr := exprStmt.X
		for {
			call, ok := curr.(*ast.CallExpr)
			if !ok {
				break
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				break
			}

			// 1. Check if the current call is the property method (e.g. .Protocol("..."))
			if strings.EqualFold(sel.Sel.Name, property) {
				if len(call.Args) >= 1 {
					call.Args[0] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
					found = true
					return false
				}
			}

			// 2. Check if the current call is the root Connect/Interacts/ComposedOf call
			if sel.Sel.Name == "Connect" || sel.Sel.Name == "Interacts" {
				if strings.EqualFold(property, "description") && len(call.Args) >= 2 {
					call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
					found = true
					return false
				}
				// ID match is already verified by containsRelationshipWithID
				// Stop here as we've hit the root
				break
			}

			// 3. Handle ComposedOf - signature: ComposedOf(id, desc, container, []string{nodes})
			if sel.Sel.Name == "ComposedOf" {
				if strings.EqualFold(property, "add-child-node") {
					// Add node to the array (4th argument)
					if len(call.Args) >= 4 {
						// Case 1: Direct array literal []string{...}
						if compLit, ok := call.Args[3].(*ast.CompositeLit); ok {
							newElem := &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
							compLit.Elts = append(compLit.Elts, newElem)
							found = true
							return false
						}
						// Case 2: Variable reference (e.g., sysNodes)
						if ident, ok := call.Args[3].(*ast.Ident); ok {
							varName := ident.Name
							// Find the variable definition and add to its array
							if addToArrayVariable(f, varName, valStr) {
								found = true
								return false
							}
						}
					}
				} else if strings.EqualFold(property, "description") && len(call.Args) >= 2 {
					call.Args[1] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
					found = true
					return false
				}
				break
			}

			curr = sel.X
		}

		return true
	})

	if !found {
		return fmt.Errorf("relationship %s property %s not found/updated", relationshipID, property)
	}
	return nil
}

// addRelationshipToAST adds a new Connect() or Interacts() call to the AST.
func addRelationshipToAST(
	f *ast.File,
	fset *token.FileSet,
	relationshipID, sourceNode, targetNode string,
	isInteracts bool,
) error {
	fn := findFunctionInAST(f, []string{"wireComponents", "build"})
	if fn == nil {
		return fmt.Errorf("wireComponents or build function not found in AST")
	}

	receiverName := getReceiverName(fn, "a")
	funcName := "Connect"
	if isInteracts {
		funcName = "Interacts"
	}

	desc := fmt.Sprintf("%s from %s to %s", funcName, sourceNode, targetNode)
	newStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(receiverName),
				Sel: ast.NewIdent(funcName),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", relationshipID)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", desc)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", sourceNode)},
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", targetNode)},
			},
		},
	}

	insertStmtBeforeReturn(fn, newStmt)

	if isInteracts {
		log.Printf("📝 Adding Interacts: %s from %s to %s", relationshipID, sourceNode, targetNode)
	} else {
		log.Printf("📝 Adding Connect: %s from %s to %s", relationshipID, sourceNode, targetNode)
	}
	return nil
}

func updateLoopVariable(f *ast.File, varName string, delta int) error {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		// 1. Const declaration: const numGateways = 2
		if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.CONST {
			for _, spec := range gen.Specs {
				vspec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vspec.Names {
					if name.Name == varName {
						if len(vspec.Values) > i {
							if lit, ok := vspec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.INT {
								val, err := strconv.Atoi(lit.Value)
								if err == nil {
									newVal := val + delta
									if newVal < 0 {
										newVal = 0
									}
									lit.Value = strconv.Itoa(newVal)
									found = true
									return false
								}
							}
						}
					}
				}
			}
		}

		// 2. Assignment: numGateways := 2 or var numGateways = 2
		// Handling var declaration
		if gen, ok := n.(*ast.GenDecl); ok && gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				vspec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vspec.Names {
					if name.Name == varName {
						if len(vspec.Values) > i {
							if lit, ok := vspec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.INT {
								val, err := strconv.Atoi(lit.Value)
								if err == nil {
									newVal := val + delta
									if newVal < 0 {
										newVal = 0
									}
									lit.Value = strconv.Itoa(newVal)
									found = true
									return false
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	if !found {
		return fmt.Errorf("variable/constant %q not found or not an integer literal", varName)
	}
	return nil
}
