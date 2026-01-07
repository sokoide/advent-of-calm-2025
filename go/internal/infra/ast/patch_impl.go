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
			// IMPORTANT: Find variable name BEFORE deleting the node definition
			var varName string
			if op.NodeID != "" {
				varName = findVariableNameForNodeID(f, op.NodeID)
				if varName != "" {
					log.Printf("Cascade delete: found variable name '%s' for nodeID '%s'", varName, op.NodeID)
				} else {
					log.Printf("Cascade delete: no variable name found for nodeID '%s'", op.NodeID)
				}
			}
			// Now delete the node
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
			// Cascade delete: Remove relationships that reference this node
			if varName != "" {
				deleteRelationshipsReferencingVariable(f, varName)
				deleteComposedOfReferencingVariable(f, varName)
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
			// Add a new Connect() or Interacts() call
			if op.NodeID == "" || op.SourceNode == "" || op.TargetNode == "" {
				log.Printf("Warning: add-relationship requires nodeId, sourceNode, and targetNode")
				continue
			}
			if err := addRelationshipToAST(f, fset, op.NodeID, op.SourceNode, op.TargetNode, op.IsInteracts); err != nil {
				return "", fmt.Errorf("failed to add relationship %s: %w", op.NodeID, err)
			}
		case domain.PatchAddInterface:
			// Add a new Interface() call for a node
			if op.NodeID == "" || op.InterfaceID == "" || op.Protocol == "" {
				log.Printf("Warning: add-interface requires nodeId, interfaceId, and protocol")
				continue
			}
			if err := addInterfaceToAST(f, fset, op.NodeID, op.InterfaceID, op.Protocol); err != nil {
				return "", fmt.Errorf("failed to add interface %s to %s: %w", op.InterfaceID, op.NodeID, err)
			}
		case domain.PatchDeleteInterface:
			// Delete Interface() call
			if op.InterfaceID == "" {
				log.Printf("Warning: delete-interface requires interfaceId")
				continue
			}
			if err := deleteInterfaceFromAST(f, op.InterfaceID); err != nil {
				return "", fmt.Errorf("failed to delete interface %s: %w", op.InterfaceID, err)
			}
		case domain.PatchDeleteFlow:
			// Delete DefineFlow() call
			if op.FlowID == "" {
				log.Printf("Warning: delete-flow requires flowId")
				continue
			}
			if err := deleteFlowFromAST(f, op.FlowID); err != nil {
				return "", fmt.Errorf("failed to delete flow %s: %w", op.FlowID, err)
			}
		case domain.PatchAddFlow:
			if op.FlowID == "" || op.FlowName == "" {
				log.Printf("Warning: add-flow requires flowId and flowName")
				continue
			}
			if err := addFlowToAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
				return "", fmt.Errorf("failed to add flow %s: %w", op.FlowID, err)
			}
		case domain.PatchUpdateFlow:
			if op.FlowID == "" {
				log.Printf("Warning: update-flow requires flowId")
				continue
			}
			if err := updateFlowInAST(f, op.FlowID, op.FlowName, op.FlowDesc, op.FlowSteps); err != nil {
				return "", fmt.Errorf("failed to update flow %s: %w", op.FlowID, err)
			}
		case domain.PatchAddComposedOf:
			if op.ContainerID == "" || len(op.ChildNodeIDs) == 0 {
				log.Printf("Warning: add-composed-of requires containerId and childNodeIds")
				continue
			}
			if err := addComposedOfToAST(f, op.NodeID, op.ContainerID, op.ChildNodeIDs); err != nil {
				return "", fmt.Errorf("failed to add composed-of: %w", err)
			}
		case domain.PatchAddControl:
			if op.ControlID == "" {
				log.Printf("Warning: add-control requires controlId")
				continue
			}
			if err := addControlToAST(f, op.ControlID, op.ControlDesc); err != nil {
				return "", fmt.Errorf("failed to add control %s: %w", op.ControlID, err)
			}
		case domain.PatchDeleteControl:
			if op.ControlID == "" {
				log.Printf("Warning: delete-control requires controlId")
				continue
			}
			if err := deleteControlFromAST(f, op.ControlID); err != nil {
				return "", fmt.Errorf("failed to delete control %s: %w", op.ControlID, err)
			}
		case domain.PatchDeleteComposedOf:
			if op.ComposedOfID == "" {
				log.Printf("Warning: delete-composed-of requires composedOfId")
				continue
			}
			if err := deleteComposedOfFromAST(f, op.ComposedOfID); err != nil {
				return "", fmt.Errorf("failed to delete composed-of %s: %w", op.ComposedOfID, err)
			}
		case domain.PatchUpdateComposedOf:
			if op.ComposedOfID == "" {
				log.Printf("Warning: update-composed-of requires composedOfId")
				continue
			}
			if err := updateComposedOfInAST(f, op.ComposedOfID, op.ComposedOfDesc, op.ChildNodeIDs); err != nil {
				return "", fmt.Errorf("failed to update composed-of %s: %w", op.ComposedOfID, err)
			}
		case domain.PatchUpdateControl:
			if op.ControlID == "" {
				log.Printf("Warning: update-control requires controlId")
				continue
			}
			if err := updateControlInAST(f, op.ControlID, op.ControlDesc); err != nil {
				return "", fmt.Errorf("failed to update control %s: %w", op.ControlID, err)
			}
		}
	}

	result, err := formatFile(fset, f)
	if err != nil {
		return "", err
	}

	// If there's a pending relationship to add, insert it into the source
	if pendingRelationship.pending {
		result = insertRelationshipIntoSource(
			result,
			pendingRelationship.id,
			pendingRelationship.source,
			pendingRelationship.target,
			pendingRelationship.isInteracts,
		)
		pendingRelationship.pending = false
	}

	// If there's a pending flow to add, insert it into the source
	if pendingFlow.pending {
		result = insertFlowIntoSource(
			result,
			pendingFlow.id,
			pendingFlow.name,
			pendingFlow.desc,
			pendingFlow.steps,
		)
		pendingFlow.pending = false
	}

	// If there's a pending composed-of to add, insert it into the source
	if pendingComposedOf.pending {
		result = insertComposedOfIntoSource(
			result,
			pendingComposedOf.id,
			pendingComposedOf.containerID,
			pendingComposedOf.childNodeIDs,
		)
		pendingComposedOf.pending = false
	}

	// If there's a pending control to add, insert it into the source
	if pendingControl.pending {
		result = insertControlIntoSource(
			result,
			pendingControl.id,
			pendingControl.desc,
		)
		pendingControl.pending = false
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
	isInteracts        bool
	pending            bool
}

func addRelationshipToAST(
	f *ast.File,
	fset *token.FileSet,
	relationshipID, sourceNode, targetNode string,
	isInteracts bool,
) error {
	// Store the relationship for later source code insertion
	pendingRelationship.id = relationshipID
	pendingRelationship.source = sourceNode
	pendingRelationship.target = targetNode
	pendingRelationship.isInteracts = isInteracts
	pendingRelationship.pending = true

	if isInteracts {
		log.Printf("📝 Adding Interacts: %s from %s to %s", relationshipID, sourceNode, targetNode)
	} else {
		log.Printf("📝 Adding Connect: %s from %s to %s", relationshipID, sourceNode, targetNode)
	}
	return nil
}

// insertRelationshipIntoSource inserts a Connect() or Interacts() call into the Go source code.
// It looks for the wireComponents function and inserts before "return lc".
func insertRelationshipIntoSource(src, relationshipID, sourceNode, targetNode string, isInteracts bool) string {
	// Generate the relationship line to insert
	var relationshipLine string
	if isInteracts {
		relationshipLine = fmt.Sprintf(`
	// GUI-generated interaction: %s
	a.Interacts("%s", "Interaction from %s to %s", "%s", "%s")
`, relationshipID, relationshipID, sourceNode, targetNode, sourceNode, targetNode)
	} else {
		relationshipLine = fmt.Sprintf(`
	// GUI-generated connection: %s
	a.Connect("%s", "Connection from %s to %s", "%s", "%s")
`, relationshipID, relationshipID, sourceNode, targetNode, sourceNode, targetNode)
	}

	// Find "return lc" in wireComponents and insert before it
	pattern := "\treturn lc\n"
	insertPoint := strings.LastIndex(src, pattern)
	if insertPoint == -1 {
		pattern = "return lc"
		insertPoint = strings.LastIndex(src, pattern)
	}

	if insertPoint == -1 {
		log.Printf("Warning: Could not find insertion point in wireComponents")
		return src
	}

	return src[:insertPoint] + relationshipLine + src[insertPoint:]
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

		// Look for the assignment statement containing DefineNode with matching ID
		var insertIndex int = -1
		var varName string

		for i, stmt := range block.List {
			// Check for assignment: varName := a.DefineNode("nodeID", ...) or varName = ...
			// 1. Short Variable Declaration: varName := ...
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
			if insertIndex != -1 {
				break
			}
		}

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
		return fmt.Errorf("definition for node %s not found (must be explicit variable assignment)", nodeID)
	}
	return nil
}

// isDefineNodeCallWithID checks if expr is a DefineNode call for the specific nodeID
func isDefineNodeCallWithID(expr ast.Expr, nodeID string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	if !isDefineNode(call) {
		return false
	}
	if len(call.Args) < 1 {
		return false
	}
	// Check first arg (ID)
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	val := strings.Trim(lit.Value, "\"")
	return val == nodeID
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
						val := strings.Trim(lit.Value, "\"")
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

// addFlowToAST adds a new DefineFlow call to the source.
// It inserts the code textually into wireComponents for simplicity.
func addFlowToAST(f *ast.File, flowID, name, desc string, steps []string) error {
	pendingFlow.id = flowID
	pendingFlow.name = name
	pendingFlow.desc = desc
	pendingFlow.steps = steps
	pendingFlow.pending = true

	log.Printf("📝 Adding Flow: %s", flowID)
	return nil
}

var pendingFlow struct {
	id, name, desc string
	steps          []string
	pending        bool
}

// insertFlowIntoSource inserts a DefineFlow call into the Go source code.
// It inserts into wireComponents before "return lc".
func insertFlowIntoSource(src string, flowID, name, desc string, steps []string) string {
	stepsCode := ""
	for _, stepID := range steps {
		// Use quotes for ID as it's a string literal in the generated call
		stepsCode += fmt.Sprintf("\t\t\tdomain.StepSpec{ID: \"%s\"},\n", stepID)
	}

	flowCode := fmt.Sprintf(`
	// GUI-generated flow: %s
	a.DefineFlow("%s", "%s", "%s").
		Steps(
%s		)
`, flowID, flowID, name, desc, stepsCode)

	// Find "return lc" in wireComponents and insert before it
	pattern := "\treturn lc\n"
	insertPoint := strings.LastIndex(src, pattern)
	if insertPoint == -1 {
		pattern = "return lc"
		insertPoint = strings.LastIndex(src, pattern)
	}

	if insertPoint == -1 {
		log.Printf("Warning: Could not find insertion point in wireComponents")
		return src
	}

	return src[:insertPoint] + flowCode + src[insertPoint:]
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

// --- ComposedOf helpers ---

var pendingComposedOf struct {
	id           string
	containerID  string
	childNodeIDs []string
	pending      bool
}

// addComposedOfToAST adds a new ComposedOf relationship to the source.
// It uses textual insertion for simplicity.
func addComposedOfToAST(f *ast.File, id, containerID string, childNodeIDs []string) error {
	if id == "" {
		id = fmt.Sprintf("composed-%s", containerID)
	}
	pendingComposedOf.id = id
	pendingComposedOf.containerID = containerID
	pendingComposedOf.childNodeIDs = childNodeIDs
	pendingComposedOf.pending = true

	log.Printf("📝 Adding ComposedOf: %s (container: %s)", id, containerID)
	return nil
}

// insertComposedOfIntoSource inserts a ComposedOf call into the Go source code.
func insertComposedOfIntoSource(src, id, containerID string, childNodeIDs []string) string {
	// Build the node IDs slice: []string{"n1", "n2"}
	nodeLiterals := ""
	for i, nid := range childNodeIDs {
		if i > 0 {
			nodeLiterals += ", "
		}
		nodeLiterals += fmt.Sprintf("%q", nid)
	}

	composedCode := fmt.Sprintf(`
	// GUI-generated composed-of: %s
	a.ComposedOf("%s", "Container relationship", %s, []string{%s})
`, id, id, containerID, nodeLiterals)

	// Find "return lc" and insert before it
	pattern := "\treturn lc\n"
	insertPoint := strings.LastIndex(src, pattern)
	if insertPoint == -1 {
		pattern = "return lc"
		insertPoint = strings.LastIndex(src, pattern)
	}

	if insertPoint == -1 {
		log.Printf("Warning: Could not find insertion point for ComposedOf")
		return src
	}

	return src[:insertPoint] + composedCode + src[insertPoint:]
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

// --- Control helpers ---

var pendingControl struct {
	id      string
	desc    string
	pending bool
}

// addControlToAST adds a new AddControl call to the source.
func addControlToAST(f *ast.File, controlID, desc string) error {
	pendingControl.id = controlID
	pendingControl.desc = desc
	pendingControl.pending = true

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

// insertControlIntoSource inserts an AddControl call into the Go source code.
func insertControlIntoSource(src, controlID, desc string) string {
	controlCode := fmt.Sprintf(`
	// GUI-generated control: %s
	arch.AddControl("%s", "%s")
`, controlID, controlID, desc)

	// Find "return lc" and insert before it
	pattern := "\treturn lc\n"
	insertPoint := strings.LastIndex(src, pattern)
	if insertPoint == -1 {
		pattern = "return lc"
		insertPoint = strings.LastIndex(src, pattern)
	}

	if insertPoint == -1 {
		log.Printf("Warning: Could not find insertion point for Control")
		return src
	}

	return src[:insertPoint] + controlCode + src[insertPoint:]
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

// findVariableNameForNodeID finds the Go variable name that holds a node with the given unique-id
// by scanning DefineNode calls like: nc.OrderReplica = a.DefineNode("order-database-replica", ...)
func findVariableNameForNodeID(f *ast.File, nodeID string) string {
	var varName string

	ast.Inspect(f, func(n ast.Node) bool {
		if varName != "" {
			return false
		}

		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		if len(assign.Rhs) == 0 {
			return true
		}

		// Check if RHS is a DefineNode call
		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}

		if !isDefineNodeCallSimple(call) {
			return true
		}

		// Check if first argument is the nodeID
		if len(call.Args) > 0 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if strings.Trim(lit.Value, "\"") == nodeID {
					// Found it! Get the variable name from LHS
					if len(assign.Lhs) > 0 {
						if sel, ok := assign.Lhs[0].(*ast.SelectorExpr); ok {
							varName = sel.Sel.Name
						} else if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
							varName = ident.Name
						}
					}
				}
			}
		}
		return true
	})

	return varName
}

// isDefineNodeCallSimple checks if a call expression is a DefineNode call
func isDefineNodeCallSimple(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "DefineNode"
}

// deleteRelationshipsReferencingVariable removes all ConnectTo calls that reference the given Go variable name
func deleteRelationshipsReferencingVariable(f *ast.File, varName string) {
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		newList := make([]ast.Stmt, 0, len(fn.Body.List))
		for _, stmt := range fn.Body.List {
			// Check both ExprStmt and AssignStmt
			shouldDelete := false

			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				if call, ok := exprStmt.X.(*ast.CallExpr); ok {
					if callChainReferencesVariable(call, varName) {
						shouldDelete = true
					}
				}
			} else if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
				if len(assignStmt.Rhs) > 0 {
					if call, ok := assignStmt.Rhs[0].(*ast.CallExpr); ok {
						if callChainReferencesVariable(call, varName) {
							shouldDelete = true
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
}

// callChainReferencesVariable checks if a call expression chain references the given Go variable name
func callChainReferencesVariable(call *ast.CallExpr, varName string) bool {
	// Use a stack to traverse all call expressions in the chain
	var checkExpr func(expr ast.Expr) bool
	checkExpr = func(expr ast.Expr) bool {
		switch e := expr.(type) {
		case *ast.CallExpr:
			// Check arguments
			for _, arg := range e.Args {
				if checkExpr(arg) {
					return true
				}
			}
			// Check the function/receiver
			return checkExpr(e.Fun)

		case *ast.SelectorExpr:
			// Check if this is n.<varName> or nc.<varName>
			if e.Sel.Name == varName {
				return true
			}
			// Continue checking the receiver
			return checkExpr(e.X)

		default:
			return false
		}
	}

	return checkExpr(call)
}

// deleteComposedOfReferencingVariable removes or updates ComposedOf calls that reference the given variable
func deleteComposedOfReferencingVariable(f *ast.File, varName string) {
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
				// 1. Check if the container (parent) matches varName -> Delete entire statement
				if composedOfContainerMatches(call, varName) {
					continue // Delete statement
				}

				// 2. Check and filter children slice
				if removeComposedOfChild(call, varName) {
					// Modified in place, keep the statement
				}
			}

			newList = append(newList, stmt)
		}
		fn.Body.List = newList
	}
}

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

func composedOfContainerMatches(call *ast.CallExpr, varName string) bool {
	baseCall := getComposedOfBaseCall(call)
	if baseCall == nil || len(baseCall.Args) < 3 {
		return false
	}
	// Container is 3rd argument
	return variableReferenceMatches(baseCall.Args[2], varName)
}

// removeComposedOfChild removes the varName from the children slice. Returns true if modified.
func removeComposedOfChild(call *ast.CallExpr, varName string) bool {
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
		if variableReferenceMatches(elt, varName) {
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

// variableReferenceMatches checks if an expression references varName (e.g. n.OrderReplica or n.OrderReplica.UniqueID)
func variableReferenceMatches(expr ast.Expr, varName string) bool {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		// Check "UniqueID" pattern: n.OrderReplica.UniqueID
		if e.Sel.Name == "UniqueID" {
			// Check X
			return variableReferenceMatches(e.X, varName)
		}
		// Check direct access: n.OrderReplica
		if e.Sel.Name == varName {
			return true
		}
	case *ast.Ident:
		if e.Name == varName {
			return true
		}
	}
	return false
}
