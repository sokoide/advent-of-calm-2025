package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strconv"
	"strings"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

// ApplyPatch applies a list of patch operations to the source.
func (GoASTSyncer) ApplyPatch(src string, ops []domain.PatchOperation) (string, error) {
	fset, f, err := parseSource(src)
	if err != nil {
		return "", err
	}

	for _, op := range ops {
		switch op.Type {
		case domain.PatchUpdateNode:
			if op.Origin == nil {
				log.Printf("Warning: update-node requires origin: %v", op)
				continue
			}
			if err := updateNodeAtLine(f, fset, op); err != nil {
				return "", fmt.Errorf("failed to update node at line %d: %w", op.Origin.Line, err)
			}
		case domain.PatchDeleteNode:
			if op.Origin == nil {
				log.Printf("Warning: delete-node requires origin: %v", op)
				continue
			}
			if op.Origin.LoopVar != "" {
				// Loop node deletion = Decrement loop variable
				if err := updateLoopVariable(f, op.Origin.LoopVar, -1); err != nil {
					return "", fmt.Errorf("failed to decrement loop var %s: %w", op.Origin.LoopVar, err)
				}
			} else {
				// Explicit node deletion = Remove statement
				if err := deleteNodeAtLine(f, fset, op.Origin.Line); err != nil {
					return "", fmt.Errorf("failed to delete node at line %d: %w", op.Origin.Line, err)
				}
			}
		case domain.PatchDeleteRelationship:
			// Delete Connect() call by relationship ID (no origin needed)
			if op.NodeID == "" {
				log.Printf("Warning: delete-relationship requires nodeId (relationship ID)")
				continue
			}
			if err := deleteRelationshipByID(f, op.NodeID); err != nil {
				return "", fmt.Errorf("failed to delete relationship %s: %w", op.NodeID, err)
			}
		case domain.PatchAddRelationship:
			// Add a new Connect() call
			if op.NodeID == "" || op.SourceNode == "" || op.TargetNode == "" {
				log.Printf("Warning: add-relationship requires nodeId, sourceNode, and targetNode")
				continue
			}
			if err := addRelationshipToAST(f, fset, op.NodeID, op.SourceNode, op.TargetNode); err != nil {
				return "", fmt.Errorf("failed to add relationship %s: %w", op.NodeID, err)
			}
		}
	}

	result, err := formatFile(fset, f)
	if err != nil {
		return "", err
	}

	// If there's a pending relationship to add, insert it into the source
	if pendingRelationship.pending {
		result = insertRelationshipIntoSource(result, pendingRelationship.id, pendingRelationship.source, pendingRelationship.target)
		pendingRelationship.pending = false
	}

	return result, nil
}

func updateNodeAtLine(f *ast.File, fset *token.FileSet, op domain.PatchOperation) error {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		// We are looking for the DefineNode call
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check line number
		// Note: fset.Position(n.Pos()).Line might be what we need
		// The Origin.Line points to the start of the defined node, likely the DefineNode call or assignment
		pos := fset.Position(call.Pos())
		if pos.Line != op.Origin.Line {
			return true
		}

		// Verify it's a DefineNode call
		if !isDefineNode(call) {
			return true
		}

		// Apply update
		if err := applyPropertyUpdate(call, op.Property, op.Value); err != nil {
			log.Printf("Failed to apply property update: %v", err)
			return false
		}

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("DefineNode call not found at line %d", op.Origin.Line)
	}
	return nil
}

func isDefineNode(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "DefineNode"
}

func applyPropertyUpdate(call *ast.CallExpr, property string, value interface{}) error {
	valStr := fmt.Sprintf("%v", value)

	// DefineNode args: id, type, name, desc, options...
	if len(call.Args) < 4 {
		return fmt.Errorf("DefineNode call has too few arguments")
	}

	switch property {
	case "name":
		call.Args[2] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
	case "description":
		call.Args[3] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", valStr)}
	case "node-type":
		// TODO: Handle type update (requires import awareness for domain.Service etc)
		// For now, simple string replacement if it was using string literals,
		// but typically it uses domain.Type constants.
		// Let's defer complex type updates or assume string based on current impl needs.
	case "owner":
		// Search for WithOwner option
		updated := false
		for _, arg := range call.Args[4:] {
			if updateOptionCall(arg, "WithOwner", 0, valStr) {
				updated = true
				break
			}
		}
		// If not found, we should append a new option?
		// For simplicity in this iteration, we only update existing options.
		if !updated {
			// append WithOwner? Implementation complex for appending
		}
	}
	return nil
}

func updateOptionCall(expr ast.Expr, funcName string, argIndex int, newValue string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	if !isOwnerCall(call.Fun, funcName) {
		return false
	}
	if len(call.Args) <= argIndex {
		return false
	}
	call.Args[argIndex] = &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", newValue)}
	return true
}

func isOwnerCall(fun ast.Expr, name string) bool {
	switch v := fun.(type) {
	case *ast.Ident:
		return v.Name == name
	case *ast.SelectorExpr:
		return v.Sel.Name == name
	}
	return false
}

func deleteNodeAtLine(f *ast.File, fset *token.FileSet, line int) error {
	// We need to find the statement containing line X and remove it from its block
	found := false

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		// Look for blocks
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		newOrderedList := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			pos := fset.Position(stmt.Pos())
			// We check if the statement STARTS at the line, or contains it.
			// Ideally the Origin line is the start line.
			if pos.Line == line {
				found = true
				continue // Skip this statement (delete)
			}
			newOrderedList = append(newOrderedList, stmt)
		}

		if found {
			block.List = newOrderedList
			return false // Stop traversing
		}
		return true
	})

	if !found {
		// Try finding it in top-level declarations (unlikely for DefineNode inside functions, but possible)
		// Actually typical DefineNode is inside Build() which is in a FuncDecl, so BlockStmt traversal should catch it.
		return fmt.Errorf("statement at line %d not found in any block", line)
	}
	return nil
}

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

		// Check if it's a Connect() or Interacts() call
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		funcName := sel.Sel.Name
		if funcName != "Connect" && funcName != "Interacts" {
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

// addRelationshipToAST records a new Connect() call to be added.
// Since AST insertion is complex, we use source code manipulation approach.
// This function stores the relationship info in a global variable that will be
// appended to the formatted source code at the end of ApplyPatch.
var pendingRelationship struct {
	id, source, target string
	pending            bool
}

func addRelationshipToAST(f *ast.File, fset *token.FileSet, relationshipID, sourceNode, targetNode string) error {
	// Store the relationship for later source code insertion
	pendingRelationship.id = relationshipID
	pendingRelationship.source = sourceNode
	pendingRelationship.target = targetNode
	pendingRelationship.pending = true

	log.Printf("📝 Adding relationship: %s from %s to %s", relationshipID, sourceNode, targetNode)
	return nil
}

// insertRelationshipIntoSource inserts a Connect() call into the Go source code.
// It looks for the wireComponents function and inserts before "return lc".
func insertRelationshipIntoSource(src, relationshipID, sourceNode, targetNode string) string {
	// Generate the Connect() line to insert
	// Format: a.Connect("id", "Connection from source to target", nodeVar1.Unique, nodeVar2.UniqueID)
	// Since we don't have node variable names, we use a simplified inline format
	connectLine := fmt.Sprintf(`
	// GUI-generated connection: %s
	a.Connect("%s", "Connection from %s to %s", "%s", "%s")
`, relationshipID, relationshipID, sourceNode, targetNode, sourceNode, targetNode)

	// Find "return lc" in wireComponents and insert before it
	// Use a pattern that matches the return statement with proper indentation
	pattern := "\treturn lc\n"
	insertPoint := strings.LastIndex(src, pattern)
	if insertPoint == -1 {
		// Fallback: try to find "return lc" anywhere
		pattern = "return lc"
		insertPoint = strings.LastIndex(src, pattern)
	}

	if insertPoint == -1 {
		log.Printf("Warning: Could not find insertion point for Connect() in wireComponents")
		return src
	}

	// Insert before the return statement
	return src[:insertPoint] + connectLine + src[insertPoint:]
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
